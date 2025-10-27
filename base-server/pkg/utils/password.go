package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PasswordRequirements defines the password security requirements
type PasswordRequirements struct {
	MinLength int  `json:"minLength"`
	Capital   bool `json:"capital"`
	Numbers   bool `json:"numbers"`
	Special   bool `json:"special"`
}

// GetPasswordRequirements returns the current password requirements configuration
func GetPasswordRequirements() PasswordRequirements {
	// Default requirements
	requirements := PasswordRequirements{
		MinLength: 8,
		Capital:   true,
		Numbers:   true,
		Special:   true,
	}

	// Allow override from environment variables
	minLenStr := GetEnv("PASSWORD_MIN_LENGTH", "8")
	if parsed, err := strconv.Atoi(minLenStr); err == nil {
		requirements.MinLength = parsed
	}

	capitalStr := GetEnv("PASSWORD_REQUIRE_CAPITAL", "true")
	requirements.Capital = strings.ToLower(capitalStr) == "true"

	numbersStr := GetEnv("PASSWORD_REQUIRE_NUMBERS", "true")
	requirements.Numbers = strings.ToLower(numbersStr) == "true"

	specialStr := GetEnv("PASSWORD_REQUIRE_SPECIAL", "true")
	requirements.Special = strings.ToLower(specialStr) == "true"

	return requirements
}

// ValidatePassword validates a password against the current requirements
func ValidatePassword(password string) error {
	requirements := GetPasswordRequirements()

	// Check minimum length
	if len(password) < requirements.MinLength {
		return fmt.Errorf("password must be at least %d characters long", requirements.MinLength)
	}

	// Check for capital letters
	if requirements.Capital {
		hasCapital := regexp.MustCompile(`[A-Z]`).MatchString(password)
		if !hasCapital {
			return fmt.Errorf("password must contain at least one uppercase letter")
		}
	}

	// Check for numbers
	if requirements.Numbers {
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
		if !hasNumber {
			return fmt.Errorf("password must contain at least one number")
		}
	}

	// Check for special characters
	if requirements.Special {
		hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)
		if !hasSpecial {
			return fmt.Errorf("password must contain at least one special character")
		}
	}

	return nil
}
