package web

import (
	"net/http"
	"strconv"

	"github.com/unluckythoughts/go-microservice/tools/sockets"
	"go.uber.org/zap"
)

type (
	// Server struct for the simple http server
	Server struct {
		addr         string
		logger       *zap.Logger
		router       *router
		socketServer *sockets.Server
	}

	Options struct {
		Logger      *zap.Logger
		Port        int    `env:"WEB_PORT" envDefault:"8080"`
		SocketPath  string `env:"WEB_SOCKET_PATH" envDefault:"/socket"`
		WorkerCount int    `env:"WEB_WORKER_COUNT" envDefault:"20"`
		EnableCORS  bool   `env:"WEB_CORS" envDefault:"false"`
	}
)

// NewServer returns a new server object
func NewServer(opts Options) *Server {
	socketServer := sockets.New(opts.Logger, opts.WorkerCount)
	s := &Server{
		addr:         ":" + strconv.Itoa(opts.Port),
		logger:       opts.Logger,
		router:       newRouter(opts.Logger),
		socketServer: socketServer,
	}

	http.Handle("/", s.router._int)
	http.Handle(opts.SocketPath, http.HandlerFunc(s.upgradeConnection))

	if opts.EnableCORS {
		s.router._int.GlobalOPTIONS = http.HandlerFunc(s.router.corsHandler)
	}

	return s
}

// Start runs http listener on the given address
func (s *Server) Start() {
	s.socketServer.StartSocketWorkers()
	s.logger.Info("Web server started on address: " + s.addr)
	s.logger.Fatal(http.ListenAndServe(s.addr, nil).Error())
}

// GetRouter returns the router instance
func (s *Server) GetRouter() Router {
	return s.router
}
