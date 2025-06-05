package web

import (
	"regexp"

	"github.com/asaskevich/govalidator"
)

func init() {
	govalidator.CustomTypeTagMap.Set("mobile", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		if i == nil {
			return false
		}

		mobile, ok := i.(string)
		if !ok {
			return false
		}

		return isMobile(mobile)
	}))
}

func isMobile(mobile string) bool {
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
