package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DocumentConverter handles document format conversion
type DocumentConverter struct {
	// Supported input formats
	SupportedInputFormats map[string]bool
	// Supported output formats
	SupportedOutputFormats map[string]bool
}

// NewDocumentConverter creates a new document converter
func NewDocumentConverter() *DocumentConverter {
	return &DocumentConverter{
		SupportedInputFormats: map[string]bool{
			".doc":  true,
			".docx": true,
			".odt":  true,
			".rtf":  true,
			".md":   true,
			".txt":  true,
			".html": true,
			".htm":  true,
		},
		SupportedOutputFormats: map[string]bool{
			".xml":     true, // DocBook XML
			".dbk":     true, // DocBook
			".docbook": true, // DocBook
		},
	}
}

// ConvertToDocBook converts a document to DocBook format using pandoc
func (c *DocumentConverter) ConvertToDocBook(inputPath string) (string, error) {
	// Check if input format is supported
	ext := strings.ToLower(filepath.Ext(inputPath))
	if !c.SupportedInputFormats[ext] {
		return "", fmt.Errorf("unsupported input format: %s", ext)
	}

	// Create temporary output file
	outputPath := filepath.Join(os.TempDir(), filepath.Base(inputPath)+".xml")

	// Prepare pandoc command
	cmd := exec.Command("pandoc",
		"-f", getPandocFormat(ext),
		"-t", "docbook",
		"-o", outputPath,
		inputPath)

	// Run pandoc
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pandoc conversion failed: %v", err)
	}

	return outputPath, nil
}

// getPandocFormat returns the pandoc format identifier for a file extension
func getPandocFormat(ext string) string {
	switch ext {
	case ".doc", ".docx":
		return "docx"
	case ".odt":
		return "odt"
	case ".rtf":
		return "rtf"
	case ".md":
		return "markdown"
	case ".txt":
		return "plain"
	case ".html", ".htm":
		return "html"
	default:
		return "plain"
	}
}

// IsConversionNeeded checks if the file needs to be converted to DocBook
func (c *DocumentConverter) IsConversionNeeded(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return !c.SupportedOutputFormats[ext] && c.SupportedInputFormats[ext]
}

// CleanupTempFile removes temporary converted file
func (c *DocumentConverter) CleanupTempFile(filePath string) error {
	return os.Remove(filePath)
}
