package db

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Type uint

const (
	Postgres Type = 1
	SQLite   Type = 2
)

type Options struct {
	Type   Type `env:"DB_TYPE" envDefault:"1"`
	Debug  bool `env:"DB_DEBUG" envDefault:"false"`
	Logger *zap.Logger
	// for SQLite
	FilePath string `env:"DB_FILE_PATH" envDefault:"./data/sqlite.db"`
	// for Postgres
	Host           string `env:"DB_HOST" envDefault:"localhost"`
	Port           int    `env:"DB_PORT" envDefault:"5432"`
	User           string `env:"DB_USER" envDefault:"postgres"`
	Password       string `env:"DB_PASSWORD" envDefault:"password"`
	Name           string `env:"DB_NAME" envDefault:"postgres"`
	ConnectTimeout int    `env:"DB_CONNECT_TIMEOUT" envDefault:"10"`
	SSLMode        string `env:"DB_SSL_MODE" envDefault:"disable"`
}

func sanityCheck(db *gorm.DB) {
	err := db.Exec("SELECT 1").Error
	if err != nil {
		panic(fmt.Errorf("could not sanity check on db: %w", err))
	}
}

func getConnectionString(opts Options) string {
	return fmt.Sprintf(
		`host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d`,
		opts.Host, opts.Port, opts.User, opts.Password, opts.Name, opts.SSLMode, opts.ConnectTimeout,
	)
}

func newPostgres(opts Options) *gorm.DB {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  getConnectionString(opts),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("could not connect to db: %w", err))
	}
	return db
}

func newSQLite(opts Options) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(opts.FilePath), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("could not connect to db: %w", err))
	}

	return db
}

func New(opts Options) (db *gorm.DB) {
	l := opts.Logger
	switch opts.Type {
	case Postgres:
		db = newPostgres(opts)
	case SQLite:
		db = newSQLite(opts)
		l = opts.Logger.WithOptions(zap.AddCallerSkip(3))
	default:
		panic(fmt.Errorf("incorrect db type: %d", opts.Type))
	}

	if opts.Debug {
		db = db.Session(&gorm.Session{Logger: &dbLogger{l: l}})
	} else {
		db = db.Session(&gorm.Session{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})
	}
	sanityCheck(db)
	opts.Logger.Info("Connected to db")
	return db
}
