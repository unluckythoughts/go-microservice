package web

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Context interface {
	context.Context
	Logger() *zap.SugaredLogger
}

type ctx struct {
	context.Context
	st time.Time
	l  *zap.Logger
}

func NewContext(l *zap.Logger) *ctx {
	return &ctx{
		st:      time.Now(),
		l:       l,
		Context: context.Background(),
	}
}

func (c *ctx) Logger() *zap.SugaredLogger {
	return c.l.Sugar()
}

func (c *ctx) GoContext() context.Context {
	return c.Context
}
