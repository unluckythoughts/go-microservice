package auth

import (
	"fmt"
	"regexp"
)

// Mobile is the mobile number of the user
type Mobile string

func (m *Mobile) String() string {
	mobile := string(*m)
	re := regexp.MustCompile(`[\s\+\-\(\)]`)
	mobile = re.ReplaceAllString(mobile, "")
	return mobile
}

func (m *Mobile) IsValid() bool {
	mobile := m.String()

	mobileRegex := regexp.MustCompile(`^[0-9]{10,15}$`)
	return mobileRegex.MatchString(mobile)
}

func (m *Mobile) Set(value string) error {
	if value == "" {
		return fmt.Errorf("mobile number cannot be empty")
	}
	*m = Mobile(value)
	*m = Mobile(m.String()) // Normalize the mobile number
	if !m.IsValid() {
		return fmt.Errorf("invalid mobile number format")
	}
	return nil
}

func (m *Mobile) GetCountryCode() (string, bool) {
	if m.IsValid() {
		mobileStr := string(*m)
		return mobileStr[:len(mobileStr)-10], true
	}
	return "", false
}

func (m *Mobile) GetNumber() (string, bool) {
	if m.IsValid() {
		mobileStr := string(*m)
		if len(mobileStr) >= 10 {
			return mobileStr[len(mobileStr)-10:], true
		}
	}
	return "", false
}

// UserRole represents the role of a user in the application
type UserRole uint

func (r *UserRole) Value() string {
	return fmt.Sprintf("%d", r)
}
