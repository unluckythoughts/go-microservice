package client

import (
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
)

type Client struct {
	auth.ClientWithAuth
}

func NewClient(baseURL string) *Client {
	c := auth.NewClientWithAuth(baseURL)

	return &Client{c}
}
