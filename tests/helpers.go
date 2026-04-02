package tests

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var dbCounter uint64

// TestService bundles the services needed for unit tests.
type TestService struct {
	Auth *auth.Service
	DB   *gorm.DB
}

// NewTestService creates an in-memory SQLite auth service.
// Registers a cleanup that closes the underlying connection when the test ends.
func NewTestService(t *testing.T) *TestService {
	t.Helper()

	dbID := atomic.AddUint64(&dbCounter, 1)
	dsn := fmt.Sprintf("file:testdb_%d?mode=memory&cache=private", dbID)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err, "open in-memory sqlite")

	err = db.AutoMigrate(&auth.User{}, &auth.Verify{})
	assert.NoError(t, err, "auto-migrate auth models")

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err, "create logger")

	svc := auth.New(auth.Options{
		DB:                db,
		Logger:            logger,
		JwtKey:            "test-jwt-key-at-least-32-bytes!!",
		TokenValidInHours: 1,
	})

	return &TestService{Auth: svc, DB: db}
}

// NewTestLogger returns a no-op zap logger suitable for tests.
func NewTestLogger(t *testing.T) *zap.Logger {
	t.Helper()
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	return logger
}
