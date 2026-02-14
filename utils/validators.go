package utils

import (
	"fmt"
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

		return IsMobile(mobile)
	}))

	// Register custom validator for password
	govalidator.CustomTypeTagMap.Set("password", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		if i == nil {
			log.Printf("[PASSWORD VALIDATOR] i is nil")
			return false
		}

		// Log what type we received
		log.Printf("[PASSWORD VALIDATOR] Received type: %T, value: %v", i, i)

		// Handle string type directly
		if password, ok := i.(string); ok {
			log.Printf("[PASSWORD VALIDATOR] Matched as string: %s", password)
			result := IsValidPassword(password)
			log.Printf("[PASSWORD VALIDATOR] String validation result: %v", result)
			return result
		}

		// Handle types with String() method
		if stringer, ok := i.(interface{ String() string }); ok {
			password := stringer.String()
			log.Printf("[PASSWORD VALIDATOR] Matched as Stringer, extracted: %s", password)
			result := IsValidPassword(password)
			log.Printf("[PASSWORD VALIDATOR] Stringer validation result: %v", result)
			return result
		}

		// Handle types with IsValid() method
		if validator, ok := i.(interface{ IsValid() bool }); ok {
			log.Printf("[PASSWORD VALIDATOR] Matched as validator with IsValid()")
			result := validator.IsValid()
			log.Printf("[PASSWORD VALIDATOR] IsValid() result: %v", result)
			return result
		}

		// For type aliases like `type Password string`, use reflection to extract underlying string
		v := reflect.ValueOf(i)
		log.Printf("[PASSWORD VALIDATOR] Reflection - Kind: %v", v.Kind())

		if v.Kind() == reflect.String {
			password := v.String()
			log.Printf("[PASSWORD VALIDATOR] Extracted via reflection: %s", password)
			result := IsValidPassword(password)
			log.Printf("[PASSWORD VALIDATOR] Reflection validation result: %v", result)
			return result
		}

		// Fallback: try to convert to string
		str := fmt.Sprintf("%v", i)
		log.Printf("[PASSWORD VALIDATOR] Fallback string conversion: %s", str)
		if str != "" && str != "<nil>" {
			result := IsValidPassword(str)
			log.Printf("[PASSWORD VALIDATOR] Fallback validation result: %v", result)
			return result
		}

		log.Printf("[PASSWORD VALIDATOR] No matching handler, returning false")
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
