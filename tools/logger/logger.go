package logger

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"debug"`
	Fields   map[string]interface{}
}

var (
	defaultLoggerConfig = zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Encoding:    "json",
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "m",
			LevelKey:       "l",
			TimeKey:        "ts",
			CallerKey:      "c",
			StacktraceKey:  "trace",
			FunctionKey:    "fn",
			NameKey:        "s",
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
)

func New(opts Options) *zap.Logger {
	defaultLoggerConfig.Level = getLogLevel(opts.LogLevel)
	defaultLoggerConfig.InitialFields = opts.Fields
	l, err := defaultLoggerConfig.Build()
	if err != nil {
		panic(errors.Wrapf(err, "could not create logger"))
	}

	return l
}

func getLogLevel(level string) zap.AtomicLevel {
	var zapLevel zapcore.Level

	switch level {
	case zapcore.InfoLevel.String():
		zapLevel = zapcore.InfoLevel
	case zapcore.WarnLevel.String():
		zapLevel = zapcore.WarnLevel
	case zapcore.ErrorLevel.String():
		zapLevel = zapcore.ErrorLevel
	case zapcore.DPanicLevel.String():
		zapLevel = zapcore.DPanicLevel
	case zapcore.PanicLevel.String():
		zapLevel = zapcore.PanicLevel
	case zapcore.FatalLevel.String():
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.DebugLevel
	}

	return zap.NewAtomicLevelAt(zapLevel)
}
