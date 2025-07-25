package web

import (
	"net/http"
)

// Router interface implementing general router
type Router interface {
	// GET params: path, ...web.Middleware, web.Handler
	GET(string, ...interface{})
	// POST params: path, ...web.Middleware, web.Handler
	POST(string, ...interface{})
	// PUT params: path, ...web.Middleware, web.Handler
	PUT(string, ...interface{})
	// PATCH params: path, ...web.Middleware, web.Handler
	PATCH(string, ...interface{})
	// DELETE params: path, ...web.Middleware, web.Handler
	DELETE(string, ...interface{})
	// ServeFiles attaches path to root dir and serve static files
	ServeFiles(path string, root http.FileSystem)
	// Use attaches middlewares to all routes
	Use(...Middleware)
	// UseFor set middlewares for a specific path prefix
	// This allows you to apply middlewares only to routes that start with the given path prefix.
	// it does not support glob patterns, so you need to specify the exact prefix.
	UseFor(pathPrefix string, middlewares ...Middleware)
}

// Request interface implementing general server request
type Request interface {
	GetValidatedBody(ptr interface{}) error
	GetHeaders() http.Header
	GetHeader(key string) string
	GetURLParam(key string) string
	GetRouteParam(key string) string
	GetPath() string
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
