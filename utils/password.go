package utils

import (
	"errors"
	"regexp"
	"unicode"
)

// PasswordStrength represents password strength levels
type PasswordStrength int

const (
	PasswordWeak PasswordStrength = iota
	PasswordMedium
	PasswordStrong
	PasswordVeryStrong
)

// PasswordRequirements defines password complexity requirements
type PasswordRequirements struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
	MaxLength        int
}

// DefaultPasswordRequirements returns the default password requirements
func DefaultPasswordRequirements() *PasswordRequirements {
	return &PasswordRequirements{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSpecial:   true,
		MaxLength:        128,
	}
}

// ValidatePassword checks if a password meets the requirements
func ValidatePassword(password string, requirements *PasswordRequirements) error {
	if len(password) < requirements.MinLength {
		return errors.New("password must be at least " + string(rune(requirements.MinLength)) + " characters")
	}

	if len(password) > requirements.MaxLength {
		return errors.New("password must be less than " + string(rune(requirements.MaxLength)) + " characters")
	}

	if requirements.RequireUppercase && !hasUppercase(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	if requirements.RequireLowercase && !hasLowercase(password) {
		return errors.New("password must contain at least one lowercase letter")
	}

	if requirements.RequireNumber && !hasNumber(password) {
		return errors.New("password must contain at least one number")
	}

	if requirements.RequireSpecial && !hasSpecial(password) {
		return errors.New("password must contain at least one special character (!@#$%^&*(),.?\":{}|<>)")
	}

	return nil
}

// CalculatePasswordStrength calculates the strength of a password
func CalculatePasswordStrength(password string) PasswordStrength {
	score := 0

	// Length score
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}

	// Complexity score
	if hasUppercase(password) {
		score++
	}
	if hasLowercase(password) {
		score++
	}
	if hasNumber(password) {
		score++
	}
	if hasSpecial(password) {
		score++
	}

	// Determine strength
	switch {
	case score <= 3:
		return PasswordWeak
	case score <= 5:
		return PasswordMedium
	case score <= 7:
		return PasswordStrong
	default:
		return PasswordVeryStrong
	}
}

// Helper functions
func hasUppercase(s string) bool {
	for _, c := range s {
		if unicode.IsUpper(c) {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, c := range s {
		if unicode.IsLower(c) {
			return true
		}
	}
	return false
}

func hasNumber(s string) bool {
	for _, c := range s {
		if unicode.IsDigit(c) {
			return true
		}
	}
	return false
}

func hasSpecial(s string) bool {
	specialChars := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)
	return specialChars.MatchString(s)
}

// CheckCommonPasswords checks if password is in common passwords list
func CheckCommonPasswords(password string) bool {
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"monkey", "1234567", "letmein", "trustno1", "dragon",
		"baseball", "iloveyou", "master", "sunshine", "ashley",
		"bailey", "passw0rd", "shadow", "123123", "654321",
		"superman", "qazwsx", "michael", "football",
	}

	for _, common := range commonPasswords {
		if password == common {
			return true
		}
	}
	return false
}
