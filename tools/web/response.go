package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

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

func logResponse(resp *response, statusCode int, respBody *bytes.Buffer) {
	method := resp.request._int.Method
	url := resp.request._int.URL.String()
	strDuration, duration := timeElapsed(resp.request.startTime)
	reqBody := map[string]interface{}{}
	resp.request.GetValidatedBody(&reqBody)

	msg := fmt.Sprintf("%d %s %s %s", statusCode, strings.ToUpper(method), url, strDuration)
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("url", url),
		zap.Float64("duration", duration),
		zap.Any("request.routeparams", resp.request.routeParams),
		zap.Any("request.body", reqBody),
		zap.String("response.body", respBody.String()),
	}

	resp.request.log.With(fields...).Debug(msg)
}

// sendResponse function to send response to http requests
func sendResponse(resp *response, data interface{}, respErr error, statusCode int) {
	base := HTTPResponse{
		Success:   true,
		RequestID: resp.request.id,
		Error:     "Internal Error",
	}

	resp.respWriter.Header().Set("Content-Type", "application/json")
	if respErr == nil {
		statusCode = 200
		base.Error = ""
		base.Result = data
	}

	resp.respWriter.WriteHeader(statusCode)
	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(base)
	if err != nil {
		resp.respWriter.WriteHeader(500)
		base.Success = false
		base.Error = "Internal Error"
		base.Result = nil
		json.NewEncoder(body).Encode(base)
	}

	fmt.Fprint(resp.respWriter, body)
	logResponse(resp, statusCode, body)
}
