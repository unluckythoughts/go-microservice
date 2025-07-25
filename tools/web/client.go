package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/unluckythoughts/go-microservice/tools/web/proxy/httpproxy"
	"github.com/unluckythoughts/go-microservice/tools/web/proxy/socks5"
)

type Client interface {
	Log(message string)
	SetBearerToken(token string)
	ClearBearerToken()

	// GetResponse, PostResponse, PutResponse, PatchResponse, DeleteResponse
	// are convenience methods for sending HTTP requests with the specified method and body.
	GetResponse(url string, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	PostResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	PutResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	PatchResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	DeleteResponse(url string, body interface{}, resp interface{}, reqHeaders ...http.Header) (status int, err error)
	// Send sends an HTTP request with the specified method, URL, body, and headers.
	// It returns the HTTP status code and an error if the request fails.
	Send(method, url string, body []byte, resp interface{}, reqHeaders ...http.Header) (status int, err error)
}

type client struct {
	BaseURL     string
	HTTPClient  *http.Client
	Headers     http.Header
	BearerToken string
}

var (
	emptyBody = []byte{}
)

func NewClient(baseURL string, defaultHeaders ...http.Header) Client {
	c := &client{
		BaseURL:     baseURL,
		HTTPClient:  &http.Client{},
		BearerToken: "",
	}

	if len(defaultHeaders) > 0 {
		c.Headers = defaultHeaders[0]
	}

	c.Headers.Add("Content-Type", "application/json")

	return c
}

func NewClientWithTransport(baseURL string, transport http.RoundTripper, defaultHeaders ...http.Header) Client {
	c := NewClient(baseURL, defaultHeaders...)
	c.(*client).HTTPClient.Transport = transport

	return c
}

func NewProxyClient(baseURL, proxyHost string, defaultHeaders ...http.Header) (Client, error) {
	httpClient, err := httpproxy.NewProxyClient(proxyHost)
	if err != nil {
		return nil, errors.Wrap(err, "could not create HTTP proxy client")
	}

	c := NewClient(baseURL, defaultHeaders...)
	c.(*client).HTTPClient = httpClient

	return c, nil
}

func NewSocks5ProxyClient(baseURL, proxyHost string, defaultHeaders ...http.Header) (Client, error) {
	httpClient, err := socks5.NewProxyClient(proxyHost)
	if err != nil {
		return nil, errors.Wrap(err, "could not create SOCKS5 proxy client")
	}

	c := NewClient(baseURL, defaultHeaders...)
	c.(*client).HTTPClient = httpClient

	return c, nil
}

// MarshalData marshals the response data into the provided interface
// (should be a pointer to a struct)
func MarshalData(data any, v any) error {
	if data == nil {
		return fmt.Errorf("no data found")
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, v); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// HandleResponse processes the web client response and checks for errors
// It returns an error if the status code indicates a failure or if the response contains an error.
// If the response is successful, it returns nil.
func HandleResponse(status int, err error, result any) error {
	if err != nil {
		return err
	}

	if status >= 400 {
		return fmt.Errorf("request failed with status %d", status)
	}

	// If result is a web.HTTPResponse, check for errors
	if baseResp, ok := result.(HTTPResponse); ok {
		if !baseResp.Ok {
			if baseResp.Error != "" {
				return fmt.Errorf("API error: %s", baseResp.Error)
			}
			return fmt.Errorf("API request failed")
		}
	}

	return nil
}

func (c *client) SetBearerToken(token string) {
	c.BearerToken = token
	if c.Headers == nil {
		c.Headers = http.Header{}
	}
	c.Headers.Add("Authorization", "Bearer "+token)
}

func (c *client) ClearBearerToken() {
	c.BearerToken = ""
	if c.Headers != nil {
		c.Headers.Del("Authorization")
	}
}

func (c *client) Log(message string) {
	message = url.PathEscape(message)
	url := fmt.Sprintf("/_log/**************%s**************", message)
	c.Send(http.MethodGet, url, nil, nil)
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

	url = c.BaseURL + url
	// nolint:noctx
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, errors.Wrap(err, "could not create http request")
	}

	headers := append(reqHeaders, c.Headers)
	for _, header := range headers {
		for key, values := range header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	httpResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "could not send request")
	}

	if resp != nil {
		defer func() { _ = httpResp.Body.Close() }()
		data, err := io.ReadAll(httpResp.Body)
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
