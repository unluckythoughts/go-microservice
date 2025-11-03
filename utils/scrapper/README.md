# Web Scrapper Utility

A lightweight web scraping utility built using [Colly v2](https://github.com/gocolly/colly) and [GoQuery](https://github.com/PuerkitoBio/goquery) for the Go microservice framework.

## Features

- **HTML Scraping**: Scrape raw HTML content from web pages
- **Content Extraction**: Extract specific content using CSS selectors
- **Batch Scraping**: Scrape multiple URLs concurrently with built-in rate limiting
- **URL Validation**: Validate and test URL accessibility before scraping
- **Retry Logic**: Automatic retry with exponential backoff for rate-limited requests (429 status)
- **Configurable**: Customize user agent, delays, timeouts, parallelism, and base URL
- **Error Handling**: Comprehensive error handling with detailed error information
- **Logging**: Integrated with the microservice framework's logging system
- **CSS Selector Support**: Extract content and attributes using flexible CSS selectors

## Installation

The scrapper is part of the go-microservice utilities. It uses Colly v2 for web scraping and GoQuery for HTML parsing. All required dependencies are included in the project's vendor directory.

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/unluckythoughts/go-microservice/tools/web"
    "github.com/unluckythoughts/go-microservice/tools/logger"
    "github.com/unluckythoughts/go-microservice/utils/scrapper"
)

func main() {
    // Create web context with logger
    ctx := web.NewContext(logger.New(logger.Options{}))

    // Configure scrapper
    config := scrapper.ScrapperConfig{
        BaseURL:     "https://example.com",
        UserAgent:   "MyBot 1.0",
        Delay:       2 * time.Second,
        Parallelism: 10,
        Timeout:     30 * time.Second,
    }

    // Create scrapper instance
    s := scrapper.NewScrapper(config, ctx)

    // Scrape a single URL (relative to BaseURL)
    data, err := s.ScrapeURL("/page", nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("URL: %s\n", data.URL)
    fmt.Printf("Status: %d\n", data.StatusCode)
    fmt.Printf("HTML Length: %d\n", len(data.HTML))
    if data.ErrorMessage != "" {
        fmt.Printf("Error: %s\n", data.ErrorMessage)
    }
}
```

### Content Extraction

Extract specific content from HTML using CSS selectors:

```go
// Extract text content from elements
titles, err := scrapper.ExtractContent(data.HTML, "h1")
if err != nil {
    log.Printf("Error extracting titles: %v", err)
} else {
    fmt.Printf("Page titles: %v\n", titles)
}

// Extract attributes from elements
links, err := scrapper.ExtractContent(data.HTML, "a[href]")
if err != nil {
    log.Printf("Error extracting links: %v", err)
} else {
    fmt.Printf("Links found: %v\n", links)
}

// Extract HTML content
htmlBlocks, err := scrapper.ExtractContent(data.HTML, "div.content[html]")
if err != nil {
    log.Printf("Error extracting HTML: %v", err)
} else {
    fmt.Printf("HTML blocks: %v\n", htmlBlocks)
}
```

### Batch Scraping

```go
// URLs relative to BaseURL
urls := []string{
    "/page1",
    "/page2", 
    "/api/data",
}

// Get results channel for concurrent scraping
resultChan := s.ScrapeURLs(urls)

// Process results as they come in
for result := range resultChan {
    if result.ErrorMessage == "" {
        fmt.Printf("Successfully scraped %s (Status: %d)\n", result.URL, result.StatusCode)
        
        // Extract specific content from each page
        titles, _ := scrapper.ExtractContent(result.HTML, "title")
        if len(titles) > 0 {
            fmt.Printf("Page title: %s\n", titles[0])
        }
    } else {
        fmt.Printf("Failed to scrape %s: %s\n", result.URL, result.ErrorMessage)
    }
}
```

### Advanced CSS Selectors

The scrapper supports various CSS selector patterns:

```go
// Text content extraction
paragraphs, _ := scrapper.ExtractContent(html, "p")              // All <p> text
headings, _ := scrapper.ExtractContent(html, "h1, h2, h3")      // Multiple selectors
nested, _ := scrapper.ExtractContent(html, "div.content span")   // Nested elements

// Attribute extraction
hrefs, _ := scrapper.ExtractContent(html, "a[href]")            // href attributes
images, _ := scrapper.ExtractContent(html, "img[src]")          // src attributes
classes, _ := scrapper.ExtractContent(html, "div[class]")       // class attributes

// HTML content extraction (preserves markup)
htmlContent, _ := scrapper.ExtractContent(html, "article[html]") // Inner HTML

// Complex selectors
specific, _ := scrapper.ExtractContent(html, "#main .content > p:first-child")
```

### URL Validation

```go
// Validate URL accessibility before scraping
targetURL := "/api/endpoint"
if err := s.ValidateURL(targetURL); err != nil {
    log.Printf("URL not accessible: %v", err)
} else {
    fmt.Println("URL is valid and accessible")
    // Proceed with scraping
    data, _ := s.ScrapeURL(targetURL, nil)
}
```

## Configuration Options

The `ScrapperConfig` struct supports the following options:

- **BaseURL** (required): Base URL for relative path resolution
- **UserAgent**: User agent string for requests (default: Chrome user agent)
- **Delay**: Delay between requests (default: 1 second)  
- **Parallelism**: Number of concurrent requests (default: 20)
- **Timeout**: Request timeout (default: 30 seconds)
- **Headers**: Additional HTTP headers to send with requests

**Note**: BaseURL is mandatory and must be accessible during scrapper initialization.

## Data Structures

### ScrapedData

The `ScrapedData` structure contains the results of a scraping operation:

```go
type ScrapedData struct {
    URL          string `json:"url"`           // The scraped URL (absolute)
    HTML         string `json:"html"`          // Raw HTML content
    StatusCode   int    `json:"status_code"`   // HTTP status code
    ErrorMessage string `json:"error,omitempty"` // Error message if any
}
```

### ScrapperConfig

Configuration options for the scrapper:

```go
type ScrapperConfig struct {
    UserAgent   string            // User agent for requests
    Delay       time.Duration     // Delay between requests
    BaseURL     string            // Base URL (required)
    Parallelism int               // Max concurrent requests
    Timeout     time.Duration     // Request timeout
    Debug       bool              // Enable debug mode
    Headers     map[string]string // Additional headers
}
```

## Testing

The scrapper includes comprehensive tests for content extraction functionality.

Run the tests:

```bash
go test ./utils/scrapper/... -v
```

Run specific test cases:

```bash
go test ./utils/scrapper -run TestExtractContent -v
```

The test suite covers:
- Text content extraction from various HTML elements
- Attribute extraction using CSS selectors
- HTML content extraction with `[html]` attribute
- Edge cases and error handling
- Whitespace trimming and empty content filtering

## Error Handling

The scrapper provides robust error handling for various scenarios:

- **URL Validation**: Checks URL accessibility before scraping
- **Network Errors**: Handles connection timeouts and network failures
- **HTTP Errors**: Captures HTTP status codes (4xx, 5xx) in response data
- **Rate Limiting**: Automatic retry with exponential backoff for 429 status codes
- **Content Extraction**: Detailed error messages for CSS selector issues
- **Parsing Errors**: GoQuery parsing errors for malformed HTML

All errors include contextual information for debugging. Rate limiting (429 responses) triggers automatic retry with exponential backoff up to 5 attempts.

## Best Practices

1. **BaseURL Configuration**: Always set a valid BaseURL that represents your target domain
2. **Rate Limiting**: Configure appropriate delays between requests to respect server resources
3. **User Agent**: Use descriptive user agents to identify your scraping bot
4. **Error Handling**: Always check `ErrorMessage` field in `ScrapedData` responses
5. **Timeouts**: Set reasonable timeouts based on expected response times
6. **CSS Selectors**: Use specific selectors to avoid extracting unwanted content
7. **Concurrent Scraping**: Monitor the results channel in batch operations to handle errors promptly
8. **URL Validation**: Validate URLs before batch scraping to avoid unnecessary requests

## Examples

### Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/unluckythoughts/go-microservice/tools/web"
    "github.com/unluckythoughts/go-microservice/tools/logger"
    "github.com/unluckythoughts/go-microservice/utils/scrapper"
)

func main() {
    // Initialize context with logger
    ctx := web.NewContext(logger.New(logger.Options{}))

    // Configure scrapper for target website
    config := scrapper.ScrapperConfig{
        BaseURL:     "https://httpbin.org",
        UserAgent:   "Example Bot 1.0",
        Delay:       time.Second,
        Parallelism: 5,
        Timeout:     15 * time.Second,
        Headers: map[string]string{
            "Accept": "text/html,application/xhtml+xml",
        },
    }

    s := scrapper.NewScrapper(config, ctx)

    // Single URL scraping
    data, err := s.ScrapeURL("/html", nil)
    if err != nil {
        log.Printf("Scraping failed: %v", err)
        return
    }

    fmt.Printf("Scraped %s (Status: %d)\n", data.URL, data.StatusCode)

    // Extract page title
    titles, err := scrapper.ExtractContent(data.HTML, "title")
    if err == nil && len(titles) > 0 {
        fmt.Printf("Title: %s\n", titles[0])
    }

    // Extract all paragraph text
    paragraphs, err := scrapper.ExtractContent(data.HTML, "p")
    if err == nil {
        fmt.Printf("Found %d paragraphs\n", len(paragraphs))
    }

    // Batch scraping
    urls := []string{"/html", "/json", "/xml"}
    resultChan := s.ScrapeURLs(urls)

    for result := range resultChan {
        if result.ErrorMessage == "" {
            fmt.Printf("Success: %s\n", result.URL)
        } else {
            fmt.Printf("Failed: %s - %s\n", result.URL, result.ErrorMessage)
        }
    }
}
```

See `examples/scrapper/main.go` for more comprehensive examples.

## Dependencies

- [Colly v2](https://github.com/gocolly/colly) - Fast and elegant web scraping framework
- [GoQuery](https://github.com/PuerkitoBio/goquery) - jQuery-like HTML parsing and manipulation
- [Web Tools](../../tools/web) - Microservice web framework with integrated logging
- [Logger Tools](../../tools/logger) - Structured logging framework integration
- [pkg/errors](https://github.com/pkg/errors) - Error handling with stack traces

## Framework Integration

The scrapper integrates seamlessly with the go-microservice framework through `web.Context`:

- **Structured Logging**: Uses the framework's logging system for consistent log formatting
- **Request Correlation**: Inherits request tracing and correlation IDs when used in web handlers  
- **Error Wrapping**: Leverages pkg/errors for detailed stack traces and error context
- **Configuration**: Follows framework patterns for service configuration

### Usage in Web Handlers

```go
func scrapeHandler(ctx web.Context) web.Response {
    config := scrapper.ScrapperConfig{
        BaseURL: "https://target-site.com",
        Delay:   time.Second,
    }
    
    s := scrapper.NewScrapper(config, ctx)
    data, err := s.ScrapeURL("/endpoint", nil)
    
    if err != nil {
        return ctx.ErrorResponse(500, "Scraping failed", err)
    }
    
    return ctx.JSONResponse(200, data)
}
```