package web

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
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
	ctx         *ctx
}

func (r *router) newRequest(req *http.Request, p httprouter.Params) *request {
	reqID := uuid.Must(uuid.NewV4()).String()
	l := r.l.With(zap.String("id", reqID))

	return &request{
		_int:        req,
		routeParams: &p,
		id:          reqID,
		body:        reqBody{},
		ctx:         newContext(l),
	}
}

func (r *request) timeElapsed() (string, float64) {
	d := time.Since(r.ctx.st)
	floatD := float64(d.Nanoseconds()) / math.Pow10(6)
	return d.String(), floatD
}

func (r *request) With(fields ...zapcore.Field) {
	r.ctx.l = r.ctx.l.With(fields...)
}

func (r *request) SetContextValue(key string, val interface{}) error {
	if r.ctx.Value(key) != nil {
		return errors.New("key already exists in context")
	}

	// nolint:staticcheck
	r.ctx.Context = context.WithValue(r.ctx, key, val)
	return nil
}

func (r *request) GetContext() Context {
	return r.ctx
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
