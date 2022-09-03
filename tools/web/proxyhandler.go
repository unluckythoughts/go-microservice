package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var passthruRequestHeaderKeys = [...]string{
	"Accept",
	"Accept-Encoding",
	"Accept-Language",
	"Cache-Control",
	"Cookie",
	"Referer",
	"User-Agent",
}

var passthruResponseHeaderKeys = [...]string{
	"Content-Encoding",
	"Content-Language",
	"Content-Type",
	"Cache-Control", // TODO: Is this valid in a response?
	"Date",
	"Etag",
	"Expires",
	"Last-Modified",
	"Location",
	"Server",
	"Vary",
}

func santizeProxyUrl(link *url.URL) *url.URL {
	path := link.Path

	path = strings.Replace(path, "/_proxy/", "", 1)

	newLink, err := url.Parse(path)
	if err != nil || newLink.Scheme == "" {
		newLink, err = url.Parse("http://" + path)
		if err != nil {
			return link
		}
	}

	return newLink
}

func proxyHandler(l *zap.Logger, rt http.RoundTripper) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		newUrl := santizeProxyUrl(r.URL)
		l.Debug(fmt.Sprintf("Proxing --> %v %v", r.Method, newUrl))

		// Construct filtered header to send to origin server
		hh := http.Header{}
		for _, hk := range passthruRequestHeaderKeys {
			if hk == "Referer" {
				hh.Set(hk, newUrl.Scheme+"://"+newUrl.Hostname())
			} else if hv, ok := r.Header[hk]; ok {
				hh[hk] = hv
			}
		}

		// Construct request to send to origin server
		rr := http.Request{
			Method:        r.Method,
			URL:           newUrl,
			Header:        hh,
			Body:          r.Body,
			ContentLength: r.ContentLength,
			Close:         r.Close,
		}

		// Forward request to origin server
		resp, err := rt.RoundTrip(&rr)
		if err != nil {
			msg := errors.Wrapf(err, "could not reach %s", rr.URL.String()).Error()
			http.Error(w, msg, resp.StatusCode)
			return
		}

		// Handle redirects
		if resp.StatusCode == http.StatusMovedPermanently ||
			resp.StatusCode == http.StatusPermanentRedirect ||
			resp.StatusCode == http.StatusTemporaryRedirect {
			rr.URL, _ = url.Parse(resp.Header.Get("Location"))
			resp, err = rt.RoundTrip(&rr)
			if err != nil {
				msg := errors.Wrapf(err, "could not reach %s", rr.URL.String()).Error()
				http.Error(w, msg, resp.StatusCode)
				return
			}
		}

		defer resp.Body.Close()

		// Transfer filtered header from origin server -> client
		respH := w.Header()
		for _, hk := range passthruResponseHeaderKeys {
			if hv, ok := resp.Header[hk]; ok {

				respH[hk] = hv
			}
		}
		w.WriteHeader(resp.StatusCode)

		if resp.StatusCode != http.StatusNotFound &&
			resp.StatusCode != http.StatusNoContent {
			// Transfer response from origin server -> client
			if resp.ContentLength > 0 {
				_, err = io.CopyN(w, resp.Body, resp.ContentLength)
			} else {
				_, err = io.Copy(w, resp.Body)
			}

			if err != nil {
				l.With(zap.Error(err)).Debug("error while reading the body of proxied response")
			}
		}
	}
}
