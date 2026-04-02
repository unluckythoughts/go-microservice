package auth

import (
	"testing"
)

func TestPasswordIsValid(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid password", "SecurePass1!", true},
		{"too short", "Abc1!", false},
		{"no uppercase", "securepass1!", false},
		{"no digit", "SecurePassword!", false},
		{"no special char", "SecurePass1234", false},
		{"no lowercase", "SECUREPASS1!", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Password(tt.password)
			if got := p.IsValid(); got != tt.valid {
				t.Errorf("Password(%q).IsValid() = %v, want %v", tt.password, got, tt.valid)
			}
		})
	}
}

func TestPasswordSetValid(t *testing.T) {
	var p Password
	if err := p.Set("SecurePass1!"); err != nil {
		t.Errorf("unexpected error for valid password: %v", err)
	}
	if string(p) != "SecurePass1!" {
		t.Errorf("expected password to be set, got %q", string(p))
	}
}

func TestPasswordSetInvalid(t *testing.T) {
	var p Password
	if err := p.Set("weak"); err == nil {
		t.Error("expected error for weak password")
	}
}

func TestPasswordSetEmpty(t *testing.T) {
	var p Password
	if err := p.Set(""); err == nil {
		t.Error("expected error when setting empty password")
	}
}

func TestMobileIsValid(t *testing.T) {
	tests := []struct {
		name   string
		mobile string
		valid  bool
	}{
		{"valid 10 digit", "1234567890", true},
		{"valid 15 digit", "123456789012345", true},
		{"too short", "12345", false},
		{"non-numeric", "abcdefghij", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Mobile(tt.mobile)
			if got := m.IsValid(); got != tt.valid {
				t.Errorf("Mobile(%q).IsValid() = %v, want %v", tt.mobile, got, tt.valid)
			}
		})
	}
}

func TestMobileSetValid(t *testing.T) {
	var m Mobile
	if err := m.Set("1234567890"); err != nil {
		t.Errorf("unexpected error for valid mobile: %v", err)
	}
	if m.String() != "1234567890" {
		t.Errorf("expected 1234567890, got %q", m.String())
	}
}

func TestMobileSetInvalid(t *testing.T) {
	var m Mobile
	if err := m.Set("123"); err == nil {
		t.Error("expected error for invalid mobile number")
	}
}

func TestMobileSetEmpty(t *testing.T) {
	var m Mobile
	if err := m.Set(""); err == nil {
		t.Error("expected error when setting empty mobile number")
	}
}

func TestMobileGetNumber(t *testing.T) {
	m := Mobile("11234567890")
	number, ok := m.GetNumber()
	if !ok {
		t.Fatal("GetNumber should succeed for valid mobile")
	}
	if number != "1234567890" {
		t.Errorf("expected 1234567890, got %q", number)
	}
}

func TestRoleHierarchy(t *testing.T) {
	user := Role(0)
	admin := Role(99)
	if user >= admin {
		t.Error("user role (0) should be less than admin role (99)")
	}
}

func TestRoleValue(t *testing.T) {
	r := Role(42)
	if r.Value() == "" {
		t.Error("Role.Value() should not return empty string")
	}
}
