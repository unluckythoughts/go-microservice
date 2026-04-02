package auth

import (
	"net/http"

	"github.com/unluckythoughts/go-microservice/v2/tools/web"
)

// extractData unmarshals the Data field of an HTTPResponse into T.
// It propagates network/decoding errors but does NOT treat HTTP 4xx/5xx as Go
// errors — callers should inspect the returned status code instead.
func extractData[T any](base web.HTTPResponse, status int, err error) (T, int, error) {
	var zero T
	if err != nil {
		return zero, status, err
	}
	if !base.Ok || base.Data == nil {
		return zero, status, nil
	}
	var result T
	if marshalErr := web.MarshalData(base.Data, &result); marshalErr != nil {
		return zero, status, marshalErr
	}
	return result, status, nil
}

type client struct {
	c web.Client
}

type ClientWithAuth interface {
	web.Client
	Login(req Credentials) (LoginResponse, int, error)
	Register(req RegisterRequest) (string, int, error)
	Logout() (string, int, error)
	ResetPassword(target, targetType string) (string, int, error)
	UpdatePassword(req UpdatePasswordRequest) (string, int, error)
	ChangePassword(req ChangePasswordRequest) (string, int, error)
	VerifyToken(target, token string) (bool, int, error)
	GetUser() (LoginResponse, int, error)
	UpdateUser(req UpdateUserRequest) (string, int, error)
}

func NewClientWithAuth(baseURL string, defaultHeaders ...http.Header) ClientWithAuth {
	return &client{
		c: web.NewClient(baseURL, defaultHeaders...),
	}
}

func AttachAuthMethods(c web.Client) ClientWithAuth {
	return &client{c: c}
}

func (cl *client) Login(req Credentials) (LoginResponse, int, error) {
	var base web.HTTPResponse
	status, err := cl.c.PostResponse("/auth/login", req, &base)
	return extractData[LoginResponse](base, status, err)
}

func (cl *client) Register(req RegisterRequest) (string, int, error) {
	var base web.HTTPResponse
	status, err := cl.c.PostResponse("/auth/register", req, &base)
	return extractData[string](base, status, err)
}

func (cl *client) Logout() (string, int, error) {
	var base web.HTTPResponse
	status, err := cl.c.GetResponse("/auth/logout", &base)
	return extractData[string](base, status, err)
}

func (cl *client) ResetPassword(target, targetType string) (string, int, error) {
	url := "/auth/reset-password/" + target + "?type=" + targetType
	var base web.HTTPResponse
	status, err := cl.c.PatchResponse(url, nil, &base)
	return extractData[string](base, status, err)
}

func (cl *client) UpdatePassword(req UpdatePasswordRequest) (string, int, error) {
	var base web.HTTPResponse
	status, err := cl.c.PutResponse("/auth/update-password", req, &base)
	return extractData[string](base, status, err)
}

func (cl *client) ChangePassword(req ChangePasswordRequest) (string, int, error) {
	var base web.HTTPResponse
	status, err := cl.c.PutResponse("/auth/change-password", req, &base)
	return extractData[string](base, status, err)
}

func (cl *client) VerifyToken(target, token string) (bool, int, error) {
	url := "/auth/verify/" + target + "/" + token
	var base web.HTTPResponse
	status, err := cl.c.GetResponse(url, &base)
	return extractData[bool](base, status, err)
}

func (cl *client) GetUser() (LoginResponse, int, error) {
	var base web.HTTPResponse
	status, err := cl.c.GetResponse("/auth/user", &base)
	return extractData[LoginResponse](base, status, err)
}

func (cl *client) UpdateUser(req UpdateUserRequest) (string, int, error) {
	var base web.HTTPResponse
	status, err := cl.c.PutResponse("/auth/user", req, &base)
	return extractData[string](base, status, err)
}

// Helper methods to satisfy the web.Client interface for the embedded client
func (cl *client) SetBearerToken(token string) {
	cl.c.SetBearerToken(token)
}

func (cl *client) ClearBearerToken() {
	cl.c.ClearBearerToken()
}

func (cl *client) Log(value string) {
	cl.c.Log(value)
}

func (cl *client) GetResponse(path string, response any, headers ...http.Header) (int, error) {
	return cl.c.GetResponse(path, response, headers...)
}

func (cl *client) PostResponse(path string, body any, response any, headers ...http.Header) (int, error) {
	return cl.c.PostResponse(path, body, response, headers...)
}

func (cl *client) PutResponse(path string, body any, response any, headers ...http.Header) (int, error) {
	return cl.c.PutResponse(path, body, response, headers...)
}

func (cl *client) PatchResponse(path string, body any, response any, headers ...http.Header) (int, error) {
	return cl.c.PatchResponse(path, body, response, headers...)
}

func (cl *client) DeleteResponse(path string, body any, response any, headers ...http.Header) (int, error) {
	return cl.c.DeleteResponse(path, body, response, headers...)
}

func (cl *client) Send(method, path string, body []byte, resp any, headers ...http.Header) (int, error) {
	return cl.c.Send(method, path, body, resp, headers...)
}
