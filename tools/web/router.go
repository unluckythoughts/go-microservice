package web

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type (
	// Handler function for the router
	Handler    func(Request) (interface{}, error)
	Middleware func(MiddlewareRequest) error

	router struct {
		_int        *httprouter.Router
		l           *zap.Logger
		middlewares []Middleware
	}
)

// notFoundHandler 404 http handler function
func (r *router) notFoundHandler(w http.ResponseWriter, req *http.Request) {
	msg := fmt.Sprintf("api route for %s %s not found", req.Method, req.URL.String())
	sendResponse(newResponse(w, r.newRequest(req, nil)), nil, errors.New(msg), 404)
}

// methodNotAllowedHandler 405 http handler function
func (r *router) methodNotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	msg := "not allowed"
	sendResponse(newResponse(w, r.newRequest(req, nil)), nil, errors.New(msg), 405)
}

// unAuthorizedHandler 401 http handler function
func (r *router) unAuthorizedHandler(w http.ResponseWriter, req *http.Request) {
	msg := "not authorized"
	sendResponse(newResponse(w, r.newRequest(req, nil)), nil, errors.New(msg), 401)
}

// panicHandler panic http handler function
func (r *router) panicHandler(w http.ResponseWriter, req *http.Request, err interface{}) {
	panicErr := errors.New(err.(error).Error())
	sendResponse(newResponse(w, r.newRequest(req, nil)), nil, panicErr, 500)
}

// healthcheckHandler healthcheck handler function
func (r *router) healthcheckHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

// healthcheckHandler healthcheck handler function
func (r *router) log(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	r.l.Info(p.ByName("message"))
}

// notImplemented to be developed function handler
func (r *router) notImplemented(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	msg := "not implemented yet"
	sendResponse(newResponse(w, r.newRequest(req, nil)), nil, errors.New(msg), 400)
}

func newRouter(l *zap.Logger, socketPath string) *router {
	r := &router{
		_int: httprouter.New(),
		l:    l,
	}

	r.attachBasicHandlers()
	return r
}

func (r *router) attachBasicHandlers() {
	r._int.NotFound = http.HandlerFunc(r.notFoundHandler)
	r._int.MethodNotAllowed = http.HandlerFunc(r.methodNotAllowedHandler)
	r._int.PanicHandler = r.panicHandler
	r._int.GET("/_status", r.healthcheckHandler)
	r._int.GET("/_log/:message", r.log)
}

// GetFuncName returns the name of the function
func GetFuncName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func (r *router) getHandler(f interface{}) (Handler, bool) {
	fn, ok := f.(func(Request) (interface{}, error))
	if !ok {
		r.l.Error("handler should be of type func(web.Request) (interface{}, error)")
		return nil, ok
	}
	return Handler(fn), ok
}

func (r *router) getMiddlewares(fns []interface{}) (middlewares []Middleware, ok bool) {
	if len(fns) < 1 {
		return middlewares, true
	}

	for _, fn := range fns {
		middleware, ok := fn.(func(MiddlewareRequest) error)
		if !ok {
			r.l.Error("middleware should be of type web.Middleware")
			return middlewares, false
		}
		middlewares = append(middlewares, Middleware(middleware))
	}

	return middlewares, true
}

func (r *router) routerHandler(handlers []interface{}) httprouter.Handle {
	handler, ok := r.getHandler(handlers[len(handlers)-1:][0])
	if !ok {
		panic(fmt.Sprintf("last value of handlers has to be of type - web.Handler"))
	}

	middlewares, ok := r.getMiddlewares(handlers[:len(handlers)-1])
	if !ok {
		panic(fmt.Sprintf("all non-last values of handlers have to be of type - web.Middleware"))
	}

	return httprouter.Handle(func(w http.ResponseWriter, httpReq *http.Request, p httprouter.Params) {
		req := r.newRequest(httpReq, p)

		baseLogger := req.Logger()
		for _, middleware := range append(r.middlewares, middlewares...) {
			l := baseLogger.With(zap.String("middleware", GetFuncName(middleware)))
			req.SetLogger(l)
			if err := middleware(req); err != nil {
				sendResponse(&response{request: req, respWriter: w}, nil, err, 500)
				return
			}
		}

		resp := &response{
			request:    req,
			respWriter: w,
		}

		l := baseLogger.With(zap.String("handler", GetFuncName(handler)))
		req.SetLogger(l)
		data, err := handler(req)
		if err != nil {
			sendResponse(resp, nil, err, 500)
			return
		}

		sendResponse(resp, data, nil, 0)
	})
}

// Use set router level middlewares, these apply to all routes on the router
func (r *router) Use(middlewares ...Middleware) {
	if len(middlewares) > 0 {
		r.middlewares = append(r.middlewares, middlewares...)
	}
}

// GET attaches route with given path and handlers (...Middleware, Handler)
func (r *router) GET(path string, handlers ...interface{}) {
	r._int.GET(path, r.routerHandler(handlers))
}

// POST attaches route with given path and handlers (...Middleware, Handler)
func (r *router) POST(path string, handlers ...interface{}) {
	r._int.POST(path, r.routerHandler(handlers))
}

// PUT attaches route with given path and handlers (...Middleware, Handler)
func (r *router) PUT(path string, handlers ...interface{}) {
	r._int.PUT(path, r.routerHandler(handlers))
}

// PATCH attaches route with given path and handlers (...Middleware, Handler)
func (r *router) PATCH(path string, handlers ...interface{}) {
	r._int.PATCH(path, r.routerHandler(handlers))
}

// DELETE attaches route with given path and handlers (...Middleware, Handler)
func (r *router) DELETE(path string, handlers ...interface{}) {
	r._int.DELETE(path, r.routerHandler(handlers))
}
