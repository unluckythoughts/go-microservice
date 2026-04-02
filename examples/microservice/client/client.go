package client

import (
	"net/http"

	"github.com/unluckythoughts/go-microservice/v2/tools/web"
)

type Client struct {
	client web.Client
}

func NewClient(baseURL string) *Client {
	return &Client{client: web.NewClient(baseURL, http.Header{})}
}

func (c *Client) SetBearerToken(token string) {
	c.client.SetBearerToken(token)
}

func (c *Client) ClearBearerToken() {
	c.client.ClearBearerToken()
}

func parseData[T any](status int, base web.HTTPResponse) (T, error) {
	var result T
	if err := web.HandleResponse(status, nil, base); err != nil {
		return result, err
	}

	if err := web.MarshalData(base.Data, &result); err != nil {
		return result, err
	}

	return result, nil
}
