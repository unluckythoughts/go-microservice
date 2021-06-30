package psql

import (
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type (
	Options struct {
		Logger         *zap.Logger
		Host           string `env:"DB_HOST" envDefault:"localhost"`
		Port           int    `env:"DB_PORT" envDefault:"5432"`
		SSLMode        string `env:"DB_SSLMODE" envDefault:"disable"`
		ConnectTimeout int    `env:"DB_CONNECT_TIMEOUT" envDefault:"10"`
		DebugMode      bool   `env:"DB_DEBUG_MODE" envDefault:"true"`
		Debug          bool   `env:"DB_DEBUG" envDefault:"false"`
		User           string `env:"DB_USER"`
		Name           string `env:"DB_NAME"`
		Password       string `env:"DB_PASSWORD"`
	}
)

func getConnectionString(opts Options) string {
	return fmt.Sprintf(
		`host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d`,
		opts.Host, opts.Port, opts.User, opts.Password, opts.Name, opts.SSLMode, opts.ConnectTimeout,
	)
}

func sanityCheck(db *gorm.DB) {
	err := db.Exec("SELECT 1").Error
	if err != nil {
		panic(errors.Wrap(err, "could not connect to postgres"))
	}
}

func New(opts Options) *gorm.DB {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  getConnectionString(opts),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		panic(errors.Wrapf(err, "could not connect to db"))
	}

	l := opts.Logger
	if opts.Debug {
		l = l.WithOptions(zap.IncreaseLevel(zapcore.DebugLevel))
	} else {
		l = l.WithOptions(zap.IncreaseLevel(zapcore.InfoLevel))
	}

	db.Logger = &dbLogger{l: l}
	sanityCheck(db)
	opts.Logger.Info("Connected to db")
	return db
}
