package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/unluckythoughts/go-microservice/tools/sockets"
	"go.uber.org/zap"
)

type (
	// Server struct for the simple http server
	Server struct {
		addr           string
		socketPath     string
		logger         *zap.Logger
		router         *router
		socketServer   *sockets.Server
		proxyTransport ProxyTransport
	}

	Options struct {
		Logger      *zap.Logger
		Port        int    `env:"WEB_PORT" envDefault:"8080"`
		SocketPath  string `env:"WEB_SOCKET_PATH" envDefault:"/socket"`
		WorkerCount int    `env:"WEB_WORKER_COUNT" envDefault:"20"`
		EnableCORS  bool   `env:"WEB_CORS" envDefault:"false"`
		EnableProxy bool   `env:"WEB_PROXY" envDefault:"false"`
	}

	ProxyTransport func(l *zap.Logger) http.RoundTripper
)

func (s *Server) setupRouter() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/_proxy") {
			if s.proxyTransport != nil {
				proxyHandler(s.logger, s.proxyTransport(s.logger))(w, r)
			} else {
				proxyHandler(s.logger, http.DefaultTransport)(w, r)
			}

		} else if strings.HasPrefix(r.URL.Path, s.socketPath) {
			http.HandlerFunc(s.upgradeConnection)(w, r)
		} else {
			s.router._int.ServeHTTP(w, r)
		}
	})
}

// NewServer returns a new server object
func NewServer(opts Options) *Server {
	socketServer := sockets.New(opts.Logger, opts.WorkerCount)
	s := &Server{
		addr:         ":" + strconv.Itoa(opts.Port),
		logger:       opts.Logger,
		socketPath:   opts.SocketPath,
		router:       newRouter(opts),
		socketServer: socketServer,
	}

	return s
}

func (s *Server) SetProxyTransport(rt ProxyTransport) {
	s.proxyTransport = rt
}

// Start runs http listener on the given address
func (s *Server) Start() {
	s.socketServer.StartSocketWorkers()
	s.logger.Info("Web server started on address: " + s.addr)
	s.logger.Fatal(http.ListenAndServe(s.addr, s.setupRouter()).Error())
}

// GetRouter returns the router instance
func (s *Server) GetRouter() Router {
	return s.router
}
