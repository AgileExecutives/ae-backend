package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ae-saas-basic/ae-saas-basic/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// getRateLimitConfig reads rate limit configuration from environment variables
func getRateLimitConfig() (enabled bool, requests int64, duration time.Duration) {
	// Check if rate limiting is enabled (default: true)
	enabled = os.Getenv("RATE_LIMIT_ENABLED") != "false"

	// Get number of requests (default: 100)
	requests = 100
	if reqStr := os.Getenv("RATE_LIMIT_REQUESTS"); reqStr != "" {
		if parsed, err := strconv.ParseInt(reqStr, 10, 64); err == nil {
			requests = parsed
		}
	}

	// Get duration (default: 1h)
	duration = 1 * time.Hour
	if durStr := os.Getenv("RATE_LIMIT_DURATION"); durStr != "" {
		if parsed, err := time.ParseDuration(durStr); err == nil {
			duration = parsed
		}
	}

	return
}

// getEffectiveRate returns the rate limit to use based on environment configuration
func getEffectiveRate(defaultRate limiter.Rate) limiter.Rate {
	enabled, requests, duration := getRateLimitConfig()

	if !enabled {
		// When rate limiting is disabled, return a very high limit (effectively unlimited)
		return limiter.Rate{
			Period: 1 * time.Second,
			Limit:  1000000, // 1 million requests per second
		}
	}

	// Use environment configuration if available
	if requests > 0 && duration > 0 {
		return limiter.Rate{
			Period: duration,
			Limit:  requests,
		}
	}

	// Fallback to default rate
	return defaultRate
}

// NewRateLimiter creates a new rate limiter middleware with in-memory store
func NewRateLimiter(rate limiter.Rate) gin.HandlerFunc {
	actualRate := getEffectiveRate(rate)
	store := memory.NewStore()
	instance := limiter.New(store, actualRate, limiter.WithTrustForwardHeader(true))
	return mgin.NewMiddleware(instance)
}

// NewRateLimiterWithKey creates a rate limiter with custom key function
func NewRateLimiterWithKey(rate limiter.Rate, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	actualRate := getEffectiveRate(rate)
	store := memory.NewStore()
	instance := limiter.New(store, actualRate, limiter.WithTrustForwardHeader(true))

	return func(c *gin.Context) {
		key := keyFunc(c)
		context, err := instance.Get(c, key)

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc(
				"Rate limiter error",
				err.Error(),
			))
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		if context.Reached {
			retryAfter := time.Unix(context.Reset, 0).Sub(time.Now()).Seconds()
			if retryAfter < 0 {
				retryAfter = 0
			}
			c.Header("Retry-After", fmt.Sprintf("%d", int(retryAfter)))

			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Message: "Zu viele Anfragen",
				Error:   "Sie haben zu viele Anfragen gesendet. Bitte versuchen Sie es später erneut.",
				Data: map[string]interface{}{
					"limit":       context.Limit,
					"remaining":   context.Remaining,
					"reset_at":    time.Unix(context.Reset, 0).Format(time.RFC3339),
					"retry_after": int(retryAfter),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetIPKey returns IP-based key for rate limiting
func GetIPKey(c *gin.Context) string {
	// Check for X-Forwarded-For header (if behind proxy)
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return ip
	}
	// Check for X-Real-IP header
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	// Fallback to RemoteAddr
	return c.ClientIP()
}

// GetUserKey returns user-based key for authenticated endpoints
func GetUserKey(c *gin.Context) string {
	// Try to get user ID from context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}
	// Fallback to IP
	return GetIPKey(c)
}

// Predefined rate limiters for common use cases
var (
	// LoginRateLimiter - 5 attempts per 15 minutes
	LoginRateLimiter = limiter.Rate{
		Period: 15 * time.Minute,
		Limit:  5,
	}

	// RegisterRateLimiter - 3 attempts per hour
	RegisterRateLimiter = limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  3,
	}

	// PasswordResetRateLimiter - 3 attempts per hour
	PasswordResetRateLimiter = limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  3,
	}

	// EmailVerificationRateLimiter - 5 attempts per hour
	EmailVerificationRateLimiter = limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  5,
	}

	// GlobalRateLimiter - 1000 requests per hour per IP
	GlobalRateLimiter = limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  1000,
	}

	// APIReadRateLimiter - 60 requests per minute
	APIReadRateLimiter = limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  60,
	}

	// APIWriteRateLimiter - 20 requests per minute
	APIWriteRateLimiter = limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  20,
	}
)
