package web

import (
	"net/http"
)

// Router interface implementing general router
type Router interface {
	// GET params: path, web.Handler, ...web.Middleware
	GET(string, ...interface{})
	// POST params: path, web.Handler, ...web.Middleware
	POST(string, ...interface{})
	// PUT params: path, web.Handler, ...web.Middleware
	PUT(string, ...interface{})
	// PATCH params: path, web.Handler, ...web.Middleware
	PATCH(string, ...interface{})
	// DELETE params: path, web.Handler, ...web.Middleware
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
	GetContext() Context
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
