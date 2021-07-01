package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type httpResponse struct {
	Success   bool        `json:"success"`
	RequestID string      `json:"requestId"`
	Error     string      `json:"error,omitempty"`
	Result    interface{} `json:"result,omitempty"`
}

type responseStatus string

const (
	okStatus    responseStatus = "ok"
	errorStatus responseStatus = "error"
)

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

func timeElapsed(st time.Time) (string, float64) {
	d := time.Since(st)
	floatD := float64(d.Nanoseconds()) / math.Pow10(6)
	return d.String(), floatD
}

func logResponse(req *request, statusCode int, respBody *bytes.Buffer) {
	method := req._int.Method
	url := req._int.URL.String()
	strDuration, duration := timeElapsed(req.startTime)
	reqBody := map[string]interface{}{}
	req.GetValidatedBody(&reqBody)

	msg := fmt.Sprintf("%d %s %s %s", statusCode, strings.ToUpper(method), url, strDuration)
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("url", url),
		zap.Float64("duration", duration),
		zap.Any("request.body", reqBody),
		zap.String("response.body", respBody.String()),
	}

	urlParams := req._int.URL.Query()
	if len(urlParams) > 0 {
		fields = append(fields, zap.Any("request.urlparams", urlParams))
	}

	if len(*req.routeParams) > 0 {
		fields = append(fields, zap.Any("request.routeparams", req.routeParams))
	}

	req.log.With(fields...).Debug(msg)
}

// sendResponse function to send response to http requests
func sendResponse(resp *response, data interface{}, respErr error, statusCode int) {
	base := httpResponse{
		Success:   true,
		RequestID: resp.request.id,
		Error:     InternalServerError().Error(),
	}

	webError := &httpError{}
	resp.respWriter.Header().Set("Content-Type", "application/json")
	if respErr == nil {
		statusCode = 200
		base.Error = ""
		base.Result = data
	} else if errors.As(respErr, webError) {
		statusCode = webError.code
		base.Error = webError.message
		base.Result = nil
	} else {
		base.Error = respErr.Error()
	}

	resp.respWriter.WriteHeader(statusCode)
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(base)
	if err != nil {
		resp.respWriter.WriteHeader(500)
		base.Success = false
		base.Error = "error while parsing response body"
		base.Result = nil
		json.NewEncoder(body).Encode(base)
	}

	fmt.Fprint(resp.respWriter, body)
	logResponse(resp.request, statusCode, body)
}
