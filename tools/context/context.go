package context

import (
	"context"
	"errors"
	"time"

	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	sessionContextKey = "session"
)

type Context interface {
	context.Context
	Logger(ls ...*zap.Logger) *zap.Logger
	Sugar(ls ...*zap.Logger) *zap.SugaredLogger
	SetLogger(logger *zap.Logger)
	GetSession() *sessions.Session
	SetSession(session *sessions.Session)
	GetSessionValue(key string) (any, error)
	PutSessionValue(key string, value any) error
	ClearSession() error
	Cancel()
	WithTimeout(duration time.Duration) Context
	StartTime() time.Time
	WithFields(fields ...zapcore.Field)
	WithValue(key any, value any) error
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

func (c *ctx) Logger(ls ...*zap.Logger) *zap.Logger {
	if c == nil || c.l == nil {
		for _, l := range ls {
			if l != nil {
				return l
			}
		}
		return zap.NewNop()
	}
	return c.l
}

func (c *ctx) Sugar(ls ...*zap.Logger) *zap.SugaredLogger {
	return c.Logger(ls...).Sugar()
}

func (c *ctx) SetLogger(logger *zap.Logger) {
	c.l = logger
}

func (c *ctx) StartTime() time.Time {
	return c.st
}

func (c *ctx) WithFields(fields ...zapcore.Field) {
	c.l = c.l.With(fields...)
}

func (c *ctx) WithValue(key any, value any) error {
	if c.Value(key) != nil {
		return errors.New("key already exists in context")
	}
	c.Context = context.WithValue(c.Context, key, value)
	return nil
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

func (c *ctx) GetSession() *sessions.Session {
	session, ok := c.Value(sessionContextKey).(*sessions.Session)
	if !ok {
		return nil
	}
	return session
}

func (c *ctx) ClearSession() error {
	session, ok := c.Value(sessionContextKey).(*sessions.Session)
	if !ok {
		return errors.New("session not found in request context")
	}
	// Clear all session values
	for key := range session.Values {
		delete(session.Values, key)
	}
	// Set MaxAge to -1 to delete the cookie
	session.Options.MaxAge = -1
	c.Context = context.WithValue(c, sessionContextKey, session)
	return nil
}
