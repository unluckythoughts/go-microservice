package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	localcontext "github.com/unluckythoughts/go-microservice/v2/tools/context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type reqBody struct {
	read bool
	raw  []byte
}

// request struct of the request sent to the shs handlers
type request struct {
	_int        *http.Request
	routeParams *httprouter.Params
	id          string
	body        reqBody
	ctx         localcontext.Context
}

// newRequest creates a new request object with the given http request and httprouter params
func (r *router) newRequest(req *http.Request, p httprouter.Params) *request {
	reqID := uuid.Must(uuid.NewV4()).String()
	l := r.l.With(zap.String("id", reqID))

	return &request{
		_int:        req,
		routeParams: &p,
		id:          reqID,
		body:        reqBody{},
		ctx:         localcontext.NewContext(l),
	}
}

// timeElapsed returns the time elapsed since the request was created and the time in milliseconds
func (r *request) timeElapsed() (string, float64) {
	d := time.Since(r.ctx.StartTime())
	floatD := float64(d.Nanoseconds()) / math.Pow10(6)
	return d.String(), floatD
}

// With adds the given fields to the request context logger
func (r *request) With(fields ...zapcore.Field) {
	r.ctx.WithFields(fields...)
}

// SetContextValue sets a value in the request context with the given key
// returns an error if the key already exists in the context
func (r *request) SetContextValue(cKey string, cValue any) error {
	if r.ctx.Value(cKey) != nil {
		return errors.New("key already exists in context")
	}

	// nolint:staticcheck
	r.ctx.WithValue(cKey, cValue)
	return nil
}

// GetContext returns the request context
func (r *request) GetContext() localcontext.Context {
	return r.ctx
}

// GetMethod returns the HTTP method of the request
func (r *request) GetMethod() string {
	return r._int.Method
}

// GetPath returns the request path
// This is the path without the query parameters
// e.g. /api/v1/users/123
func (r *request) GetPath() string {
	return r._int.URL.Path
}

// GetValidatedBody validates the body and updates the ptr reference, errors if any issues
func (r *request) GetValidatedBody(ptr any) (err error) {
	data := r.body.raw
	if !r.body.read {
		if data, err = io.ReadAll(r._int.Body); err != nil {
			return NewError(http.StatusBadRequest, fmt.Errorf("failed to read request body: %w", err))
		}
		r.body.read = true
		r.body.raw = data
	}

	if len(data) < 1 {
		return nil
	}

	if err = json.Unmarshal(data, ptr); err != nil {
		return NewError(http.StatusBadRequest, fmt.Errorf("failed to read request body: %w", err))
	} else if _, err = govalidator.ValidateStruct(ptr); err != nil {
		return NewError(http.StatusBadRequest, fmt.Errorf("failed to validate request body: %w", err))
	}

	return nil
}

// GetHeaders returns http header object the all headers for the request
func (r *request) GetHeaders() http.Header {
	return r._int.Header
}

// GetHeader returns the header value with the given key
func (r *request) GetHeader(key string) string {
	return r._int.Header.Get(key)
}

// GetURLParam returns the URL param value with the given key
func (r *request) GetURLParam(key string) string {
	return r._int.URL.Query().Get(key)
}

// GetRouteParam returns the route param value with the given key
func (r *request) GetRouteParam(key string) string {
	return r.routeParams.ByName(key)
}

// GetRemoteAddr returns the remote address of the client (IP only, port stripped)
func (r *request) GetRemoteAddr() string {
	addr := r._int.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i > 0 {
		return addr[:i]
	}
	return addr
}

func (r *request) GetInternalRequest() *http.Request {
	return r._int
}
