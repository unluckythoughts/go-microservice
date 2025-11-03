package scrapper

import (
	"testing"
)

func TestExtractContent(t *testing.T) {
	tests := []struct {
		name        string
		htmlText    string
		selector    string
		expected    []string
		expectError bool
	}{
		{
			name:     "extract text from simple div",
			htmlText: `<html><body><div class="content">Hello World</div></body></html>`,
			selector: ".content",
			expected: []string{"Hello World"},
		},
		{
			name:     "extract multiple elements",
			htmlText: `<html><body><p>First</p><p>Second</p><p>Third</p></body></html>`,
			selector: "p",
			expected: []string{"First", "Second", "Third"},
		},
		{
			name:     "extract from nested elements",
			htmlText: `<html><body><div><span>Nested Text</span></div></body></html>`,
			selector: "span",
			expected: []string{"Nested Text"},
		},
		{
			name:     "extract with id selector",
			htmlText: `<html><body><div id="main">Main Content</div></body></html>`,
			selector: "#main",
			expected: []string{"Main Content"},
		},
		{
			name:     "extract with attribute selector",
			htmlText: `<html><body><a href="/link">Link Text</a></body></html>`,
			selector: "a[href]",
			expected: []string{"/link"},
		},
		{
			name:     "extract and trim whitespace",
			htmlText: `<html><body><p>  Trimmed Text  </p></body></html>`,
			selector: "p",
			expected: []string{"Trimmed Text"},
		},
		{
			name:        "no elements found",
			htmlText:    `<html><body><div>Content</div></body></html>`,
			selector:    ".nonexistent",
			expectError: true,
		},
		{
			name:        "empty text content",
			htmlText:    `<html><body><div class="empty">   </div></body></html>`,
			selector:    ".empty",
			expectError: true,
		},
		{
			name:        "invalid HTML",
			htmlText:    `<invalid>broken html`,
			selector:    "div",
			expectError: true,
		},
		{
			name: "complex HTML with multiple matches",
			htmlText: `
				<html>
					<body>
						<div class="item">Item 1</div>
						<div class="item">Item 2</div>
						<div class="other">Other</div>
						<div class="item">Item 3</div>
					</body>
				</html>`,
			selector: ".item",
			expected: []string{"Item 1", "Item 2", "Item 3"},
		},
		{
			name:     "extract text with HTML entities",
			htmlText: `<html><body><p>&lt;Hello &amp; Goodbye&gt;</p></body></html>`,
			selector: "p",
			expected: []string{"<Hello & Goodbye>"},
		},
		{
			name: "extract from table cells",
			htmlText: `
				<html>
					<body>
						<table>
							<tr>
								<td>Cell 1</td>
								<td>Cell 2</td>
							</tr>
						</table>
					</body>
				</html>`,
			selector: "td",
			expected: []string{"Cell 1", "Cell 2"},
		},
		{
			name:     "extract from elements with mixed content",
			htmlText: `<html><body><div>Text <span>Nested</span> More Text</div></body></html>`,
			selector: "div",
			expected: []string{"Text Nested More Text"},
		},
		{
			name:     "extract from multiple nested levels",
			htmlText: `<html><body><div><ul><li>Item 1</li><li>Item 2</li></ul></div></body></html>`,
			selector: "li",
			expected: []string{"Item 1", "Item 2"},
		},
		{
			name:     "extract with descendant combinator",
			htmlText: `<html><body><div class="parent"><span class="child">Child 1</span></div><span class="child">Child 2</span></body></html>`,
			selector: ".parent .child",
			expected: []string{"Child 1"},
		},
		{
			name:     "extract with adjacent sibling combinator",
			htmlText: `<html><body><h1>Title</h1><p>First paragraph</p><p>Second paragraph</p></body></html>`,
			selector: "h1 + p",
			expected: []string{"First paragraph"},
		},
		{
			name:     "extract from form elements",
			htmlText: `<html><body><form><input type="text" value="input value"><textarea>textarea content</textarea></form></body></html>`,
			selector: "textarea",
			expected: []string{"textarea content"}, // textarea has text content, input does not
		},
		{
			name:     "extract from elements with newlines and tabs",
			htmlText: "<html><body><pre>\n\tFormatted\n\tText\n</pre></body></html>",
			selector: "pre",
			expected: []string{"Formatted\n\tText"},
		},
		{
			name:     "extract with nth-child selector",
			htmlText: `<html><body><ul><li>First</li><li>Second</li><li>Third</li></ul></body></html>`,
			selector: "li:nth-child(2)",
			expected: []string{"Second"},
		},
		{
			name:     "extract from elements with comments",
			htmlText: `<html><body><div>Text <!-- comment --> More</div></body></html>`,
			selector: "div",
			expected: []string{"Text  More"},
		},
		{
			name:     "extract attribute from elements",
			htmlText: `<html><body><div data-attr="value">Text More</div></body></html>`,
			selector: "div[data-attr]",
			expected: []string{"value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractContent(tt.htmlText, tt.selector)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d results, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected result[%d] = %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestGetNumber(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expected    float64
		expectError bool
	}{
		{
			name:     "simple integer",
			text:     "123",
			expected: 123.0,
		},
		{
			name:     "simple decimal",
			text:     "123.45",
			expected: 123.45,
		},
		{
			name:     "negative number",
			text:     "-123.45",
			expected: -123.45,
		},
		{
			name:     "number with prefix text",
			text:     "Price: $123.45",
			expected: 123.45,
		},
		{
			name:     "number with suffix text",
			text:     "123.45 USD",
			expected: 123.45,
		},
		{
			name:     "number with mixed text",
			text:     "Total: $1,234.56 (including tax)",
			expected: 1234.56,
		},
		{
			name:     "number with spaces",
			text:     "   123.45   ",
			expected: 123.45,
		},
		{
			name:     "number with special characters",
			text:     "$1,234.56!@#",
			expected: 1234.56,
		},
		{
			name:     "zero value",
			text:     "0",
			expected: 0.0,
		},
		{
			name:     "zero decimal",
			text:     "0.00",
			expected: 0.0,
		},
		{
			name:     "negative zero",
			text:     "-0",
			expected: 0.0,
		},
		{
			name:     "number with percentage",
			text:     "85.5%",
			expected: 85.5,
		},
		{
			name:     "scientific notation base",
			text:     "1.23e5",
			expected: 1.23, // Only extracts 1.23, ignores 'e5'
		},
		{
			name:     "number with currency symbols",
			text:     "€123.45",
			expected: 123.45,
		},
		{
			name:     "multiple decimal points (invalid)",
			text:     "123.45.67",
			expected: 123.45, // Takes first valid number part
		},
		{
			name:        "no numbers",
			text:        "Hello World",
			expectError: true,
		},
		{
			name:        "empty string",
			text:        "",
			expectError: true,
		},
		{
			name:        "only special characters",
			text:        "!@#$%^&*()",
			expectError: true,
		},
		{
			name:        "only minus sign",
			text:        "-",
			expectError: true,
		},
		{
			name:        "only decimal point",
			text:        ".",
			expectError: true,
		},
		{
			name:     "number starting with decimal",
			text:     ".123",
			expected: 0.123,
		},
		{
			name:     "number ending with decimal",
			text:     "123.",
			expected: 123.0,
		},
		{
			name:     "large number",
			text:     "1234567890.123",
			expected: 1234567890.123,
		},
		{
			name:     "very small decimal",
			text:     "0.0001",
			expected: 0.0001,
		},
		{
			name:     "negative with text prefix",
			text:     "Loss: -$45.67",
			expected: 45.67, // Minus at beginning gets removed by non-numeric filter
		},
		{
			name:     "number with brackets",
			text:     "(123.45)",
			expected: 123.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetNumber(tt.text)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. Result: %f", result)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkExtractContent(b *testing.B) {
	htmlText := `
		<html>
			<body>
				<div class="content">Content 1</div>
				<div class="content">Content 2</div>
				<div class="content">Content 3</div>
				<div class="content">Content 4</div>
				<div class="content">Content 5</div>
			</body>
		</html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractContent(htmlText, ".content")
	}
}

func BenchmarkGetNumber(b *testing.B) {
	text := "Price: $1,234.56 (including tax)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetNumber(text)
	}
}

// Example tests for documentation
func ExampleExtractContent() {
	html := `<html><body><p class="price">$19.99</p><p class="price">$29.99</p></body></html>`
	prices, err := ExtractContent(html, ".price")
	if err != nil {
		panic(err)
	}

	for _, price := range prices {
		println(price)
	}
	// Output:
	// $19.99
	// $29.99
}

func ExampleGetNumber() {
	text := "Total: $1,234.56 (including tax)"
	number, err := GetNumber(text)
	if err != nil {
		panic(err)
	}

	println(number) // Output: 1234.56
}
