package utils

import (
	"strings"
	"testing"
)

func TestGetHash(t *testing.T) {
	hash, err := GetHash("testvalue")
	if err != nil {
		t.Fatalf("GetHash returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("GetHash returned empty hash")
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("expected hash to start with $argon2id$, got %q", hash)
	}
}

func TestGetHashProducesUniqueHashes(t *testing.T) {
	hash1, err := GetHash("same-value")
	if err != nil {
		t.Fatalf("first GetHash error: %v", err)
	}
	hash2, err := GetHash("same-value")
	if err != nil {
		t.Fatalf("second GetHash error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("GetHash should produce different hashes due to random salt")
	}
}

func TestCompareValueMatch(t *testing.T) {
	value := "my-secret-value"
	hash, err := GetHash(value)
	if err != nil {
		t.Fatalf("GetHash error: %v", err)
	}
	match, err := CompareValue(value, hash)
	if err != nil {
		t.Fatalf("CompareValue error: %v", err)
	}
	if !match {
		t.Error("CompareValue should match for the same value")
	}
}

func TestCompareValueMismatch(t *testing.T) {
	hash, err := GetHash("correct-value")
	if err != nil {
		t.Fatalf("GetHash error: %v", err)
	}
	match, err := CompareValue("wrong-value", hash)
	if err != nil {
		t.Fatalf("CompareValue error: %v", err)
	}
	if match {
		t.Error("CompareValue should not match for different values")
	}
}

func TestCompareValueInvalidHash(t *testing.T) {
	_, err := CompareValue("value", "invalid-hash-format")
	if err == nil {
		t.Error("CompareValue should return error for invalid hash format")
	}
}

func TestGenerateRandomString(t *testing.T) {
	s, err := GenerateRandomString(16)
	if err != nil {
		t.Fatalf("GenerateRandomString error: %v", err)
	}
	if len(s) != 16 {
		t.Errorf("expected length 16, got %d", len(s))
	}
}

func TestGenerateRandomStringUniqueness(t *testing.T) {
	s1, err := GenerateRandomString(32)
	if err != nil {
		t.Fatalf("first GenerateRandomString error: %v", err)
	}
	s2, err := GenerateRandomString(32)
	if err != nil {
		t.Fatalf("second GenerateRandomString error: %v", err)
	}
	if s1 == s2 {
		t.Error("GenerateRandomString should produce unique strings")
	}
}
