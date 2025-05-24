package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GenerateNGrams creates n-grams from input text.
func GenerateNGrams(text string, n int) []string {
	words := strings.Fields(text)
	var ngrams []string
	for i := 0; i <= len(words)-n; i++ {
		ngrams = append(ngrams, strings.Join(words[i:i+n], " "))
	}
	return ngrams
}

// writeToFile writes data to a file at the specified path
func writeToFile(filePath string, data string) error {
	fmt.Printf("writeToFile: Attempting to write to %s\n", filePath)
	fmt.Printf("writeToFile: Data length: %d bytes\n", len(data))

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	fmt.Printf("writeToFile: Ensuring directory exists: %s\n", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("writeToFile: Failed to create directory: %v\n", err)
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create or truncate the file
	fmt.Printf("writeToFile: Creating/truncating file\n")
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("writeToFile: Failed to create file: %v\n", err)
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write data to file
	fmt.Printf("writeToFile: Writing data to file\n")
	bytesWritten, err := file.WriteString(data)
	if err != nil {
		fmt.Printf("writeToFile: Failed to write data: %v\n", err)
		return fmt.Errorf("failed to write data: %v", err)
	}
	fmt.Printf("writeToFile: Successfully wrote %d bytes to file\n", bytesWritten)

	// Ensure data is written to disk
	if err := file.Sync(); err != nil {
		fmt.Printf("writeToFile: Failed to sync file: %v\n", err)
		return fmt.Errorf("failed to sync file: %v", err)
	}

	fmt.Printf("writeToFile: File write completed successfully\n")
	return nil
}

// readFileContent reads file content and handles different file types
func readFileContent(filePath string) (string, error) {
	fmt.Printf("Reading file content: %s\n", filePath)
	converter := NewDocumentConverter()
	var tempFilePath string
	var err error

	// Check if conversion is needed
	if converter.IsConversionNeeded(filePath) {
		fmt.Printf("Conversion needed for file: %s\n", filePath)
		tempFilePath, err = converter.ConvertToDocBook(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to convert file: %v", err)
		}
		defer converter.CleanupTempFile(tempFilePath)
		filePath = tempFilePath
	} else {
		fmt.Printf("No conversion needed for file: %s\n", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	fmt.Printf("Successfully read file content, size: %d bytes\n", len(content))

	// Check if file is DocBook/XML
	ext := strings.ToLower(filepath.Ext(filePath))
	fmt.Printf("File extension: %s\n", ext)
	if ext == ".xml" || ext == ".dbk" || ext == ".docbook" {
		fmt.Printf("Processing as DocBook/XML file\n")
		parser := NewDocBookParser()
		segments, err := parser.ParseDocBook(strings.NewReader(string(content)))
		if err != nil {
			return "", fmt.Errorf("failed to parse DocBook: %v", err)
		}
		fmt.Printf("Successfully parsed DocBook, found %d segments\n", len(segments))
		// Join segments with newlines
		return strings.Join(segments, "\n"), nil
	}

	fmt.Printf("Processing as regular text file\n")
	// For other file types, return content as is
	return string(content), nil
}

// splitTextIntoParts splits text into parts (for example, by sentences).
func splitTextIntoParts(text string) []string {
	// Use regex to find punctuation marks.
	re := regexp.MustCompile(`[.!?]`)
	parts := re.Split(text, -1)

	// Remove empty lines and whitespace.
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// validateFileFormat checks if the file format is supported
func validateFileFormat(filePath string) error {
	converter := NewDocumentConverter()
	ext := strings.ToLower(filepath.Ext(filePath))

	fmt.Printf("Validating file format: %s (extension: %s)\n", filePath, ext)
	fmt.Printf("Is supported output format: %v\n", converter.SupportedOutputFormats[ext])
	fmt.Printf("Is supported input format: %v\n", converter.SupportedInputFormats[ext])

	// check, is the file DocBook or a supported format for conversion
	if !converter.SupportedOutputFormats[ext] && !converter.SupportedInputFormats[ext] {
		return fmt.Errorf("unsupported file format: %s. Supported formats are: DocBook (.xml, .dbk, .docbook), Microsoft Word (.doc, .docx), OpenDocument (.odt), RTF (.rtf), Markdown (.md), Plain Text (.txt), HTML (.html, .htm)", ext)
	}
	return nil
}
