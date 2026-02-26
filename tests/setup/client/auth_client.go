package client

import (
	"fmt"
	"net/url"

	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
	"github.com/unluckythoughts/go-microservice/v2/tools/web"
)

func (c *Client) Login(req auth.Credentials) (auth.LoginResponse, int, error) {
	base := web.HTTPResponse{}
	status, err := c.client.PostResponse("/auth/login", req, &base)
	if err = parseOnlyResponse(status, err, base); err != nil {
		return auth.LoginResponse{}, status, err
	}

	data, err := parseData[auth.LoginResponse](base)
	return data, status, err
}

func (c *Client) Register(req auth.RegisterRequest) (string, int, error) {
	base := web.HTTPResponse{}
	status, err := c.client.PostResponse("/auth/register", req, &base)
	if err = parseOnlyResponse(status, err, base); err != nil {
		return "", status, err
	}

	data, err := parseData[string](base)
	return data, status, err
}

func (c *Client) Logout() (string, int, error) {
	base := web.HTTPResponse{}
	status, err := c.client.GetResponse("/auth/logout", &base)
	if err = parseOnlyResponse(status, err, base); err != nil {
		return "", status, err
	}

	data, err := parseData[string](base)
	return data, status, err
}

func (c *Client) ResetPassword(target, targetType string) (string, int, error) {
	escapedTarget := url.PathEscape(target)
	path := fmt.Sprintf("/auth/reset-password/%s", escapedTarget)
	if targetType != "" {
		path += "?type=" + url.QueryEscape(targetType)
	}

	base := web.HTTPResponse{}
	status, err := c.client.PatchResponse(path, nil, &base)
	if err = parseOnlyResponse(status, err, base); err != nil {
		return "", status, err
	}

	data, err := parseData[string](base)
	return data, status, err
}

func (c *Client) UpdatePassword(req auth.UpdatePasswordRequest) (string, int, error) {
	base := web.HTTPResponse{}
	status, err := c.client.PutResponse("/auth/update-password", req, &base)
	if err = parseOnlyResponse(status, err, base); err != nil {
		return "", status, err
	}

	data, err := parseData[string](base)
	return data, status, err
}

func (c *Client) ChangePassword(req auth.ChangePasswordRequest) (string, int, error) {
	base := web.HTTPResponse{}
	status, err := c.client.PutResponse("/auth/change-password", req, &base)
	if err = parseOnlyResponse(status, err, base); err != nil {
		return "", status, err
	}

	data, err := parseData[string](base)
	return data, status, err
}

func (c *Client) VerifyToken(target, token string) (bool, int, error) {
	path := fmt.Sprintf("/auth/verify/%s/%s", url.PathEscape(target), url.PathEscape(token))
	base := web.HTTPResponse{}
	status, err := c.client.GetResponse(path, &base)
	if err = parseOnlyResponse(status, err, base); err != nil {
		return false, status, err
	}

	data, err := parseData[bool](base)
	return data, status, err
}

func (c *Client) GetUser() (auth.LoginResponse, int, error) {
	base := web.HTTPResponse{}
	status, err := c.client.GetResponse("/auth/user", &base)
	if err = parseOnlyResponse(status, err, base); err != nil {
		return auth.LoginResponse{}, status, err
	}

	data, err := parseData[auth.LoginResponse](base)
	return data, status, err
}

func (c *Client) UpdateUser(req auth.UpdateUserRequest) (string, int, error) {
	base := web.HTTPResponse{}
	status, err := c.client.PutResponse("/auth/user", req, &base)
	if err = parseOnlyResponse(status, err, base); err != nil {
		return "", status, err
	}

	data, err := parseData[string](base)
	return data, status, err
}
