package logger

import (
	"fmt"

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
		Development: false,
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "m",
			LevelKey:       "l",
			TimeKey:        "ts",
			CallerKey:      "c",
			StacktraceKey:  "tr",
			FunctionKey:    zapcore.OmitKey,
			NameKey:        "n",
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeLevel:    customLevelEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
)

func customLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString("D")
	case zapcore.InfoLevel:
		enc.AppendString("I")
	case zapcore.WarnLevel:
		enc.AppendString("W")
	case zapcore.ErrorLevel:
		enc.AppendString("E")
	case zapcore.DPanicLevel:
		enc.AppendString("C")
	case zapcore.PanicLevel:
		enc.AppendString("P")
	case zapcore.FatalLevel:
		enc.AppendString("F")
	default:
		enc.AppendString(fmt.Sprintf("Level(%d)", l))
	}
}

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
