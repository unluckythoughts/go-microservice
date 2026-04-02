package client

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/unluckythoughts/go-microservice/v2/tools/web"
)

func (c *Client) Hello() (string, int, error) {
	base := web.HTTPResponse{}
	status, err := c.client.GetResponse("/hello", &base)
	data, err := parseData[string](status, base)
	return data, status, err
}

func (c *Client) Status() (int, error) {
	status, err := c.client.Send(http.MethodGet, "/_status", nil, nil)
	return status, err
}

func (c *Client) Log(message string) (int, error) {
	path := fmt.Sprintf("/_log/%s", url.PathEscape(message))
	status, err := c.client.Send(http.MethodGet, path, nil, nil)
	return status, err
}
