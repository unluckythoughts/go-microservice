package web

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type Client interface {
	Send(method, url string, body []byte, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	GetResponse(url string, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	PostResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	PutResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	PatchResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	DeleteResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error)
}

type client struct {
	baseURL    string
	httpClient *http.Client
	headers    http.Header
}

var (
	emptyBody = []byte{}
)

func NewClient(url string, defaultHeaders ...http.Header) Client {
	c := &client{
		baseURL:    url,
		httpClient: &http.Client{},
	}

	if len(defaultHeaders) > 0 {
		c.headers = defaultHeaders[0]
	}

	return c
}

func (c *client) Send(
	method, url string, body []byte,
	resp interface{}, reqHeaders ...http.Header,
) (status int, err error) {
	reqBody := new(bytes.Buffer)
	_, err = reqBody.Write(body)
	if err != nil {
		return 0, errors.Wrap(err, "could not parse the request body")
	}

	url = c.baseURL + url
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, errors.Wrap(err, "could not create http request")
	}

	headers := append(reqHeaders, c.headers)
	for _, header := range headers {
		if header != nil {
			for key, values := range header {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
		}
	}

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "could not send request")
	}

	if resp != nil {
		data, err := ioutil.ReadAll(httpResp.Body)
		if err != nil {
			return 0, errors.Wrap(err, "could not read response body")
		}

		err = json.Unmarshal(data, resp)
		if err != nil {
			return 0, errors.Wrapf(err, "could not unmarshal response body: %s", string(data))
		}
	}

	return httpResp.StatusCode, nil
}

func (c *client) GetResponse(url string, resp interface{}, reqHeaders ...http.Header) (status int, err error) {
	return c.Send(http.MethodGet, url, emptyBody, resp, reqHeaders...)
}

func getRequestBody(body interface{}) (reqBody []byte, err error) {
	reqBody = emptyBody
	if body != nil {
		reqBody, err = json.Marshal(body)
	}

	return reqBody, err
}

func (c *client) PostResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error) {
	reqBody, err := getRequestBody(body)
	if err != nil {
		return 0, errors.Wrapf(err, "could not marshal request body: %+v", body)
	}

	return c.Send(http.MethodPost, url, reqBody, resp, reqHeaders...)
}

func (c *client) PutResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error) {
	reqBody, err := getRequestBody(body)
	if err != nil {
		return 0, errors.Wrapf(err, "could not marshal request body: %+v", body)
	}

	return c.Send(http.MethodPut, url, reqBody, resp, reqHeaders...)
}

func (c *client) PatchResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error) {
	reqBody, err := getRequestBody(body)
	if err != nil {
		return 0, errors.Wrapf(err, "could not marshal request body: %+v", body)
	}

	return c.Send(http.MethodPatch, url, reqBody, resp, reqHeaders...)
}

func (c *client) DeleteResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error) {
	reqBody, err := getRequestBody(body)
	if err != nil {
		return 0, errors.Wrapf(err, "could not marshal request body: %+v", body)
	}

	return c.Send(http.MethodDelete, url, reqBody, resp, reqHeaders...)
}
