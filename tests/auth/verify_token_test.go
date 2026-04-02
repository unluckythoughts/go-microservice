package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type VerifyTokenSuite struct {
	Suite
	user  *auth.User
	token string
}

func TestVerifyTokenSuite(t *testing.T) {
	suite.Run(t, new(VerifyTokenSuite))
}

func (s *VerifyTokenSuite) SetupTest() {
	var err error
	s.user, s.token, err = s.registerAndLogin(s.T())
	s.Require().NoError(err)
	s.client.SetBearerToken(s.token)
}

func (s *VerifyTokenSuite) TearDownTest() {
	err := s.deleteUser(s.T(), s.user.ID)
	s.Assert().NoError(err)
	s.user = nil
	s.client.ClearBearerToken()
}

func (s *VerifyTokenSuite) TestVerifyToken_Success() {
	verifyToken, err := s.as.CreateVerifyToken(s.user.Email)
	s.Require().NoError(err)

	ok, status, err := s.client.VerifyToken(s.user.Email, verifyToken)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status)
	s.Assert().True(ok, "token verification should return true")
}

func (s *VerifyTokenSuite) TestVerifyToken_InvalidToken() {
	_, status, err := s.client.VerifyToken(s.user.Email, "invalid-token")
	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "invalid token must be rejected")
}

func (s *VerifyTokenSuite) TestVerifyToken_TokenAlreadyUsed() {
	verifyToken, err := s.as.CreateVerifyToken(s.user.Email)
	s.Require().NoError(err)

	// Use the token once
	_, status, err := s.client.VerifyToken(s.user.Email, verifyToken)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status)

	// Server treats re-verification as idempotent; same token returns 200 again
	_, status, err = s.client.VerifyToken(s.user.Email, verifyToken)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, status, "re-verification of the same token should return 200")
}

func (s *VerifyTokenSuite) TestVerifyToken_WrongTarget() {
	verifyToken, err := s.as.CreateVerifyToken(s.user.Email)
	s.Require().NoError(err)

	_, status, err := s.client.VerifyToken("wrong@example.com", verifyToken)
	s.Assert().NoError(err)
	s.Assert().NotEqual(http.StatusOK, status, "token for a different target must be rejected")
}
