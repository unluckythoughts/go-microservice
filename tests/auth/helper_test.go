package auth_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

func uniqueCredentials() (email, username, password, name string) {
	ts := time.Now().UnixNano()
	username = fmt.Sprintf("test_user_%d", ts)
	email = fmt.Sprintf("%s@example.com", username)
	password = "TestPass12!"
	name = fmt.Sprintf("Test User_%d", ts)
	return
}

func (s *Suite) getUser(t *testing.T, email string) (*auth.User, error) {
	t.Helper()

	return s.as.GetUserByEmail(email)
}

func (s *Suite) newUser(t *testing.T) (*auth.User, error) {
	t.Helper()

	var name string
	email, _, password, name := uniqueCredentials()
	_, status, regErr := s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password(password),
		Name:     name,
	})
	if regErr != nil {
		return nil, fmt.Errorf("registration failed: %w", regErr)
	}
	if status != http.StatusCreated {
		return nil, fmt.Errorf("expected status 201, got %d", status)
	}

	return s.getUser(t, email)
}

func (s *Suite) deleteUser(t *testing.T, userID uint) error {
	t.Helper()

	return s.as.HardDeleteUser(userID)
}

func (s *Suite) loginUser(t *testing.T, email, password string) (string, error) {
	t.Helper()

	resp, status, err := s.client.Login(auth.Credentials{
		Email:    email,
		Password: auth.Password(password),
	})
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}
	if status != http.StatusCreated {
		return "", fmt.Errorf("expected status 201, got %d", status)
	}
	if resp.Token == "" {
		return "", fmt.Errorf("expected a JWT token in the login response")
	}

	return resp.Token, nil
}

func (s *Suite) registerAndLogin(t *testing.T) (user *auth.User, token string, err error) {
	t.Helper()

	user, err = s.newUser(t)
	if err != nil {
		return nil, "", fmt.Errorf("user creation failed: %w", err)
	}

	token, err = s.loginUser(t, user.Email, "TestPass12!")
	if err != nil {
		return nil, "", fmt.Errorf("login failed: %w", err)
	}

	return user, token, nil
}
