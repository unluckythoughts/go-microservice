package web

import (
	"net/http"

	"github.com/pkg/errors"
)

type httpError struct {
	err     error
	code    int
	message string
}

func (he httpError) Error() string {
	return he.message
}

func NewError(code int, err error) *httpError {
	return &httpError{code: code, err: err, message: err.Error()}
}

func BadRequest(err ...error) *httpError {
	if len(err) == 0 || err[0] == nil {
		return NewError(http.StatusBadRequest, errors.New("bad request"))
	}
	return NewError(http.StatusBadRequest, err[0])
}

func InternalServerError(err ...error) *httpError {
	if len(err) == 0 || err[0] == nil {
		return NewError(http.StatusInternalServerError, errors.New("unknown error"))
	}
	return NewError(http.StatusInternalServerError, err[0])
}

func NotFound(err ...error) *httpError {
	if len(err) == 0 || err[0] == nil {
		return NewError(http.StatusNotFound, errors.New("resource not found"))
	}
	return NewError(http.StatusNotFound, err[0])
}
