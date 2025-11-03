package scrapper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ExtractHTML(htmlText string, selector string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlText))
	if err != nil {
		return nil, err
	}

	selection := doc.Find(selector)
	if selection.Length() == 0 {
		return nil, fmt.Errorf("no elements found for selector: %s", selector)
	}

	var results []string
	selection.Each(func(i int, s *goquery.Selection) {
		htmlContent, err := s.Html()
		if err != nil {
			return
		}
		text := strings.TrimSpace(htmlContent)
		if text != "" {
			results = append(results, text)
		}
	})

	if len(results) == 0 {
		return nil, fmt.Errorf("no text content found for selector: %s", selector)
	}

	return results, nil
}

// ExtractContent retrives content from html text based on css selector
func ExtractContent(htmlText string, selector string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlText))
	if err != nil {
		return nil, err
	}

	selection := doc.Find(selector)
	if selection.Length() == 0 {
		return nil, fmt.Errorf("no elements found for selector: %s", selector)
	}

	var results []string
	selection.Each(func(i int, s *goquery.Selection) {
		var text string
		if hasAttribute(selector) {
			attrName := getAttributeName(selector)
			if attrName == "" {
				return
			}

			if attrName == "html" {
				htmlContent, err := goquery.OuterHtml(s)
				if err != nil {
					return
				}
				text = strings.TrimSpace(htmlContent)
			} else {
				text = strings.TrimSpace(s.AttrOr(attrName, ""))
			}
		} else {
			text = strings.TrimSpace(s.Text())
		}
		if text != "" {
			results = append(results, text)
		}
	})

	if len(results) == 0 {
		return nil, fmt.Errorf("no text content found for selector: %s", selector)
	}

	return results, nil
}

// GetNumber retrieves float after removing non number characters from text in colly element
func GetNumber(text string) (float64, error) {
	// Remove all non-numeric characters except decimal point and minus sign
	var numStr strings.Builder
	for i, char := range text {
		// Allow leading minus sign
		if i == 0 && char == '-' {
			numStr.WriteRune(char)
			continue
		}

		if (char >= '0' && char <= '9') || char == '.' {
			numStr.WriteRune(char)
		}
	}

	cleanedText := numStr.String()
	if cleanedText == "" {
		return 0, fmt.Errorf("no valid number found in text: %s", text)
	}

	// Convert to float64
	result, err := strconv.ParseFloat(cleanedText, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse number from text '%s': %v", text, err)
	}

	return result, nil
}
