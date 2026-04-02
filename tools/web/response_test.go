package web

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSensitiveKey(t *testing.T) {
	cases := []struct {
		key       string
		sensitive bool
	}{
		{"password", true},
		{"user_password", true},
		{"passwd", true},
		{"secret", true},
		{"api_key", true},
		{"apikey", true},
		{"token", true},
		{"authorization", true},
		{"credential", true},
		{"private", true},
		{"jwt", true},
		{"ssn", true},
		{"cvv", true},
		{"card_number", true},
		{"cardnumber", true},
		{"PASSWORD", true},
		{"email", false},
		{"name", false},
		{"role", false},
		{"id", false},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			assert.Equal(t, tc.sensitive, isSensitiveKey(tc.key))
		})
	}
}

func TestRedactSensitiveFields(t *testing.T) {
	data := map[string]interface{}{
		"email":    "user@example.com",
		"password": "secret123",
		"name":     "Alice",
		"nested": map[string]interface{}{
			"token": "abc123",
			"role":  "admin",
		},
	}
	redactSensitiveFields(data)

	assert.Equal(t, "user@example.com", data["email"])
	assert.Equal(t, "[REDACTED]", data["password"])
	assert.Equal(t, "Alice", data["name"])

	nested := data["nested"].(map[string]interface{})
	assert.Equal(t, "[REDACTED]", nested["token"])
	assert.Equal(t, "admin", nested["role"])
}

func TestSanitizeBodyJSON(t *testing.T) {
	body := `{"email":"user@example.com","password":"secret123"}`
	result := sanitizeBody(body)
	assert.Contains(t, result, "user@example.com")
	assert.Contains(t, result, "[REDACTED]")
	assert.NotContains(t, result, "secret123")
}

func TestSanitizeBodyInvalidJSON(t *testing.T) {
	body := "not valid json"
	result := sanitizeBody(body)
	assert.Equal(t, body, result)
}

func TestSanitizeBodyEmpty(t *testing.T) {
	result := sanitizeBody("")
	assert.Equal(t, "", result)
}

func TestNotImplemented(t *testing.T) {
	_, err := NotImplemented(nil)
	assert.Error(t, err)
	e, ok := err.(*httpError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusMethodNotAllowed, e.code)
}

func TestUnAuthorized(t *testing.T) {
	_, err := UnAuthorized(nil)
	assert.Error(t, err)
	e, ok := err.(*httpError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, e.code)
}
