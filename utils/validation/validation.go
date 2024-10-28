package validation

import (
	"errors"
	"regexp"
)

// validateEmail checks if the provided email is in a valid format
func ValidateEmail(email string) error {
	// Regex pattern for validating email format
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		return errors.New("error validating email format")
	}
	if !matched {
		return errors.New("invalid email format")
	}
	return nil
}
