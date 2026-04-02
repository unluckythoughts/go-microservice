package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type ResetPasswordSuite struct {
	Suite
	user  *auth.User
	token string
}

func TestResetPasswordSuite(t *testing.T) {
	suite.Run(t, new(ResetPasswordSuite))
}

func (s *ResetPasswordSuite) SetupTest() {
	var err error
	s.user, s.token, err = s.registerAndLogin(s.T())
	s.Require().NoError(err)
	s.client.SetBearerToken(s.token)
}

func (s *ResetPasswordSuite) TearDownTest() {
	err := s.deleteUser(s.T(), s.user.ID)
	s.Assert().NoError(err)
	s.user = nil
	s.client.ClearBearerToken()
}

func (s *ResetPasswordSuite) TestResetPassword_Success() {
	_, status, err := s.client.ResetPassword(s.user.Email, "email")
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status)
}

func (s *ResetPasswordSuite) TestResetPassword_Unauthenticated() {
	s.client.ClearBearerToken()

	_, status, err := s.client.ResetPassword(s.user.Email, "email")
	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "unauthenticated request must be rejected")
}

func (s *ResetPasswordSuite) TestResetPassword_UnknownTarget() {
	_, status, err := s.client.ResetPassword("nobody@nowhere.example.com", "email")
	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "reset for unknown target must be rejected")
}
