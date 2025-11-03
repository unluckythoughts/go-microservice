package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/unluckythoughts/go-microservice/tools/logger"
	"github.com/unluckythoughts/go-microservice/tools/web"
	"github.com/unluckythoughts/go-microservice/utils/scrapper"
)

func main() {
	// Create web context with logger
	ctx := web.NewContext(logger.New(logger.Options{}))

	// Example 1: Basic scraping with default configuration
	fmt.Println("=== Example 1: Basic Scraping ===")
	basicConfig := scrapper.ScrapperConfig{
		BaseURL:     "https://httpbin.org",
		UserAgent:   "Example Bot 1.0",
		Delay:       time.Second,
		Parallelism: 10,
		Timeout:     30 * time.Second,
	}
	basicScrapper := scrapper.NewScrapper(basicConfig, ctx)

	data, err := basicScrapper.ScrapeURL("/html", nil)
	if err != nil {
		log.Printf("Error scraping: %v", err)
	} else {
		fmt.Printf("URL: %s\n", data.URL)
		fmt.Printf("Status Code: %d\n", data.StatusCode)
		fmt.Printf("HTML Length: %d bytes\n", len(data.HTML))

		// Extract title from HTML
		titles, err := scrapper.ExtractContent(data.HTML, "title")
		if err == nil && len(titles) > 0 {
			fmt.Printf("Title: %s\n", titles[0])
		}

		// Extract links
		links, err := scrapper.ExtractContent(data.HTML, "a[href]")
		if err == nil {
			fmt.Printf("Number of links: %d\n", len(links))
		}
	}

	// Example 2: Custom configuration with different base URL
	fmt.Println("\n=== Example 2: Custom Configuration ===")
	config := scrapper.ScrapperConfig{
		BaseURL:     "https://example.com",
		UserAgent:   "Custom Bot 1.0",
		Delay:       2 * time.Second,
		Parallelism: 1,
		Timeout:     15 * time.Second,
		Headers: map[string]string{
			"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		},
	}

	customScrapper := scrapper.NewScrapper(config, ctx)

	// Validate URL before scraping
	testPath := "/"
	if err := customScrapper.ValidateURL(testPath); err != nil {
		log.Printf("URL validation failed: %v", err)
	} else {
		fmt.Println("URL is valid and accessible")

		data, err := customScrapper.ScrapeURL(testPath, nil)
		if err != nil {
			log.Printf("Error scraping: %v", err)
		} else {
			// Pretty print JSON
			jsonData, _ := json.MarshalIndent(data, "", "  ")
			fmt.Printf("Scraped data:\n%s\n", string(jsonData))

			// Extract some content
			paragraphs, err := scrapper.ExtractContent(data.HTML, "p")
			if err == nil {
				fmt.Printf("Found %d paragraphs\n", len(paragraphs))
			}
		}
	}

	// Example 3: Scraping multiple URLs
	fmt.Println("\n=== Example 3: Multiple URLs ===")
	urls := []string{
		"/html",
		"/json",
		"/xml",
	}

	resultChan := basicScrapper.ScrapeURLs(urls)

	for result := range resultChan {
		if result.ErrorMessage == "" {
			fmt.Printf("Success: %s (Status: %d)\n", result.URL, result.StatusCode)

			// Try to extract title from HTML pages
			if result.StatusCode == 200 && len(result.HTML) > 0 {
				titles, err := scrapper.ExtractContent(result.HTML, "title")
				if err == nil && len(titles) > 0 {
					fmt.Printf("  Title: %s\n", titles[0])
				}
			}
		} else {
			fmt.Printf("Failed: %s - %s\n", result.URL, result.ErrorMessage)
		}
	}

	// Example 4: Content extraction with CSS selectors
	fmt.Println("\n=== Example 4: Content Extraction ===")

	// First get some HTML content
	htmlData, err := basicScrapper.ScrapeURL("/html", nil)
	if err != nil {
		log.Printf("Error getting HTML for extraction: %v", err)
		return
	}

	// Extract different types of content
	fmt.Println("Extracting content from HTML...")

	// Extract all paragraph text
	paragraphs, err := scrapper.ExtractContent(htmlData.HTML, "p")
	if err == nil {
		fmt.Printf("Found %d paragraphs:\n", len(paragraphs))
		for i, p := range paragraphs {
			if i < 3 { // Show first 3
				fmt.Printf("  - %s\n", p)
			}
		}
	}

	// Extract all link URLs
	linkURLs, err := scrapper.ExtractContent(htmlData.HTML, "a[href]")
	if err == nil {
		fmt.Printf("Found %d links:\n", len(linkURLs))
		for i, link := range linkURLs {
			if i < 5 { // Show first 5
				fmt.Printf("  - %s\n", link)
			}
		}
	}

	// Extract all heading text
	headings, err := scrapper.ExtractContent(htmlData.HTML, "h1, h2, h3, h4, h5, h6")
	if err == nil {
		fmt.Printf("Found %d headings:\n", len(headings))
		for _, heading := range headings {
			fmt.Printf("  - %s\n", heading)
		}
	}
}
