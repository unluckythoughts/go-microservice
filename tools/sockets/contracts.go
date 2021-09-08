package sockets

import (
	"net"
	"time"

	"go.uber.org/zap"
)

type RequestBody struct {
	ID     string   `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type Request struct {
	Conn      net.Conn
	ID        string
	Logger    *zap.Logger
	Body      RequestBody
	Timestamp time.Time `json:"-"`
}

type Response struct {
	ID        string      `json:"id,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
	Success   bool        `json:"success"`
	Method    string      `json:"method,omitempty"`
	Error     string      `json:"error,omitempty"`
	Result    interface{} `json:"result,omitempty"`
}
