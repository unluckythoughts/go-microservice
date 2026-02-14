package utils

import (
	"regexp"

	"github.com/asaskevich/govalidator"
)

// Custom validators for govalidator
// This package registers custom validation tags that can be used with the govalidator library

func init() {
	// Register custom validator for mobile
	govalidator.CustomTypeTagMap.Set("mobile", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		if i == nil {
			return false
		}

		mobile, ok := i.(string)
		if !ok {
			return false
		}

		return IsMobile(mobile)
	}))

	// Register custom validator for password
	govalidator.CustomTypeTagMap.Set("password", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		if i == nil {
			return false
		}

		// Handle string type
		if password, ok := i.(string); ok {
			return IsValidPassword(password)
		}

		// Handle types with String() method
		if stringer, ok := i.(interface{ String() string }); ok {
			return IsValidPassword(stringer.String())
		}

		// Handle types with IsValid() method
		if validator, ok := i.(interface{ IsValid() bool }); ok {
			return validator.IsValid()
		}

		return false
	}))
}

func IsMobile(mobile string) bool {
	if mobile == "" {
		return false
	}

	// Remove any valid non-numeric characters from the mobile number
	extrasPattern := regexp.MustCompile(`[.() \-+]`)
	mobile = extrasPattern.ReplaceAllString(mobile, "")

	if !govalidator.IsNumeric(mobile) {
		return false
	}

	if len(mobile) < 10 || len(mobile) > 15 {
		return false
	}

	return true
}

func IsValidPassword(password string) bool {
	if len(password) < 10 || len(password) > 64 {
		return false
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// Check for at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[\W_]`).MatchString(password)

	return hasLower && hasUpper && hasDigit && hasSpecial
}
