package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims represents the JWT claims structure
type UserClaims struct {
	UserID   uint   `json:"user_id"`
	TenantID uint   `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	// JWT secret (same as in the server)
	secretKey := "default-development-secret-key"

	// Create claims for the default user/tenant
	claims := UserClaims{
		UserID:   1,
		TenantID: 1,
		Email:    "admin@unburdy.dev",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "unburdy-api",
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("JWT Token for testing:\n%s\n\n", tokenString)
	fmt.Printf("Use this token in the Authorization header as:\n")
	fmt.Printf("Authorization: Bearer %s\n\n", tokenString)
	fmt.Printf("Test with curl:\n")
	fmt.Printf("curl -H \"Authorization: Bearer %s\" http://localhost:8081/api/v1/clients\n", tokenString)
}
