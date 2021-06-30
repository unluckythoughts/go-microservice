package web

import (
	"net/http"

	"github.com/gobwas/ws"
	"github.com/investing-bot/microservice/tools/sockets"
)

func (s *Server) upgradeConnection(
	w http.ResponseWriter,
	req *http.Request,
) {
	s.logger.Info("got a socket request!")
	conn, _, _, err := ws.UpgradeHTTP(req, w)
	if err != nil {
		s.logger.Error("Could not upgrade connection")
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("Could not upgrade connection"))
		if err != nil {
			s.logger.Error("could not write error to http response")
		}

		return
	}

	go s.socketServer.HandleSocketConnection(conn)
}

func (s *Server) AddSocketHandler(method string, handler sockets.Handler) {
	s.socketServer.AddHandler(method, handler)
}
