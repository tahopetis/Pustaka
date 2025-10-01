package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/pkg/errors"
)

type PasswordService struct {
	params *argon2id.Params
}

func NewPasswordService() *PasswordService {
	return &PasswordService{
		params: argon2id.DefaultParams,
	}
}

// HashPassword hashes a password using Argon2id
func (p *PasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hash, err := argon2id.CreateHash(password, p.params)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return hash, nil
}

// VerifyPassword verifies a password against its hash
func (p *PasswordService) VerifyPassword(password, hash string) (bool, error) {
	if password == "" || hash == "" {
		return false, errors.New("password and hash cannot be empty")
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("failed to verify password: %w", err)
	}

	return match, nil
}

// ValidatePassword checks if a password meets the requirements
func (p *PasswordService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.New("password must be no more than 128 characters long")
	}

	// Check for at least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password must contain at least one digit")
	}

	// Check for at least one special character
	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		return errors.New("password must contain at least one special character")
	}

	// Check for common weak patterns
	if p.isCommonPassword(password) {
		return errors.New("password is too common, please choose a stronger one")
	}

	return nil
}

// isCommonPassword checks against a list of common passwords
func (p *PasswordService) isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "password123", "admin", "qwerty",
		"letmein", "welcome", "monkey", "dragon", "master",
		"sunshine", "princess", "football", "baseball", "shadow",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}

	return false
}

// GenerateSecurePassword generates a random secure password
func (p *PasswordService) GenerateSecurePassword(length int) (string, error) {
	if length < 8 {
		length = 12
	}

	if length > 128 {
		length = 128
	}

	// Define character sets
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
		special   = "!@#$%^&*(),.?\":{}|<>"
	)

	allChars := lowercase + uppercase + digits + special

	// Generate random password
	password := make([]byte, length)
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	for i := 0; i < length; i++ {
		password[i] = allChars[randomBytes[i]%byte(len(allChars))]
	}

	// Ensure password contains at least one character from each set
	hasLowercase := false
	hasUppercase := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case strings.ContainsRune(lowercase, rune(char)):
			hasLowercase = true
		case strings.ContainsRune(uppercase, rune(char)):
			hasUppercase = true
		case strings.ContainsRune(digits, rune(char)):
			hasDigit = true
		case strings.ContainsRune(special, rune(char)):
			hasSpecial = true
		}
	}

	// If password doesn't meet requirements, generate a new one
	if !hasLowercase || !hasUppercase || !hasDigit || !hasSpecial {
		return p.GenerateSecurePassword(length)
	}

	return string(password), nil
}

// GenerateRandomToken generates a cryptographically secure random token
func (p *PasswordService) GenerateRandomToken(length int) (string, error) {
	if length < 16 {
		length = 32
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateResetToken generates a password reset token
func (p *PasswordService) GenerateResetToken() (string, error) {
	return p.GenerateRandomToken(32)
}

// ValidateResetToken validates a reset token format
func (p *PasswordService) ValidateResetToken(token string) bool {
	if token == "" {
		return false
	}

	// Base64 URL encoded tokens should be valid
	_, err := base64.URLEncoding.DecodeString(token)
	return err == nil
}