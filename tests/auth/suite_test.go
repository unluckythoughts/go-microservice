package auth_integration_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/examples/microservice/client"
)

type AuthSuite struct {
	suite.Suite
	client *client.Client
}

func (s *AuthSuite) SetupSuite() {
	s.client = client.NewClient("http://localhost:8080/api/v1/")
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}

func uniqueCredentials() (email, username, password, name string) {
	ts := time.Now().UnixNano()
	username = fmt.Sprintf("user%d", ts)
	email = fmt.Sprintf("%s@example.com", username)
	password = "TestPass1!"
	name = fmt.Sprintf("Test User %d", ts)
	return
}
