package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type UpdateUserSuite struct {
	Suite
	user  *auth.User
	token string
}

func TestUpdateUserSuite(t *testing.T) {
	suite.Run(t, new(UpdateUserSuite))
}

func (s *UpdateUserSuite) SetupTest() {
	var err error
	s.user, s.token, err = s.registerAndLogin(s.T())
	s.Require().NoError(err)
	s.client.SetBearerToken(s.token)
}

func (s *UpdateUserSuite) TearDownTest() {
	err := s.deleteUser(s.T(), s.user.ID)
	s.Assert().NoError(err)
	s.user = nil
	s.client.ClearBearerToken()
}

func (s *UpdateUserSuite) TestUpdateUser_Success() {
	_, status, err := s.client.UpdateUser(auth.UpdateUserRequest{
		Name: "Updated Name",
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status)
}

func (s *UpdateUserSuite) TestUpdateUser_Unauthenticated() {
	s.client.ClearBearerToken()

	_, status, err := s.client.UpdateUser(auth.UpdateUserRequest{
		Name: "Updated Name",
	})
	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "unauthenticated request must be rejected")
}

func (s *UpdateUserSuite) TestUpdateUser_MissingName() {
	_, status, err := s.client.UpdateUser(auth.UpdateUserRequest{
		Name: "",
	})
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusBadRequest, status, "missing name must return 400")
}
