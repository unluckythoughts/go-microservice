package web

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	e := NewError(http.StatusBadRequest, errors.New("bad input"))
	assert.Equal(t, "bad input", e.Error())
	assert.Equal(t, http.StatusBadRequest, e.code)
}

func TestBadRequestDefault(t *testing.T) {
	e := BadRequest()
	assert.Equal(t, http.StatusBadRequest, e.code)
	assert.NotEmpty(t, e.Error())
}

func TestBadRequestWithError(t *testing.T) {
	e := BadRequest(errors.New("field required"))
	assert.Equal(t, http.StatusBadRequest, e.code)
	assert.Equal(t, "field required", e.Error())
}

func TestBadRequestWithNilError(t *testing.T) {
	e := BadRequest(nil)
	assert.Equal(t, http.StatusBadRequest, e.code)
	assert.NotEmpty(t, e.Error())
}

func TestInternalServerErrorDefault(t *testing.T) {
	e := InternalServerError()
	assert.Equal(t, http.StatusInternalServerError, e.code)
	assert.NotEmpty(t, e.Error())
}

func TestInternalServerErrorWithError(t *testing.T) {
	e := InternalServerError(errors.New("db failure"))
	assert.Equal(t, http.StatusInternalServerError, e.code)
	assert.Equal(t, "db failure", e.Error())
}

func TestInternalServerErrorWithNilError(t *testing.T) {
	e := InternalServerError(nil)
	assert.Equal(t, http.StatusInternalServerError, e.code)
}

func TestNotFoundDefault(t *testing.T) {
	e := NotFound()
	assert.Equal(t, http.StatusNotFound, e.code)
	assert.NotEmpty(t, e.Error())
}

func TestNotFoundWithError(t *testing.T) {
	e := NotFound(errors.New("user not found"))
	assert.Equal(t, http.StatusNotFound, e.code)
	assert.Equal(t, "user not found", e.Error())
}

func TestNotFoundWithNilError(t *testing.T) {
	e := NotFound(nil)
	assert.Equal(t, http.StatusNotFound, e.code)
}
