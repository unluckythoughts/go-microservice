package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type RegisterSuite struct {
	Suite
}

func TestRegisterSuite(t *testing.T) {
	suite.Run(t, new(RegisterSuite))
}

func (s *RegisterSuite) TestRegister_Success() {
	email, _, password, name := uniqueCredentials()

	_, status, err := s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password(password),
		Name:     name,
	})

	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status)
}

func (s *RegisterSuite) TestRegister_DuplicateEmail() {
	email, _, password, name := uniqueCredentials()

	_, status, err := s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password(password),
		Name:     name,
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status, "first registration should succeed")

	_, status, err = s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password(password),
		Name:     name,
	})

	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "duplicate email registration must be rejected")
}

func (s *RegisterSuite) TestRegister_MissingName() {
	email, _, password, _ := uniqueCredentials()

	_, status, err := s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password(password),
		Name:     "",
	})

	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusBadRequest, status, "registration without a name must return 400")
}

func (s *RegisterSuite) TestRegister_InvalidPassword() {
	email, _, _, name := uniqueCredentials()

	_, status, err := s.client.Register(auth.RegisterRequest{
		Email:    email,
		Password: auth.Password("weak"),
		Name:     name,
	})

	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusBadRequest, status, "registration with a weak password must return 400")
}

func (s *RegisterSuite) TestRegister_InvalidEmail() {
	_, _, password, name := uniqueCredentials()

	_, status, err := s.client.Register(auth.RegisterRequest{
		Email:    "not-an-email",
		Password: auth.Password(password),
		Name:     name,
	})

	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusBadRequest, status, "registration with an invalid email must return 400")
}
