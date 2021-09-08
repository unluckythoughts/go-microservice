package sockets

import (
	"encoding/json"
	"net"
	"time"

	"github.com/gobwas/ws/wsutil"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

var (
	reqchan = make(chan Request, 2000)
)

func (s *Server) StartSocketWorkers() {
	for i := 0; i < s.workerCount; i++ {
		go func(reqChan <-chan Request) {
			for req := range reqChan {
				s.handleSocketRequest(req)
			}
		}(reqchan)
	}
}

func (s *Server) HandleSocketConnection(conn net.Conn) {
	s.addConn(conn)
	for {
		bytes, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			s.l.Error(
				"could not read data from connection ",
				zap.String("data", string(bytes)),
				zap.Error(err),
			)
			s.closeSocket(conn, "could not read data from connection")
			return
		}

		body := RequestBody{}
		if err = json.Unmarshal(bytes, &body); err != nil {
			s.l.Error("could not parse the request body")
			s.closeSocket(conn, "could not parse request body")
			return
		}

		reqID := uuid.Must(uuid.NewV4()).String()
		reqchan <- Request{
			Conn:      conn,
			ID:        reqID,
			Logger:    s.l.With(zap.String("reqId", reqID)),
			Body:      body,
			Timestamp: time.Now(),
		}
	}
}
