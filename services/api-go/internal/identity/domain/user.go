package domain

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrInvalidEmail  = errors.New("signup requires a valid Gmail address (name@gmail.com)")
	ErrWeakPassword  = errors.New("password must be at least 8 characters and include letters and numbers")
	ErrInvalidRole   = errors.New("invalid role")
	ErrInvalidStatus = errors.New("invalid status")
)

const GmailSuffix = "@gmail.com"

// ValidateGmail enforces the platform signup rule: a Gmail address only.
func ValidateGmail(email string) bool {
	e := strings.ToLower(strings.TrimSpace(email))
	if !strings.HasSuffix(e, GmailSuffix) {
		return false
	}
	local := strings.TrimSuffix(e, GmailSuffix)
	return len(local) >= 3 && !strings.ContainsAny(local, " @[]\"(),:;<>\\")
}

// ValidatePassword enforces the platform password policy.
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter, hasDigit := false, false
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsSpace(r):
			return false
		}
	}
	return hasLetter && hasDigit
}
