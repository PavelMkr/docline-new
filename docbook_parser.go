package main

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"strings"
)

// DocBookElement represents a DocBook XML element
type DocBookElement struct {
	XMLName  xml.Name
	Content  string           `xml:",chardata"`
	Elements []DocBookElement `xml:",any"`
	Attrs    []xml.Attr       `xml:",any,attr"`
}

// DocBookParser handles parsing of DocBook XML files
type DocBookParser struct {
	// Elements to extract text from (e.g., para, section, chapter)
	TextElements map[string]bool
}

// NewDocBookParser creates a new DocBook parser with default settings
func NewDocBookParser() *DocBookParser {
	return &DocBookParser{
		TextElements: map[string]bool{
			"para":      true,
			"section":   true,
			"chapter":   true,
			"title":     true,
			"simpara":   true,
			"note":      true,
			"warning":   true,
			"important": true,
			"tip":       true,
		},
	}
}

// ParseDocBook parses a DocBook XML file and returns extracted text segments
func (p *DocBookParser) ParseDocBook(reader io.Reader) ([]string, error) {
	fmt.Printf("Starting DocBook parsing...\n")

	// read file content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %v", err)
	}
	fmt.Printf("Read %d bytes from file\n", len(content))

	// check if the file starts with XML declaration
	contentStr := string(content)
	if !strings.HasPrefix(strings.TrimSpace(contentStr), "<?xml") {
		fmt.Printf("Warning: File does not start with XML declaration\n")
		fmt.Printf("First 100 characters: %s\n", contentStr[:min(len(contentStr), 100)])
	}

	// parse XML directly, without preprocessing
	var doc DocBookElement
	decoder := xml.NewDecoder(strings.NewReader(contentStr))
	decoder.Strict = false // allow more flexible parsing

	// set handler for HTML entities
	decoder.Entity = map[string]string{
		"ndash": "–",  // long dash
		"mdash": "—",  // em-dash
		"nbsp":  " ",  // non-breaking space
		"lt":    "<",  // less than
		"gt":    ">",  // greater than
		"amp":   "&",  // ampersand
		"quot":  "\"", // quotation mark
		"apos":  "'",  // apostrophe
	}

	if err := decoder.Decode(&doc); err != nil {
		fmt.Printf("Error decoding XML: %v\n", err)
		fmt.Printf("Last 100 characters of content: %s\n",
			contentStr[max(0, len(contentStr)-100):])
		return nil, fmt.Errorf("failed to decode XML: %v", err)
	}
	fmt.Printf("Successfully decoded XML document, root element: %s\n", doc.XMLName.Local)

	var segments []string
	p.extractText(&doc, &segments)
	fmt.Printf("Extracted %d text segments from DocBook\n", len(segments))
	return segments, nil
}

// extractText recursively extracts text from DocBook elements
func (p *DocBookParser) extractText(element *DocBookElement, segments *[]string) {
	// Check if this is a text element we want to extract
	if p.TextElements[element.XMLName.Local] {
		// process HTML entities in text content
		text := html.UnescapeString(strings.TrimSpace(element.Content))
		if text != "" {
			// fmt.Printf("Found text in element <%s>: %s\n", element.XMLName.Local, text[:min(len(text), 50)]+"...")
			*segments = append(*segments, text)
		}
	}

	// Process child elements
	for i := range element.Elements {
		p.extractText(&element.Elements[i], segments)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// AddTextElement adds a new element type to extract text from
func (p *DocBookParser) AddTextElement(elementName string) {
	p.TextElements[elementName] = true
}

// RemoveTextElement removes an element type from text extraction
func (p *DocBookParser) RemoveTextElement(elementName string) {
	delete(p.TextElements, elementName)
}
