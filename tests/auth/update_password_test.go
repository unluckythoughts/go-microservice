package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type UpdatePasswordSuite struct {
	Suite
	user  *auth.User
	token string
}

func TestUpdatePasswordSuite(t *testing.T) {
	suite.Run(t, new(UpdatePasswordSuite))
}

func (s *UpdatePasswordSuite) SetupTest() {
	var err error
	s.user, s.token, err = s.registerAndLogin(s.T())
	s.Require().NoError(err)
	s.client.SetBearerToken(s.token)
}

func (s *UpdatePasswordSuite) TearDownTest() {
	err := s.deleteUser(s.T(), s.user.ID)
	s.Assert().NoError(err)
	s.user = nil
	s.client.ClearBearerToken()
}

func (s *UpdatePasswordSuite) TestUpdatePassword_Success() {
	// Create a reset token directly via the service as the email delivery is not available in tests.
	resetTarget := s.user.Email + ":email-reset-password"
	verifyToken, err := s.as.CreateVerifyToken(resetTarget)
	s.Require().NoError(err)

	_, status, err := s.client.UpdatePassword(auth.UpdatePasswordRequest{
		VerifyToken: verifyToken,
		NewPassword: auth.Password("NewPassw0rd!"),
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status)
}

func (s *UpdatePasswordSuite) TestUpdatePassword_InvalidToken() {
	_, status, err := s.client.UpdatePassword(auth.UpdatePasswordRequest{
		VerifyToken: "invalid-token",
		NewPassword: auth.Password("NewPass1!"),
	})
	s.Assert().Error(err)
	s.Assert().NotEqual(http.StatusOK, status, "invalid verify token must be rejected")
}

func (s *UpdatePasswordSuite) TestUpdatePassword_WeakNewPassword() {
	resetTarget := s.user.Email + ":email-reset-password"
	verifyToken, err := s.as.CreateVerifyToken(resetTarget)
	s.Require().NoError(err)

	_, status, err := s.client.UpdatePassword(auth.UpdatePasswordRequest{
		VerifyToken: verifyToken,
		NewPassword: auth.Password("weak"),
	})
	s.Assert().Error(err)
	s.Assert().NotEqual(http.StatusOK, status, "weak new password must be rejected")
}

func (s *UpdatePasswordSuite) TestUpdatePassword_MissingToken() {
	_, status, err := s.client.UpdatePassword(auth.UpdatePasswordRequest{
		VerifyToken: "",
		NewPassword: auth.Password("NewPass1!"),
	})
	s.Assert().Error(err)
	s.Assert().Equal(http.StatusBadRequest, status, "missing verify token must return 400")
}
