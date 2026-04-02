package auth_test

import (
"net/http"
"testing"

"github.com/stretchr/testify/suite"
"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type GetUserSuite struct {
Suite
user  *auth.User
token string
}

func TestGetUserSuite(t *testing.T) {
suite.Run(t, new(GetUserSuite))
}

func (s *GetUserSuite) SetupTest() {
var err error
s.user, s.token, err = s.registerAndLogin(s.T())
s.Require().NoError(err)
s.client.SetBearerToken(s.token)
}

func (s *GetUserSuite) TearDownTest() {
err := s.deleteUser(s.T(), s.user.ID)
s.Assert().NoError(err)
s.user = nil
s.client.ClearBearerToken()
}

func (s *GetUserSuite) TestGetUser_Success() {
_, status, err := s.client.GetUser()
s.Assert().NoError(err)
s.Assert().Equal(http.StatusOK, status)
}

func (s *GetUserSuite) TestGetUser_Unauthenticated() {
s.client.ClearBearerToken()

_, status, err := s.client.GetUser()
s.Assert().NoError(err)
s.Assert().NotEqual(http.StatusOK, status, "unauthenticated request must be rejected")
}
