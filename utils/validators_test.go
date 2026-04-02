package utils

import (
	"strings"
	"testing"
)

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid password", "SecurePass1!", true},
		{"too short (< 10 chars)", "Abc1!", false},
		{"no uppercase letter", "securepass1!", false},
		{"no digit", "SecurePassword!", false},
		{"no special character", "SecurePass1234", false},
		{"no lowercase letter", "SECUREPASS1!", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPassword(tt.password)
			if got != tt.valid {
				t.Errorf("IsValidPassword(%q) = %v, want %v", tt.password, got, tt.valid)
			}
		})
	}
}

func TestIsValidPasswordTooLong(t *testing.T) {
	// 65 characters - one over the 64 char limit
	long := "SecurePass1!" + strings.Repeat("a", 53)
	if IsValidPassword(long) {
		t.Error("IsValidPassword should reject passwords longer than 64 characters")
	}
}

func TestIsMobile(t *testing.T) {
	tests := []struct {
		name   string
		mobile string
		valid  bool
	}{
		{"valid 10 digit", "1234567890", true},
		{"valid with dashes", "123-456-7890", true},
		{"valid with plus prefix", "+11234567890", true},
		{"valid 15 digit", "123456789012345", true},
		{"too short (< 10 digits)", "12345", false},
		{"non-numeric characters", "abcdefghij", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMobile(tt.mobile)
			if got != tt.valid {
				t.Errorf("IsMobile(%q) = %v, want %v", tt.mobile, got, tt.valid)
			}
		})
	}
}
