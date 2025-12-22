package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService provides generic token generation and validation
type TokenService struct {
	secretKey []byte
}

// NewTokenService creates a new token service
func NewTokenService(secretKey string) *TokenService {
	if secretKey == "" {
		secretKey = "your-super-secret-jwt-key-change-in-production"
	}
	return &TokenService{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken generates a JWT token with custom claims
// The claims parameter must implement jwt.Claims interface
func (ts *TokenService) GenerateToken(claimsInterface interface{}) (string, error) {
	claims, ok := claimsInterface.(jwt.Claims)
	if !ok {
		return "", fmt.Errorf("claims must implement jwt.Claims interface")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ts.secretKey)
}

// ValidateToken validates a JWT token and populates the provided claims structure
// The claims parameter should be a pointer to a struct implementing jwt.Claims
func (ts *TokenService) ValidateToken(tokenString string, claimsInterface interface{}) error {
	claims, ok := claimsInterface.(jwt.Claims)
	if !ok {
		return fmt.Errorf("claims must implement jwt.Claims interface")
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ts.secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

// ParseTokenID extracts the token ID (jti) from a token without full validation
// Useful for blacklist checking before expensive validation
func (ts *TokenService) ParseTokenID(tokenString string) (string, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &jwt.RegisteredClaims{})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		return claims.ID, nil
	}

	return "", fmt.Errorf("unable to extract token ID")
}

// GetTokenExpiration extracts expiration time from a token without full validation
func (ts *TokenService) GetTokenExpiration(tokenString string) (time.Time, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &jwt.RegisteredClaims{})
	if err != nil {
		return time.Time{}, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		if claims.ExpiresAt != nil {
			return claims.ExpiresAt.Time, nil
		}
		return time.Time{}, nil
	}

	return time.Time{}, fmt.Errorf("unable to extract expiration")
}

// SetSecret updates the JWT secret (for configuration)
func (ts *TokenService) SetSecret(secret string) {
	ts.secretKey = []byte(secret)
}

// GetSharedTokenService returns the global token service instance
// This uses the same secret as the existing JWT functions for compatibility
func GetSharedTokenService() *TokenService {
	return NewTokenService(string(jwtSecret))
}
