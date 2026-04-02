package auth_integration_test

import (
	"fmt"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/examples/microservice/client"
)

type Suite struct {
	suite.Suite
	client *client.Client
}

func (s *Suite) SetupSuite() {
	s.client = client.NewClient("http://localhost:8080/api/v1/")
}

func uniqueCredentials() (email, username, password, name string) {
	ts := time.Now().UnixNano()
	username = fmt.Sprintf("user%d", ts)
	email = fmt.Sprintf("%s@example.com", username)
	password = "TestPass1!"
	name = fmt.Sprintf("Test User %d", ts)
	return
}
