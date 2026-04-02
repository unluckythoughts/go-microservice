package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unluckythoughts/go-microservice/v2/tests"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

func TestCreateUserAndVerifyPasswordByEmail(t *testing.T) {
	ts := tests.NewTestService(t)

	user := &auth.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "ValidPass1!",
		Role:     1,
	}
	err := ts.Auth.CreateUser(user)
	assert.NoError(t, err, "create user should not fail")
	assert.NotZero(t, user.ID, "user ID should be set after creation")

	foundUser, ok, err := ts.Auth.VerifyUserPasswordByEmail("test@example.com", "ValidPass1!")
	assert.NoError(t, err)
	assert.True(t, ok, "correct password should verify successfully")
	assert.Equal(t, "test@example.com", foundUser.Email)
}

func TestVerifyPasswordByEmailWrongPassword(t *testing.T) {
	ts := tests.NewTestService(t)

	user := &auth.User{
		Name:     "Test User",
		Email:    "user2@example.com",
		Password: "ValidPass1!",
		Role:     1,
	}
	assert.NoError(t, ts.Auth.CreateUser(user))

	_, ok, err := ts.Auth.VerifyUserPasswordByEmail("user2@example.com", "WrongPass1!")
	assert.NoError(t, err)
	assert.False(t, ok, "wrong password must not verify")
}

func TestVerifyPasswordByEmailNotFound(t *testing.T) {
	ts := tests.NewTestService(t)

	_, _, err := ts.Auth.VerifyUserPasswordByEmail("nobody@example.com", "SomePass1!")
	assert.Error(t, err, "looking up a non-existent user should return an error")
}

func TestGetUserByID(t *testing.T) {
	ts := tests.NewTestService(t)

	user := &auth.User{
		Name:     "Lookup User",
		Email:    "lookup@example.com",
		Password: "ValidPass1!",
		Role:     1,
	}
	assert.NoError(t, ts.Auth.CreateUser(user))
	assert.NotZero(t, user.ID)

	found, err := ts.Auth.GetUserByID(user.ID)
	assert.NoError(t, err)
	if found != nil {
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, "lookup@example.com", found.Email)
		assert.Empty(t, string(found.Password), "password must be cleared from returned user")
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	ts := tests.NewTestService(t)

	_, err := ts.Auth.GetUserByID(99999)
	assert.Error(t, err, "fetching a non-existent user ID should return an error")
}

func TestPasswordStoredAsHash(t *testing.T) {
	ts := tests.NewTestService(t)

	var raw auth.User
	user := &auth.User{
		Name:     "Hash Check",
		Email:    "hash@example.com",
		Password: "ValidPass1!",
		Role:     1,
	}
	assert.NoError(t, ts.Auth.CreateUser(user))

	err := ts.DB.First(&raw, user.ID).Error
	assert.NoError(t, err)

	assert.NotEqual(t, "ValidPass1!", string(raw.Password), "plaintext password must never be stored")
	assert.Contains(t, string(raw.Password), "$argon2id$", "stored password must be an argon2id hash")
}
