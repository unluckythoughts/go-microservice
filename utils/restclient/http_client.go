package restclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type RESTClient struct {
	client  http.Client
	baseURL string
}

func NewRESTClient(url string) *RESTClient {
	return &RESTClient{
		client:  http.Client{},
		baseURL: url,
	}
}

func getRequestBody(body interface{}) (reader io.Reader, err error) {
	if strBody, ok := body.(string); ok {
		return strings.NewReader(strBody), nil
	}

	data := []byte{}
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	return bytes.NewReader(data), nil
}

func parseBody(req *http.Request, resp *http.Response, result interface{}) error {
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "could not read GET %s request body", req.RequestURI)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("GET %s response code: %d. body: %s", req.RequestURI, resp.StatusCode, string(bytes))
	}

	if result != nil {
		if err := json.Unmarshal(bytes, result); err != nil {
			return errors.Wrapf(err, "could not parse body %s to type %T", string(bytes), result)
		}
	}

	return nil
}

func (c *RESTClient) send(req *http.Request, result interface{}) error {
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "could not send %s %s", req.Method, req.RequestURI)
	}

	return parseBody(req, resp, result)
}

func (c *RESTClient) GET(path string, result interface{}) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return errors.Wrapf(err, "could not create GET request for path %s", path)
	}

	return c.send(req, result)
}

func (c *RESTClient) POST(path string, body interface{}, result interface{}) error {
	reqBody, err := getRequestBody(body)
	if err != nil {
		return errors.Wrapf(err, "could not json marshal body %+v", body)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, reqBody)
	if err != nil {
		return errors.Wrapf(err, "could not create GET request for path %s", path)
	}

	return c.send(req, result)
}

func (c *RESTClient) Send(method string, path string, body interface{}, result interface{}) error {
	reqBody, err := getRequestBody(body)
	if err != nil {
		return errors.Wrapf(err, "could not json marshal body %+v", body)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return errors.Wrapf(err, "could not create GET request for path %s", path)
	}

	return c.send(req, result)
}
