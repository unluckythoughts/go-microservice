package utils

import (
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Custom validators for govalidator
// This package registers custom validation tags that can be used with the govalidator library

func init() {
	// Register custom validator for email slices
	govalidator.CustomTypeTagMap.Set("emails", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		if i == nil {
			return false
		}

		emails, ok := i.([]string)
		if !ok {
			return false
		}

		if len(emails) < 1 {
			return false
		}

		for _, email := range emails {
			if !govalidator.IsEmail(strings.TrimSpace(email)) {
				return false
			}
		}

		return true
	}))

	// Register custom validator for mobile
	govalidator.CustomTypeTagMap.Set("mobile", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		if i == nil {
			return false
		}

		mobile, ok := i.(string)
		if !ok {
			// For type aliases like `type Mobile string`, use reflection to extract underlying string
			v := reflect.ValueOf(i)
			if v.Kind() != reflect.String {
				return false
			}
			mobile = v.String()
		}

		return IsMobile(mobile)
	}))

	// Register custom validator for indian mobile
	govalidator.CustomTypeTagMap.Set("indian_mobile", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		if i == nil {
			return false
		}

		mobile, ok := i.(string)
		if !ok {
			// For type aliases like `type Mobile string`, use reflection to extract underlying string
			v := reflect.ValueOf(i)
			if v.Kind() != reflect.String {
				return false
			}
			mobile = v.String()
		}

		return IsIndianMobile(mobile)
	}))

	// Register custom validator for indian mobiles
	govalidator.CustomTypeTagMap.Set("indian_mobiles", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		if i == nil {
			return false
		}

		mobiles, ok := i.([]string)
		if !ok {
			return false
		}

		for _, mobile := range mobiles {
			if !IsIndianMobile(mobile) {
				return false
			}
		}

		return true
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

func ValidateStruct(s interface{}) error {
	_, err := govalidator.ValidateStruct(s)
	return err
}

func IsEmail(email string) bool {
	return govalidator.IsEmail(strings.TrimSpace(email))
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

func IsIndianMobile(mobile string) bool {
	if mobile == "" {
		return false
	}

	// Remove any valid non-numeric characters from the mobile number
	extrasPattern := regexp.MustCompile(`[.() \-+]`)
	mobile = extrasPattern.ReplaceAllString(mobile, "")

	if !govalidator.IsNumeric(mobile) {
		return false
	}

	if len(mobile) != 10 && len(mobile) != 12 {
		return false
	}

	if len(mobile) == 12 && !strings.HasPrefix(mobile, "91") {
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
