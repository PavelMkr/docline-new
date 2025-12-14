package framework

import (
	"io"
)

// CloneFinder defines the interface for clone detection algorithms
type CloneFinder interface {
	// FindClones searches for duplicate text fragments in the given text
	// Returns a list of clone groups found
	FindClones(text string, config CloneFinderConfig) ([]CloneGroup, error)
	
	// Name returns the name/identifier of this finder
	Name() string
	
	// Description returns a human-readable description of the algorithm
	Description() string
}

// SimilarityCalculator defines the interface for similarity calculation algorithms
type SimilarityCalculator interface {
	// CalculateSimilarity computes similarity score between two text fragments
	// Returns a value between 0.0 (completely different) and 1.0 (identical)
	CalculateSimilarity(text1, text2 string) float64
	
	// Name returns the name of the similarity algorithm
	Name() string
}

// DocumentParser defines the interface for parsing different document formats
type DocumentParser interface {
	// Parse extracts text content from a document
	// Returns text segments extracted from the document
	Parse(reader io.Reader) ([]string, error)
	
	// SupportedFormats returns list of file extensions this parser supports
	SupportedFormats() []string
	
	// Name returns the name of the parser
	Name() string
}

// DocumentConverter defines the interface for converting documents between formats
type DocumentConverter interface {
	// Convert converts a document from one format to another
	Convert(inputPath string, outputFormat string) (string, error)
	
	// IsConversionNeeded checks if conversion is required for the given file
	IsConversionNeeded(filePath string) bool
	
	// SupportedInputFormats returns list of input formats supported
	SupportedInputFormats() []string
	
	// SupportedOutputFormats returns list of output formats supported
	SupportedOutputFormats() []string
	
	// Name returns the name of the converter
	Name() string
}

// ReportGenerator defines the interface for generating analysis reports
type ReportGenerator interface {
	// Generate creates a report from clone groups
	Generate(groups []CloneGroup, config ReportConfig, outputPath string) error
	
	// Format returns the output format identifier (e.g., "html", "json", "csv")
	Format() string
	
	// Name returns the name of the report generator
	Name() string
}

// TextTokenizer defines the interface for tokenizing text
type TextTokenizer interface {
	// Tokenize splits text into tokens
	Tokenize(text string) []string
	
	// Name returns the name of the tokenizer
	Name() string
}

// Filter defines the interface for filtering clone groups
type Filter interface {
	// Filter removes or modifies clone groups based on criteria
	Filter(groups []CloneGroup, config FilterConfig) []CloneGroup
	
	// Name returns the name of the filter
	Name() string
}

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string
	
	// Version returns the version of the plugin
	Version() string
	
	// Initialize is called when the plugin is registered
	Initialize(config map[string]interface{}) error
	
	// Shutdown is called when the plugin is unregistered
	Shutdown() error
}



