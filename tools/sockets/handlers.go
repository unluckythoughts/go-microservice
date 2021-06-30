package sockets

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"go.uber.org/zap"
)

type Handler func(req Request) (result interface{}, err error)

func (s *Server) sendResponse(req Request, data interface{}, errors ...error) {
	resp := Response{
		ID:        req.Body.ID,
		RequestID: req.ID,
		Method:    req.Body.Method,
	}

	if len(errors) > 0 && errors[0] != nil {
		resp.Success = false
		resp.Error = errors[0].Error()
	} else {
		resp.Success = true
		resp.Result = data
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		req.Logger.Error("could not marshall data", zap.Any("data", data), zap.Error(err))
	}

	err = wsutil.WriteServerMessage(req.Conn, ws.OpText, bytes)
	if err != nil {
		s.closeSocket(req.Conn, "could not write message to connection")
	}
	logSocketRequest(req, string(bytes))
}

func logSocketRequest(req Request, resp string) {
	req.Logger.Info(
		"Processed socket request",
		zap.String("duration", time.Since(req.Timestamp).String()),
		zap.String("id", req.Body.ID),
		zap.String("method", req.Body.Method),
		zap.Strings("params", req.Body.Params),
		zap.String("response", resp),
	)
}

func (s *Server) handleSocketRequest(req Request) {
	s.handlerMutex.RLock()

	handler, ok := s.handlers[req.Body.Method]
	if !ok {
		s.sendResponse(req, nil, errors.New("unknown method"))
		return
	}
	s.handlerMutex.RUnlock()

	data, err := handler(req)
	s.sendResponse(req, data, err)
}

func (s *Server) AddHandler(method string, handler Handler) {
	s.handlerMutex.Lock()
	s.handlers[method] = handler
	s.handlerMutex.Unlock()
}
