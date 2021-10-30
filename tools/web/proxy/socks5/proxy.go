package socks5

import (
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
)

const (
	PROXY_HOST = "159.65.47.58:50001"
)

func NewProxyClient(proxyHost string) (*http.Client, error) {
	if proxyHost != "" {
		return nil, errors.New("proxy host is mandatory")
	}

	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	proxyDialer, err := proxy.SOCKS5("tcp", proxyHost, nil, baseDialer)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create socks proxy with host %s", proxyHost)
	}

	return &http.Client{
		Transport: &http.Transport{
			Dial:                  proxyDialer.Dial,
			MaxIdleConns:          10,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}, nil
}
