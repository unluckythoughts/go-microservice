package sockets

import (
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"go.uber.org/zap"
)

type Server struct {
	l            *zap.Logger
	connMutex    sync.RWMutex
	handlerMutex sync.RWMutex
	connections  map[string]net.Conn
	handlers     map[string]Handler
	workerCount  int
}

func New(l *zap.Logger, count int) *Server {
	s := &Server{
		l:            l,
		connMutex:    sync.RWMutex{},
		handlerMutex: sync.RWMutex{},
		connections:  make(map[string]net.Conn),
		handlers:     make(map[string]Handler),
		workerCount:  20,
	}

	if count > 0 {
		s.workerCount = count
	}

	s.AddHandler("_status", func(r Request) (data interface{}, err error) {
		return "ok", nil
	})

	return s
}

func (s *Server) addConn(conn net.Conn) {
	defer s.connMutex.Unlock()
	s.connMutex.Lock()

	s.connections[conn.RemoteAddr().String()] = conn
}

func (s *Server) delConn(conn net.Conn) {
	defer s.connMutex.Unlock()
	s.connMutex.Lock()

	delete(s.connections, conn.RemoteAddr().String())
	err := conn.Close()
	if err != nil {
		s.l.Debug("error while closing connection", zap.String("conn", conn.RemoteAddr().String()))
	}
}

func (s *Server) closeSocket(conn net.Conn, msg string) {
	_ = wsutil.WriteServerMessage(conn, ws.OpClose, []byte(msg))
	s.delConn(conn)
}
