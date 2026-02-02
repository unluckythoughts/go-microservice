package web

import (
	"context"
	"errors"
	"time"

	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

const (
	sessionContextKey = "session"
)

type Context interface {
	context.Context
	Logger() *zap.SugaredLogger
	SetSession(session *sessions.Session)
	GetSessionValue(key string) (any, error)
	PutSessionValue(key string, value any) error
	Cancel()
	WithTimeout(duration time.Duration) Context
}

type ctx struct {
	context.Context
	st     time.Time
	l      *zap.Logger
	cancel context.CancelFunc
}

func IsWebContext(c context.Context) (Context, bool) {
	ctx, ok := c.(Context)
	return ctx, ok
}

func NewContext(l *zap.Logger) *ctx {
	c, cancel := context.WithCancel(context.Background())
	return &ctx{
		st:      time.Now(),
		l:       l,
		Context: c,
		cancel:  cancel,
	}
}

func (c *ctx) WithTimeout(duration time.Duration) Context {
	contextWithTimeout, cancel := context.WithTimeout(c.Context, duration)
	return &ctx{
		Context: contextWithTimeout,
		st:      c.st,
		l:       c.l,
		cancel:  cancel,
	}
}

func (c *ctx) Cancel() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *ctx) Logger() *zap.SugaredLogger {
	return c.l.Sugar()
}

func (c *ctx) GetSessionValue(key string) (any, error) {
	session, ok := c.Value(sessionContextKey).(*sessions.Session)
	if !ok {
		return nil, errors.New("session not found in request context")
	}

	value, ok := session.Values[key]
	if !ok {
		return nil, errors.New("key not found in session")
	}

	return value, nil
}

func (c *ctx) PutSessionValue(key string, value any) error {
	session, ok := c.Value(sessionContextKey).(*sessions.Session)
	if !ok {
		return errors.New("session not found in request context")
	}
	session.Values[key] = value

	c.Context = context.WithValue(c, sessionContextKey, session)

	return nil
}

func (c *ctx) SetSession(session *sessions.Session) {
	c.Context = context.WithValue(c, sessionContextKey, session)
}
