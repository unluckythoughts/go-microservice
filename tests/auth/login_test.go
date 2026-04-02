package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type LoginSuite struct {
	Suite
}

func TestLoginSuite(t *testing.T) {
	suite.Run(t, new(LoginSuite))
}

func (s *LoginSuite) TestLogin_Success() {
	email, _, password, name := uniqueCredentials()

	_, status, err := s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password(password),
		Name:     name,
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusCreated, status, "register should succeed")

	resp, status, err := s.client.Login(auth.Credentials{
		Email:    email,
		Password: auth.Password(password),
	})

	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusCreated, status)
	s.Assert().NotEmpty(resp.Token, "expected a JWT token in the login response")
}

func (s *LoginSuite) TestLogin_InvalidPassword() {
	email, _, password, name := uniqueCredentials()

	_, status, err := s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password(password),
		Name:     name,
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusCreated, status, "register should succeed")

	_, status, err = s.client.Login(auth.Credentials{
		Email:    email,
		Password: auth.Password("WrongPass1!"),
	})

	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "login with wrong password must be rejected")
}

func (s *LoginSuite) TestLogin_UnknownEmail() {
	_, status, err := s.client.Login(auth.Credentials{
		Email:    "nobody@nowhere.example.com",
		Password: auth.Password("TestPass1!"),
	})

	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status)
}
