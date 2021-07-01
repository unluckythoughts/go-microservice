package web

import (
	"net/http"

	"go.uber.org/zap"
)

// Router interface implementing general router
type Router interface {
	GET(string, ...interface{})
	POST(string, ...interface{})
	PUT(string, ...interface{})
	PATCH(string, ...interface{})
	DELETE(string, ...interface{})
	Use(...Middleware)
}

// Request interface implementing general server request
type Request interface {
	GetValidatedBody(ptr interface{}) error
	GetHeaders() http.Header
	GetHeader(key string) string
	GetURLParam(key string) string
	GetRouteParam(key string) string
	GetContextValue(key string) interface{}
	GetContext() map[string]interface{}
	Logger() *zap.Logger
	SetLogger(l *zap.Logger)
}

// MiddlewareRequest interface implementing general middleware request
type MiddlewareRequest interface {
	Request
	SetContextValue(key string, value interface{}) error
}

// Response interface implementing general server response
type Response interface {
	AddHeader(key string, values ...string)
}