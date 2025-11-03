package scrapper

import (
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/pkg/errors"
)

func hasAttribute(sel string) bool {
	return strings.ContainsAny(sel, "[]")
}

func getAttributeName(sel string) string {
	start := strings.LastIndex(sel, "[")
	end := strings.LastIndex(sel, "]")
	if start == -1 || end == -1 || start > end {
		return ""
	}
	return sel[start+1 : end]
}

func (s *Scrapper) ValidateURL(targetURL string) error {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return errors.Wrap(err, "invalid URL")
	}

	col := s.createCollector()

	var respError error
	col.OnError(func(r *colly.Response, err error) {
		respError = errors.Wrap(err, "error occurred during HEAD request")
	})

	col.OnResponse(func(r *colly.Response) {
		if r.StatusCode < 200 || r.StatusCode >= 300 {
			respError = errors.Errorf("URL returned status code %d", r.StatusCode)
		}
	})

	err = col.Head(parsedURL.String())
	if err != nil {
		return errors.Wrap(err, "URL is not accessible")
	}

	if respError != nil {
		return errors.Wrap(respError, "URL is not accessible")
	}

	return nil
}

// createCollector creates a new colly collector with the configured settings
func (s *Scrapper) createCollector() *colly.Collector {
	var c *colly.Collector

	if s.config.Debug {
		c = colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))
	} else {
		c = colly.NewCollector()
	}

	if s.config.AllowURLRevisit {
		c.AllowURLRevisit = true
	}

	if s.config.RoundTripper != nil {
		c.WithTransport(s.config.RoundTripper)
	}

	// Configure limits
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: s.config.Parallelism,
		Delay:       s.config.Delay,
	})

	// Set user agent
	c.UserAgent = s.config.UserAgent

	// Set headers
	c.OnRequest(func(r *colly.Request) {
		for key, value := range s.config.Headers {
			r.Headers.Set(key, value)
		}
	})

	// Set timeout
	c.SetRequestTimeout(s.config.Timeout)

	return c
}

// GetFullURL constructs the full URL based on the base URL and the provided path
func (s *Scrapper) GetFullURL(url string) string {
	if strings.HasPrefix(url, "/") {
		return s.config.BaseURL + url
	}

	return url
}

// StripBaseURL removes the base URL from the provided URL path
func (s *Scrapper) StripBaseURL(url string) string {
	if strings.HasPrefix(url, s.config.BaseURL) {
		url = strings.TrimPrefix(url, s.config.BaseURL)
	}

	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}

	return url
}
