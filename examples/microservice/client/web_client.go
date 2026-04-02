package client

import (
	"github.com/unluckythoughts/go-microservice/v2/tools/web"
)

func (c *Client) Example() (string, int, error) {
	base := web.HTTPResponse{}
	status, err := c.GetResponse("/api/v1/example", &base)
	data, err := web.GetResponseData[string](status, base)
	return data, status, err
}
