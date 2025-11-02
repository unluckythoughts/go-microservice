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
	RemoveSession() error
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

func (c *ctx) RemoveSession() error {
	session, ok := c.Value(sessionContextKey).(*sessions.Session)
	if !ok {
		return errors.New("session not found in request context")
	}

	session.Options.MaxAge = -1
	c.Context = context.WithValue(c, sessionContextKey, session)

	return nil
}
