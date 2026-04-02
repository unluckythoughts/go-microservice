package web

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestCreateJWT(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!!"
	claims := jwt.MapClaims{
		"sub": "42",
		"iss": "test-app",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token, err := CreateJWT(secret, claims)
	if err != nil {
		t.Fatalf("CreateJWT returned error: %v", err)
	}
	if token == "" {
		t.Fatal("CreateJWT returned empty token")
	}
}

func TestParseJWTValid(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!!"
	claims := jwt.MapClaims{
		"sub": "42",
		"iss": "test-app",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token, err := CreateJWT(secret, claims)
	if err != nil {
		t.Fatalf("CreateJWT error: %v", err)
	}
	parsed, err := ParseJWT(secret, token)
	if err != nil {
		t.Fatalf("ParseJWT error: %v", err)
	}
	if !parsed.Valid {
		t.Error("parsed token should be valid")
	}
	parsedClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to assert claims as MapClaims")
	}
	sub, err := parsedClaims.GetSubject()
	if err != nil {
		t.Fatalf("GetSubject error: %v", err)
	}
	if sub != "42" {
		t.Errorf("expected sub=42, got %q", sub)
	}
}

func TestParseJWTWrongSecret(t *testing.T) {
	token, err := CreateJWT("correct-secret", jwt.MapClaims{
		"sub": "42",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("CreateJWT error: %v", err)
	}
	_, err = ParseJWT("wrong-secret", token)
	if err == nil {
		t.Error("ParseJWT should return error when secret does not match")
	}
}

func TestParseJWTExpiredToken(t *testing.T) {
	token, err := CreateJWT("secret", jwt.MapClaims{
		"sub": "42",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("CreateJWT error: %v", err)
	}
	_, err = ParseJWT("secret", token)
	if err == nil {
		t.Error("ParseJWT should return error for an expired token")
	}
}

func TestParseJWTInvalidToken(t *testing.T) {
	_, err := ParseJWT("secret", "not.a.valid.jwt")
	if err == nil {
		t.Error("ParseJWT should return error for a malformed token string")
	}
}
