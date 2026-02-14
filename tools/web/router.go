package web

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/unluckythoughts/go-microservice/v2/utils"
	"go.uber.org/zap"
)

type (
	// Handler function for the router
	Handler    func(Request) (any, error)
	Middleware func(MiddlewareRequest) error

	router struct {
		_int         *httprouter.Router
		l            *zap.Logger
		cors         bool
		middlewares  map[string][]Middleware
		sessionStore SessionStore
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

// panicHandler panic http handler function
func (r *router) panicHandler(w http.ResponseWriter, req *http.Request, err any) {
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

// getSessionStore returns a new session store based on the provided logger
func getSessionStore(l *zap.Logger) SessionStore {
	sessionOpts := SessionOptions{}
	utils.ParseEnvironmentVars(&sessionOpts)
	sessionOpts.Logger = l
	return NewSessionStore(sessionOpts)
}

// newRouter creates a new router with the provided options
func newRouter(opts Options) *router {
	r := &router{
		_int:         httprouter.New(),
		l:            opts.Logger.Named("router"),
		cors:         opts.EnableCORS,
		sessionStore: getSessionStore(opts.Logger.Named("session")),
	}

	r.Use(SessionMiddleware(r.sessionStore))
	r.attachBasicHandlers(opts.EnableCORS)
	return r
}

// setCORSHeaders sets the CORS headers for the response
func setCORSHeaders(w http.ResponseWriter, req *http.Request) {
	origin := req.Header.Get("Origin")

	// Allow localhost origins for development
	if strings.HasPrefix(origin, "http://localhost:") ||
		strings.HasPrefix(origin, "http://127.0.0.1:") ||
		strings.HasPrefix(origin, "https://localhost:") {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else if origin == "" {
		// No origin header (non-browser requests)
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// cors http handler function
func (r *router) corsHandler(w http.ResponseWriter, req *http.Request) {
	setCORSHeaders(w, req)
	sendResponse(newResponse(w, r.newRequest(req, nil)), nil, nil, 200)
}

// attachBasicHandlers attaches basic handlers to the router
// such as NotFound, MethodNotAllowed, Panic, and healthcheck handlers.
// It also sets up CORS headers if enabled.
func (r *router) attachBasicHandlers(enableCors bool) {
	r._int.NotFound = http.HandlerFunc(r.notFoundHandler)
	r._int.MethodNotAllowed = http.HandlerFunc(r.methodNotAllowedHandler)
	r._int.PanicHandler = r.panicHandler
	r._int.GET("/_status", r.healthcheckHandler)
	r._int.GET("/_log/:message", r.log)

	if enableCors {
		r._int.GlobalOPTIONS = http.HandlerFunc(r.corsHandler)
	}
}

// getFuncName returns the name of the function
func getFuncName(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// getHandler extracts the Handler from the provided function.
// It expects the function to be of type func(Request) (any, error).
func (r *router) getHandler(f any) (Handler, bool) {
	fn, ok := f.(Handler)
	if !ok {
		fn, ok = f.(func(Request) (any, error))
		if !ok {
			r.l.Error(fmt.Sprintf("handler should be of type web.Handler or func(web.Request) (any, error) but got: %T", f))
			return nil, ok
		}
	}
	return Handler(fn), ok
}

// getMiddlewares extracts the Middleware functions from the provided slice.
// It expects each function to be of type func(MiddlewareRequest) error.
// If the slice is empty, it returns an empty slice and true.
// If any function is not of the expected type, it logs an error and returns false.
func (r *router) getMiddlewares(fns []any) (middlewares []Middleware, ok bool) {
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

// getMiddlewaresForPath returns the middlewares for a specific path.
// It first gets global middlewares, then checks for path-specific middlewares.
// and combines them.
func (r *router) getMiddlewaresForPath(path string) []Middleware {
	middlewares := r.middlewares["*"]

	for prefix, mw := range r.middlewares {
		if strings.HasPrefix(path, prefix) {
			middlewares = append(middlewares, mw...)
		}
	}

	return middlewares
}

// getRequestHandler returns a httprouter.Handle function that processes the request.
// It extracts the handler and middlewares from the provided handlers slice
// also find path specific middlewares on the router and combines them
// with the provided middlewares.
// The last element in the handlers slice is expected to be a Handler, while the rest are
// expected to be Middleware functions
func (r *router) getRouterHandlerForPath(path string, handlers []any) httprouter.Handle {
	handler, ok := r.getHandler(handlers[len(handlers)-1:][0])
	if !ok {
		panic(fmt.Errorf("last value of handlers has to be of type - web.Handler"))
	}

	middlewares, ok := r.getMiddlewares(handlers[:len(handlers)-1])
	if !ok {
		panic(fmt.Errorf("all non-last values of handlers have to be of type - web.Middleware"))
	}

	pathMiddlewares := r.getMiddlewaresForPath(path)
	middlewares = append(middlewares, pathMiddlewares...)

	return httprouter.Handle(func(w http.ResponseWriter, httpReq *http.Request, p httprouter.Params) {
		if r.cors {
			setCORSHeaders(w, httpReq)
		}

		req := r.newRequest(httpReq, p)
		resp := &response{request: req, respWriter: w}

		baseLogger := req.ctx.l
		for _, middleware := range middlewares {
			req.ctx.l = baseLogger.With(zap.String("fn", getFuncName(middleware)))
			if err := middleware(req); err != nil {
				sendResponse(resp, nil, err, 500)
				return
			}
		}

		req.ctx.l = baseLogger
		data, err := handler(req)

		session := req.ctx.Value(sessionContextKey).(*sessions.Session)
		if session != nil {
			r.sessionStore.Save(httpReq, w, session)
		}

		if err != nil {
			sendResponse(resp, nil, err, 500)
			return
		}

		sendResponse(resp, data, nil, 0)
	})
}

// Use set router level middlewares, these apply to all routes on the router
func (r *router) Use(middlewares ...Middleware) {
	if len(r.middlewares["*"]) < 1 {
		r.middlewares = make(map[string][]Middleware)
	}

	if len(middlewares) > 0 {
		r.middlewares["*"] = append(r.middlewares["*"], middlewares...)
	}
}

// UseFor set middlewares for a specific path prefix
// This allows you to apply middlewares only to routes that start with the given path prefix.
// it does not support glob patterns, so you need to specify the exact prefix.
func (r *router) UseFor(pathPrefix string, middlewares ...Middleware) {
	if len(r.middlewares[pathPrefix]) < 1 {
		r.middlewares[pathPrefix] = make([]Middleware, 0)
	}

	if len(middlewares) > 0 {
		r.middlewares[pathPrefix] = append(r.middlewares[pathPrefix], middlewares...)
	}
}

// GET attaches route with given path and handlers (...Middleware, Handler)
func (r *router) GET(path string, handlers ...any) {
	r._int.GET(path, r.getRouterHandlerForPath(path, handlers))
}

// POST attaches route with given path and handlers (...Middleware, Handler)
func (r *router) POST(path string, handlers ...any) {
	r._int.POST(path, r.getRouterHandlerForPath(path, handlers))
}

// PUT attaches route with given path and handlers (...Middleware, Handler)
func (r *router) PUT(path string, handlers ...any) {
	r._int.PUT(path, r.getRouterHandlerForPath(path, handlers))
}

// PATCH attaches route with given path and handlers (...Middleware, Handler)
func (r *router) PATCH(path string, handlers ...any) {
	r._int.PATCH(path, r.getRouterHandlerForPath(path, handlers))
}

// DELETE attaches route with given path and handlers (...Middleware, Handler)
func (r *router) DELETE(path string, handlers ...any) {
	r._int.DELETE(path, r.getRouterHandlerForPath(path, handlers))
}

// ServeFiles attaches path to root dir and serve static files
func (r *router) ServeFiles(path string, root http.FileSystem) {
	r._int.ServeFiles(path, root)
}
