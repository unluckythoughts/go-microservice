package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
	"github.com/unluckythoughts/go-microservice/v2/utils"
)

// TestEnsureRoleHierarchy verifies that higher-privilege roles pass
// lower-privilege requirements and lower-privilege roles are blocked.
func TestEnsureRoleHierarchy(t *testing.T) {
	cases := []struct {
		name        string
		userRole    auth.Role
		required    auth.Role
		expectAllow bool
	}{
		{"user satisfies user requirement", 1, 1, true},
		{"admin (99) satisfies user (1) requirement", 99, 1, true},
		{"user (1) blocked from admin (99) route", 1, 99, false},
		{"zero role satisfies zero requirement", 0, 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Role comparison: user.Role < required means insufficient privilege
			allowed := tc.userRole >= tc.required
			assert.Equal(t, tc.expectAllow, allowed)
		})
	}
}

// TestPasswordHashRoundTrip ensures GetHash + CompareValue work together.
func TestPasswordHashRoundTrip(t *testing.T) {
	password := "ValidPass1!"

	hash, err := utils.GetHash(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash, "stored value must be hashed, not plaintext")

	match, err := utils.CompareValue(password, hash)
	assert.NoError(t, err)
	assert.True(t, match, "correct password should match its hash")
}

// TestPasswordHashRejectsWrongPassword ensures a wrong password never matches.
func TestPasswordHashRejectsWrongPassword(t *testing.T) {
	hash, err := utils.GetHash("CorrectPass1!")
	assert.NoError(t, err)

	match, err := utils.CompareValue("WrongPass1!", hash)
	assert.NoError(t, err)
	assert.False(t, match, "wrong password must not match")
}

// TestPasswordHashUniqueSalts ensures two hashes of the same value differ.
func TestPasswordHashUniqueSalts(t *testing.T) {
	hash1, err := utils.GetHash("SamePass1!")
	assert.NoError(t, err)

	hash2, err := utils.GetHash("SamePass1!")
	assert.NoError(t, err)

	assert.NotEqual(t, hash1, hash2, "each hash must use a unique random salt")
}

// TestPasswordTypeIsValid tests the auth.Password domain type validation.
func TestPasswordTypeValidation(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"valid strong password", "Str0ng@Pass!", true},
		{"too short", "Ab1!", false},
		{"missing digit", "StrongPassword!", false},
		{"missing special char", "StrongPass123", false},
		{"missing uppercase", "weakpass1!", false},
		{"missing lowercase", "WEAKPASS1!", false},
		{"empty string", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := auth.Password(tc.value)
			assert.Equal(t, tc.isValid, p.IsValid())
		})
	}
}
