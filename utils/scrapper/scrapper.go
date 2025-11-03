package scrapper

import (
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/unluckythoughts/go-microservice/tools/logger"
	"github.com/unluckythoughts/go-microservice/tools/web"
)

// ScrapperConfig holds configuration for the web scrapper
type ScrapperConfig struct {
	// UserAgent to use for requests
	UserAgent string
	// Delay between requests
	Delay time.Duration
	// BaseURL string
	BaseURL string
	// Maximum number of threads
	Parallelism int
	// Request timeout
	Timeout time.Duration
	// Debug mode
	Debug bool
	// Headers to add to requests
	Headers map[string]string
	// AllowURLRevisit allows revisiting the same URL multiple times
	AllowURLRevisit bool
	// RoundTripper for custom HTTP transport
	RoundTripper http.RoundTripper
}

// ScrapedData represents the data scraped from a web page
type ScrapedData struct {
	URL          string `json:"url"`
	HTML         string `json:"html"`
	StatusCode   int    `json:"status_code"`
	ErrorMessage string `json:"error,omitempty"`
}

// Scrapper handles web scraping operations
type Scrapper struct {
	config *ScrapperConfig
	ctx    web.Context
}

// NewScrapper creates a new scrapper instance
func NewScrapper(cfg ScrapperConfig, ctx web.Context) *Scrapper {
	if cfg.BaseURL == "" {
		panic("BaseURL must be provided")
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	}

	if cfg.Parallelism <= 0 {
		cfg.Parallelism = 20
	}

	if cfg.Delay <= 0 {
		cfg.Delay = 1 * time.Second
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	if ctx == nil {
		// Create a default context with no-op logger
		ctx = web.NewContext(logger.New(logger.Options{}))
	}

	s := &Scrapper{
		config: &cfg,
		ctx:    ctx,
	}

	if s.ValidateURL(s.config.BaseURL) != nil {
		panic("BaseURL '" + s.config.BaseURL + "' is not accessible")
	}

	return s
}

// setupHandlers sets up the event handlers for the colly collector
func (s *Scrapper) setupHandlers(c *colly.Collector, data *ScrapedData) {
	// Handle responses
	c.OnResponse(func(r *colly.Response) {
		data.StatusCode = r.StatusCode
		data.HTML = string(r.Body)
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		data.StatusCode = r.StatusCode
		data.ErrorMessage = err.Error()
	})
}

// ScrapeURL scrapes a single URL and returns the extracted data
func (s *Scrapper) ScrapeURL(targetURL string, col *colly.Collector) (*ScrapedData, error) {
	targetURL = s.GetFullURL(targetURL)
	s.ctx.Logger().Infow("Starting to scrape URL", "url", targetURL)

	data := &ScrapedData{
		URL:        targetURL,
		HTML:       "",
		StatusCode: 0,
	}

	// Create a fresh collector for single URL to avoid conflicts with batch operations
	if col == nil {
		col = s.createCollector()
	}

	// Set up handlers
	s.setupHandlers(col, data)

	// Set up retry mechanism for 429 status codes
	maxRetries := 5

	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Visit the URL
		err = col.Visit(targetURL)
		if err == nil || !strings.Contains(err.Error(), "429") {
			s.ctx.Logger().Infow("Successfully scraped URL", "url", targetURL)
			return data, nil
		}

		if attempt == maxRetries {
			break
		}

		// Calculate exponential backoff delay
		delay := time.Duration(1<<uint(attempt)) * s.config.Delay
		s.ctx.Logger().Warnw("Rate limited, retrying after delay",
			"url", targetURL,
			"attempt", attempt+1,
			"delay", delay,
			"error", err)

		time.Sleep(delay)
	}

	return data, err
}

// ScrapeURLs scrapes multiple URLs concurrently and returns change of results
func (s *Scrapper) ScrapeURLs(urls []string) <-chan *ScrapedData {
	resultChan := make(chan *ScrapedData, len(urls))

	// Create a collector for batch operations
	col := s.createCollector()

	go func() {
		defer close(resultChan)

		for _, url := range urls {
			scraped, err := s.ScrapeURL(url, col)
			if err != nil {
				s.ctx.Logger().Errorw("Failed to scrape URL", "url", url, "error", err)
				scraped.ErrorMessage = err.Error()
			}

			resultChan <- scraped
		}
	}()

	return resultChan
}
