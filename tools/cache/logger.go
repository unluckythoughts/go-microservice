package cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type cacheLogger struct {
	l *zap.Logger
}

func (l *cacheLogger) Printf(ctx context.Context, format string, v ...interface{}) {
	l.l.Debug(fmt.Sprintf(format, v...))
}

func (l *cacheLogger) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (l *cacheLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	l.l.Debug(cmd.String())
	return nil
}

func (l *cacheLogger) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (l *cacheLogger) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	var strCmds []string
	for _, cmd := range cmds {
		strCmds = append(strCmds, cmd.String())
	}

	l.l.Debug(strings.Join(strCmds, "; "))
	return nil
}
