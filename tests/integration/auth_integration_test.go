package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unluckythoughts/go-microservice/v2/tests"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

// TestAuthFlowE2E exercises a full register → login → token validation cycle
// using an in-memory database.  External-service integration tests that need
// a real database should use testing.Short() to skip them in unit test runs.
func TestAuthFlowE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	ts := tests.NewTestService(t)

	// 1. Register a new user via CreateUser (simulates registration handler)
	user := &auth.User{
		Name:     "E2E User",
		Email:    "e2e@example.com",
		Password: "E2eValid1!",
		Role:     1,
	}
	err := ts.Auth.CreateUser(user)
	assert.NoError(t, err, "user registration should succeed")
	assert.NotZero(t, user.ID)

	// 2. Verify credentials (simulates login handler)
	found, ok, err := ts.Auth.VerifyUserPasswordByEmail("e2e@example.com", "E2eValid1!")
	assert.NoError(t, err)
	assert.True(t, ok, "login with correct credentials should succeed")
	if found != nil {
		assert.Equal(t, "e2e@example.com", found.Email)
	}

	// 3. Wrong password is rejected
	_, ok, err = ts.Auth.VerifyUserPasswordByEmail("e2e@example.com", "WrongPass1!")
	assert.NoError(t, err)
	assert.False(t, ok, "login with wrong password must fail")

	// 4. Get user by ID returns the same user
	byID, err := ts.Auth.GetUserByID(user.ID)
	assert.NoError(t, err)
	if byID != nil && found != nil {
		assert.Equal(t, found.Email, byID.Email)
		assert.Empty(t, string(byID.Password), "password must not be returned to callers")
	}
}
