package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type LogoutSuite struct {
	Suite
	user *auth.User
}

func TestLogoutSuite(t *testing.T) {
	suite.Run(t, new(LogoutSuite))
}

func (s *LogoutSuite) SetupTest() {
	user, token, err := s.registerAndLogin(s.T())
	s.Assert().NoError(err)
	s.Assert().NotNil(user)
	s.user = user
	s.client.SetBearerToken(token)
}

func (s *LogoutSuite) TearDownTest() {
	err := s.deleteUser(s.T(), s.user.ID)
	s.Assert().NoError(err)
	s.user = nil
	s.client.ClearBearerToken()
}

func (s *LogoutSuite) TestLogout_Success() {
	_, status, err := s.client.Logout()
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status)
}

func (s *LogoutSuite) TestLogout_Unauthenticated() {
	s.client.ClearBearerToken()

	_, status, err := s.client.Logout()
	s.Assert().Error(err)
	s.Assert().NotEqual(http.StatusOK, status, "logout without authentication must be rejected")
}

func (s *LogoutSuite) TestLogout_ProtectedRouteRejectedAfterTokenCleared() {
	_, status, err := s.client.Logout()
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status)

	s.client.ClearBearerToken()

	_, status, err = s.client.GetUser()
	s.Assert().Error(err)
	s.Assert().NotEqual(http.StatusOK, status, "accessing a protected route without a token must be rejected")
}
