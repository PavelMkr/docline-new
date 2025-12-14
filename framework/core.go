package framework

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Framework is the main entry point for the DocLine framework
type Framework struct {
	registry *PluginRegistry
	config   *Config
}

// Config holds framework-wide configuration
type Config struct {
	DefaultCloneFinder     string
	DefaultSimilarityCalc  string
	DefaultTokenizer       string
	DefaultReportFormat    string
	ResultsDirectory       string
	EnableLogging          bool
	CustomSettings         map[string]interface{}
}

// NewFramework creates a new framework instance
func NewFramework(config *Config) *Framework {
	if config == nil {
		config = &Config{
			ResultsDirectory: "./results",
			DefaultReportFormat: "html",
		}
	}
	
	return &Framework{
		registry: NewPluginRegistry(),
		config:   config,
	}
}

// GetRegistry returns the plugin registry
func (f *Framework) GetRegistry() *PluginRegistry {
	return f.registry
}

// AnalyzeDocument performs complete analysis of a document
func (f *Framework) AnalyzeDocument(filePath string, finderName string, finderConfig CloneFinderConfig) (*AnalysisResult, error) {
	// Read and parse document
	content, err := f.readDocument(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read document: %v", err)
	}
	
	// Get clone finder
	finder, err := f.registry.GetCloneFinder(finderName)
	if err != nil {
		return nil, fmt.Errorf("failed to get clone finder: %v", err)
	}
	
	// Find clones
	groups, err := finder.FindClones(content, finderConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to find clones: %v", err)
	}
	
	// Calculate statistics
	stats := f.calculateStatistics(groups)
	
	// Build result
	result := &AnalysisResult{
		Groups:     groups,
		Statistics: stats,
		Config:     finderConfig,
		Metadata: map[string]interface{}{
			"source_file": filePath,
			"finder":      finderName,
		},
	}
	
	return result, nil
}

// GenerateReport generates a report from analysis results
func (f *Framework) GenerateReport(result *AnalysisResult, format string, outputPath string) error {
	generator, err := f.registry.GetReportGenerator(format)
	if err != nil {
		return fmt.Errorf("failed to get report generator: %v", err)
	}
	
	reportConfig := ReportConfig{
		Title:      "Clone Analysis Report",
		SourceFile: result.Metadata["source_file"].(string),
		Settings:   result.Metadata,
		OutputDir:  filepath.Dir(outputPath),
	}
	
	return generator.Generate(result.Groups, reportConfig, outputPath)
}

// readDocument reads and parses a document using appropriate parser/converter
func (f *Framework) readDocument(filePath string) (string, error) {
	ext := filepath.Ext(filePath)
	
	// Try to get parser for this format
	parser, err := f.registry.GetDocumentParser(ext)
	if err == nil {
		// Parser found, use it
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer file.Close()
		
		segments, err := parser.Parse(file)
		if err != nil {
			return "", err
		}
		
		// Join segments
		content := ""
		for i, seg := range segments {
			if i > 0 {
				content += "\n"
			}
			content += seg
		}
		return content, nil
	}
	
	// No parser found, check if conversion is needed
	converter, err := f.registry.GetDocumentConverter("pandoc")
	if err == nil && converter.IsConversionNeeded(filePath) {
		// Convert to DocBook first
		tempPath, err := converter.Convert(filePath, ".xml")
		if err != nil {
			return "", fmt.Errorf("conversion failed: %v", err)
		}
		defer os.Remove(tempPath)
		
		// Try parsing the converted file
		parser, err := f.registry.GetDocumentParser(".xml")
		if err == nil {
			file, err := os.Open(tempPath)
			if err != nil {
				return "", err
			}
			defer file.Close()
			
			segments, err := parser.Parse(file)
			if err != nil {
				return "", err
			}
			
			content := ""
			for i, seg := range segments {
				if i > 0 {
					content += "\n"
				}
				content += seg
			}
			return content, nil
		}
	}
	
	// Fallback: read as plain text
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	
	return string(content), nil
}

// calculateStatistics computes statistics from clone groups
func (f *Framework) calculateStatistics(groups []CloneGroup) AnalysisStatistics {
	stats := AnalysisStatistics{
		TotalGroups:    len(groups),
		TotalFragments: 0,
		MinTokens:      -1,
	}
	
	totalTokens := 0
	tokenCount := 0
	
	for _, group := range groups {
		stats.TotalFragments += len(group.Fragments)
		
		for _, frag := range group.Fragments {
			// Simple token count (split by spaces)
			tokens := len(f.tokenize(frag.Content))
			totalTokens += tokens
			tokenCount++
			
			if stats.MinTokens == -1 || tokens < stats.MinTokens {
				stats.MinTokens = tokens
			}
			if tokens > stats.MaxTokens {
				stats.MaxTokens = tokens
			}
		}
	}
	
	if tokenCount > 0 {
		stats.AvgTokens = float64(totalTokens) / float64(tokenCount)
	}
	
	return stats
}

// tokenize splits text into tokens (simple implementation)
func (f *Framework) tokenize(text string) []string {
	// Use default tokenizer if available
	tokenizer, err := f.registry.GetTextTokenizer(f.config.DefaultTokenizer)
	if err == nil {
		return tokenizer.Tokenize(text)
	}
	
	// Fallback to simple space-based tokenization
	words := make([]string, 0)
	current := ""
	for _, r := range text {
		if r == ' ' || r == '\n' || r == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

// ReadDocument is a convenience method that reads a document using the framework
func ReadDocument(f *Framework, filePath string) (io.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}



