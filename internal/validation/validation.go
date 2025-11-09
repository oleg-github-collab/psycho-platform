package validation

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidUsername = errors.New("username must be 3-50 characters and contain only letters, numbers, and underscores")
	ErrInvalidPassword = errors.New("password must be at least 6 characters")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrContentTooLong  = errors.New("content exceeds maximum length")
	ErrEmptyContent    = errors.New("content cannot be empty")
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 6 {
		return ErrInvalidPassword
	}
	return nil
}

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

func ValidateContent(content string, maxLength int) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return ErrEmptyContent
	}
	if len(content) > maxLength {
		return ErrContentTooLong
	}
	return nil
}

func SanitizeString(s string) string {
	s = strings.TrimSpace(s)
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	return s
}
