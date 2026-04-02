package auth

import (
	"net/http"

	"github.com/unluckythoughts/go-microservice/v2/tools/web"
)

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
	var resp LoginResponse
	status, err := cl.c.PostResponse("/auth/login", req, &resp)
	return resp, status, err
}

func (cl *client) Register(req RegisterRequest) (string, int, error) {
	var message string
	status, err := cl.c.PostResponse("/auth/register", req, &message)
	return message, status, err
}

func (cl *client) Logout() (string, int, error) {
	var message string
	status, err := cl.c.GetResponse("/auth/logout", &message)
	return message, status, err
}

func (cl *client) ResetPassword(target, targetType string) (string, int, error) {
	url := "/auth/reset-password/" + target + "?type=" + targetType
	var message string
	status, err := cl.c.PatchResponse(url, nil, &message)
	return message, status, err
}

func (cl *client) UpdatePassword(req UpdatePasswordRequest) (string, int, error) {
	var message string
	status, err := cl.c.PostResponse("/auth/update-password", req, &message)
	return message, status, err
}

func (cl *client) ChangePassword(req ChangePasswordRequest) (string, int, error) {
	var message string
	status, err := cl.c.PostResponse("/auth/change-password", req, &message)
	return message, status, err
}

func (cl *client) VerifyToken(target, token string) (bool, int, error) {
	url := "/auth/verify/" + target + "/" + token
	var resp bool
	status, err := cl.c.GetResponse(url, &resp)
	return resp, status, err
}

func (cl *client) GetUser() (LoginResponse, int, error) {
	var resp LoginResponse
	status, err := cl.c.GetResponse("/auth/user", &resp)
	return resp, status, err
}

func (cl *client) UpdateUser(req UpdateUserRequest) (string, int, error) {
	var message string
	status, err := cl.c.PutResponse("/auth/user", req, &message)
	return message, status, err
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
