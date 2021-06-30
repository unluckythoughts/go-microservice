package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type reqBody struct {
	read bool
	raw  []byte
}

// request struct of the request sent to the shs handlers
type request struct {
	_int        *http.Request
	id          string
	routeParams *httprouter.Params
	log         *zap.Logger
	startTime   time.Time
	body        reqBody
	context     map[string]interface{}
	config      interface{}
}

func (r *router) newRequest(req *http.Request, p httprouter.Params) *request {
	reqID := uuid.Must(uuid.NewV4()).String()
	reqLogger := r.l.With(zap.String("reqId", reqID))

	return &request{
		_int:        req,
		id:          reqID,
		routeParams: &p,
		log:         reqLogger,
		startTime:   time.Now(),
		body:        reqBody{},
	}
}

// GetValidatedBody validates the body and updates the ptr reference, errors if any issues
func (r *request) GetValidatedBody(ptr interface{}) (err error) {
	data := r.body.raw
	if !r.body.read {
		if data, err = ioutil.ReadAll(r._int.Body); err != nil {
			return err
		}
		r.body.read = true
		r.body.raw = data
	}

	if len(data) < 1 {
		return nil
	}

	if err = json.Unmarshal(data, ptr); err != nil {
		return err
	} else if _, err = govalidator.ValidateStruct(ptr); err != nil {
		return err
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

// GetContext returns the request context
func (r *request) GetContext() map[string]interface{} {
	return r.context
}

// GetContextValue returns the value for the given key in its context
func (r *request) GetContextValue(key string) interface{} {
	return r.context[key]
}

// SetContextValue sets the value for the given key in its context, errors if the key already exists
func (r *request) SetContextValue(key string, value interface{}) error {
	if _, ok := r.context[key]; ok {
		return errors.New("key already exists in context")
	}

	r.context[key] = value
	return nil
}

// Log returns the logger interface for this request
func (r *request) Logger() *zap.Logger {
	return r.log
}

// Log returns the logger interface for this request
func (r *request) SetLogger(l *zap.Logger) {
	r.log = l
}
