package psql

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
)

var (
	spacePattern = regexp.MustCompile(`(\t|\n)+`)
)

type dbLogger struct {
	l *zap.Logger
}

func (dbl *dbLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return &dbLogger{}
}

func (dbl *dbLogger) Info(ctx context.Context, s string, v ...interface{}) {
	dbl.l.Info(fmt.Sprintf(spacePattern.ReplaceAllString(s, " "), v...))
}

func (dbl *dbLogger) Warn(ctx context.Context, s string, v ...interface{}) {
	dbl.l.Warn(fmt.Sprintf(spacePattern.ReplaceAllString(s, " "), v...))
}

func (dbl *dbLogger) Error(ctx context.Context, s string, v ...interface{}) {
	dbl.l.Error(fmt.Sprintf(spacePattern.ReplaceAllString(s, " "), v...))
}

func (dbl *dbLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rows int64), err error) {
	elapsed := time.Since(begin)

	sql, rows := fc()
	sql = spacePattern.ReplaceAllString(sql, " ")

	dbl.l.With(
		zap.String("type", "query"),
		zap.String("duration", elapsed.String()),
		zap.String("query", sql),
		zap.Int64("count", rows),
	).Debug("ran db query")
}
