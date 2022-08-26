package sqlite

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type (
	Options struct {
		Logger   *zap.Logger
		Filepath string `env:"DB_FILE_PATH" envDefault:"gorm.db"`
		Debug    bool   `env:"DB_DEBUG" envDefault:"false"`
	}
)

func sanityCheck(db *gorm.DB) {
	err := db.Exec("SELECT 1").Error
	if err != nil {
		panic(errors.Wrap(err, "could not connect to postgres"))
	}
}

func New(opts Options) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(opts.Filepath), &gorm.Config{})
	if err != nil {
		panic(errors.Wrapf(err, "could not connect to db"))
	}

	l := opts.Logger.WithOptions(zap.AddCallerSkip(3))
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
