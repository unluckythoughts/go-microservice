package httpproxy

import (
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/pkg/errors"
)

func NewProxyClient(proxyHost string) (*http.Client, error) {
	if proxyHost == "" {
		return nil, errors.New("proxy host is mandatory")
	}
	proxyUrl, err := url.Parse(proxyHost)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyURL(proxyUrl),
			MaxIdleConns:          10,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}, nil
}
