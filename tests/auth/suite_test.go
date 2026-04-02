package auth_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/examples/microservice/client"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
	"github.com/unluckythoughts/go-microservice/v2/tools/logger"
	"github.com/unluckythoughts/go-microservice/v2/tools/psql"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Suite struct {
	suite.Suite
	client *client.Client
	db     *gorm.DB
	as     *auth.Service
}

func initializeTestDB(l *zap.Logger) *gorm.DB {
	return psql.New(psql.Options{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "test",
		Name:     "test",
		SSLMode:  "disable",
		Logger:   l.Named("db"),
	})
}

func (s *Suite) SetupSuite() {
	s.client = client.NewClient("http://localhost:8080/api/v1/")
	// Initialize the database connection here if needed
	l := logger.New(logger.Options{
		LogLevel: zap.ErrorLevel.String(),
	})
	s.db = initializeTestDB(l)
	s.as = auth.New(auth.Options{
		DB:     s.db,
		Logger: l.Named("auth"),
	})
}
