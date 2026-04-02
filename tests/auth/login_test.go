package auth_integration_test

import (
	"net/http"

	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

func (s *AuthSuite) TestLogin_Success() {
	email, _, password, name := uniqueCredentials()

	_, status, err := s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password(password),
		Name:     name,
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status, "register should succeed")

	resp, status, err := s.client.Login(auth.Credentials{
		Email:    email,
		Password: auth.Password(password),
	})

	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status)
	s.Assert().NotEmpty(resp.Token, "expected a JWT token in the login response")
}

func (s *AuthSuite) TestLogin_InvalidPassword() {
	email, _, password, name := uniqueCredentials()

	_, status, err := s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password(password),
		Name:     name,
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status, "register should succeed")

	_, status, err = s.client.Login(auth.Credentials{
		Email:    email,
		Password: auth.Password("WrongPass1!"),
	})

	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "login with wrong password must be rejected")
}

func (s *AuthSuite) TestLogin_UnknownEmail() {
	_, status, err := s.client.Login(auth.Credentials{
		Email:    "nobody@nowhere.example.com",
		Password: auth.Password("TestPass1!"),
	})

	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status)
}
