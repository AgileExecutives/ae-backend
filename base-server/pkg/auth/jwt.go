package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID      uint   `json:"user_id"`
	TenantID    uint   `json:"tenant_id"`
	Role        string `json:"role"`
	TokenType   string `json:"token_type,omitempty"`  // "auth", "onboarding", "verification"
	Permissions string `json:"permissions,omitempty"` // For onboarding: "limited"
	jwt.RegisteredClaims
}

var jwtSecret = []byte("your-secret-key") // TODO: Move to config

// GenerateJWT generates a JWT token for the user
func GenerateJWT(userID, tenantID uint, role string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		TenantID:  tenantID,
		Role:      role,
		TokenType: "auth",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("%d_%d", userID, time.Now().Unix()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateOnboardingToken generates a limited token for onboarding before email verification
func GenerateOnboardingToken(userID, tenantID uint, role string) (string, error) {
	claims := JWTClaims{
		UserID:      userID,
		TenantID:    tenantID,
		Role:        role,
		TokenType:   "onboarding",
		Permissions: "limited", // Can only access onboarding endpoints
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("onboarding_%d_%d", userID, time.Now().Unix()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)), // 3 days for onboarding
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ParseTokenClaims parses a token and returns the token ID and expiration
func ParseTokenClaims(tokenString string) (string, time.Time, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return "", time.Time{}, err
	}

	return claims.ID, claims.ExpiresAt.Time, nil
}

// SetJWTSecret sets the JWT secret (for configuration)
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// ResetTokenClaims represents the JWT claims for password reset tokens
type ResetTokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateResetToken generates a JWT token for password reset
func GenerateResetToken(email string, expiryDuration time.Duration) (string, error) {
	claims := ResetTokenClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("reset_%s_%d", email, time.Now().Unix()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateResetToken validates a password reset JWT token and returns the email
func ValidateResetToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ResetTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*ResetTokenClaims); ok && token.Valid {
		return claims.Email, nil
	}

	return "", fmt.Errorf("invalid reset token")
}

// GenerateVerificationToken generates a token for email verification
func GenerateVerificationToken(email string, userID uint) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		TokenType: "verification",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			ID:        fmt.Sprintf("verify_%d_%d", userID, time.Now().Unix()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours to verify
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateVerificationToken validates an email verification token and returns userID and email
func ValidateVerificationToken(tokenString string) (uint, string, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return 0, "", err
	}

	if claims.TokenType != "verification" {
		return 0, "", fmt.Errorf("invalid token type")
	}

	return claims.UserID, claims.Subject, nil
}

// GenerateUserSignupToken generates a token for user signup invitation to a specific tenant
func GenerateUserSignupToken(tenantID uint, inviterEmail string) (string, error) {
	claims := JWTClaims{
		TenantID:  tenantID,
		TokenType: "user-signup",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   inviterEmail, // Who sent the invitation
			ID:        fmt.Sprintf("signup_%d_%d", tenantID, time.Now().Unix()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days to use signup link
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateUserSignupToken validates a user signup token and returns tenant ID
func ValidateUserSignupToken(tokenString string) (uint, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return 0, err
	}

	if claims.TokenType != "user-signup" {
		return 0, fmt.Errorf("invalid token type")
	}

	return claims.TenantID, nil
}
