package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type ChangePasswordSuite struct {
	Suite
	user  *auth.User
	token string
}

func TestChangePasswordSuite(t *testing.T) {
	suite.Run(t, new(ChangePasswordSuite))
}

func (s *ChangePasswordSuite) SetupTest() {
	var err error
	s.user, s.token, err = s.registerAndLogin(s.T())
	s.Require().NoError(err)
	s.client.SetBearerToken(s.token)
}

func (s *ChangePasswordSuite) TearDownTest() {
	err := s.deleteUser(s.T(), s.user.ID)
	s.Assert().NoError(err)
	s.user = nil
	s.client.ClearBearerToken()
}

func (s *ChangePasswordSuite) TestChangePassword_Success() {
	_, status, err := s.client.ChangePassword(auth.ChangePasswordRequest{
		OldPassword: auth.Password("TestPass12!"),
		NewPassword: auth.Password("NewPassw0rd!"),
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status)
}

func (s *ChangePasswordSuite) TestChangePassword_WrongOldPassword() {
	_, status, err := s.client.ChangePassword(auth.ChangePasswordRequest{
		OldPassword: auth.Password("WrongPass1!"),
		NewPassword: auth.Password("NewPass1!"),
	})
	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "wrong old password must be rejected")
}

func (s *ChangePasswordSuite) TestChangePassword_WeakNewPassword() {
	_, status, err := s.client.ChangePassword(auth.ChangePasswordRequest{
		OldPassword: auth.Password("TestPass12!"),
		NewPassword: auth.Password("weak"),
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusBadRequest, status, "weak new password must return 400")
}

func (s *ChangePasswordSuite) TestChangePassword_Unauthenticated() {
	s.client.ClearBearerToken()

	_, status, err := s.client.ChangePassword(auth.ChangePasswordRequest{
		OldPassword: auth.Password("TestPass12!"),
		NewPassword: auth.Password("NewPass1!"),
	})
	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "unauthenticated request must be rejected")
}
