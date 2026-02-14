package utils

import (
	"log"
	"reflect"
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

		// For type aliases like `type Mobile string`, use reflection to extract underlying string
		v := reflect.ValueOf(i)

		if v.Kind() == reflect.String {
			password := v.String()
			result := IsValidPassword(password)
			return result
		}
		return IsMobile(mobile)
	}))

	// Register custom validator for password
	govalidator.CustomTypeTagMap.Set("password", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		if i == nil {
			log.Printf("[PASSWORD VALIDATOR] i is nil")
			return false
		}

		// Handle string type directly
		if password, ok := i.(string); ok {
			result := IsValidPassword(password)
			return result
		}

		// For type aliases like `type Password string`, use reflection to extract underlying string
		v := reflect.ValueOf(i)

		if v.Kind() == reflect.String {
			password := v.String()
			result := IsValidPassword(password)
			return result
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
