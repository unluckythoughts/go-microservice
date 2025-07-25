package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type HTTPResponse struct {
	Ok    bool        `json:"ok"`
	ID    string      `json:"id"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

// response struct for handlers to set response
type response struct {
	respWriter http.ResponseWriter
	request    *request
}

func newResponse(w http.ResponseWriter, r *request) *response {
	return &response{respWriter: w, request: r}
}

// NotImplemented to be developed function handler
func NotImplemented(req Request) (interface{}, error) {
	return nil, NewError(http.StatusMethodNotAllowed, errors.New("not implemented"))
}

// UnAuthorizedHandler 401 http handler function
func UnAuthorized(req Request) (interface{}, error) {
	return nil, NewError(http.StatusUnauthorized, errors.New("not authorized"))
}

// AddHeader adds the headers for response
func (r *response) AddHeader(key string, values ...string) {
	if len(values) == 0 {
		return
	}

	headers := r.respWriter.Header()
	for _, value := range values {
		headers.Add(key, value)
	}
}

func logResponse(req *request, statusCode int, respBody *bytes.Buffer, respErr error) {
	method := req._int.Method
	url := req._int.URL.String()

	resp := ""
	if respBody.Len() > 100 {
		resp = string(respBody.Bytes()[:45]) + "..." + string(respBody.Bytes()[respBody.Len()-45:])
	} else {
		resp = respBody.String()
	}

	strDuration, duration := req.timeElapsed()
	msg := fmt.Sprintf("%d %s %s %s", statusCode, strings.ToUpper(method), url, strDuration)
	fields := []zap.Field{
		zap.String("p", method+" "+url),
		zap.Float64("t", duration),
		zap.String("b", string(req.body.raw)),
		zap.String("r", resp),
	}

	urlParams := req._int.URL.Query()
	if len(urlParams) > 0 {
		fields = append(fields, zap.Any("request.urlparams", urlParams))
	}

	if len(*req.routeParams) > 0 {
		fields = append(fields, zap.Any("request.routeparams", req.routeParams))
	}

	if statusCode >= 500 {
		fields = append(fields, zap.Error(respErr))
	}

	req.ctx.l.With(fields...).Debug(msg)
}

// sendResponse function to send response to http requests
func sendResponse(resp *response, data interface{}, respErr error, statusCode int) {
	base := HTTPResponse{
		Ok:    true,
		ID:    resp.request.id,
		Error: InternalServerError().Error(),
	}

	webError := &httpError{}
	resp.respWriter.Header().Set("Content-Type", "application/json")
	if respErr == nil {
		statusCode = 200
		base.Error = ""
		base.Data = data
	} else if e, ok := respErr.(*httpError); ok {
		base.Ok = false
		statusCode = e.code
		base.Error = e.message
		base.Data = nil
	} else if errors.As(respErr, webError) {
		base.Ok = false
		statusCode = webError.code
		base.Error = webError.message
		base.Data = nil
	} else {
		base.Ok = false
		base.Error = respErr.Error()
	}

	resp.respWriter.WriteHeader(statusCode)
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(base)
	if err != nil {
		resp.respWriter.WriteHeader(500)
		base.Ok = false
		base.Error = "error while parsing response body"
		base.Data = nil
		_ = json.NewEncoder(body).Encode(base)
	}

	fmt.Fprint(resp.respWriter, body)
	logResponse(resp.request, statusCode, body, respErr)
}
