package auth_test

import (
	"os"

	"github.com/stretchr/testify/suite"
	"github.com/unluckythoughts/go-microservice/v2/examples/microservice/client"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
	"github.com/unluckythoughts/go-microservice/v2/tools/db"
	"github.com/unluckythoughts/go-microservice/v2/tools/logger"
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
	host := "localhost"
	if os.Getenv("SERVICE_DB_HOST") != "" {
		host = os.Getenv("SERVICE_DB_HOST")
	}

	return db.New(db.Options{
		Host:     host,
		Port:     5432,
		User:     "test",
		Password: "test",
		Name:     "test",
		SSLMode:  "disable",
		Logger:   l.Named("db"),
	})
}

func (s *Suite) SetupSuite() {
	baseURL := os.Getenv("SERVICE_ENDPOINT_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api/v1"
	}

	s.client = client.NewClient(baseURL)
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
