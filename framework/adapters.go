package framework

import (
	"fmt"
	"io"
	"strings"
)

// Adapters for existing implementations to work with framework interfaces

// AutomaticModeAdapter adapts the automatic mode to CloneFinder interface
type AutomaticModeAdapter struct {
	// This will be implemented by wrapping existing ProcessAutomaticMode
}

func (a *AutomaticModeAdapter) Name() string {
	return "automatic"
}

func (a *AutomaticModeAdapter) Description() string {
	return "Automatic mode clone finder using window-based exact matching"
}

func (a *AutomaticModeAdapter) FindClones(text string, config CloneFinderConfig) ([]CloneGroup, error) {
	// This would call the existing ProcessAutomaticMode function
	// For now, return empty - will be implemented when refactoring
	return nil, fmt.Errorf("not yet implemented - requires refactoring")
}

// InteractiveModeAdapter adapts the interactive mode to CloneFinder interface
type InteractiveModeAdapter struct{}

func (a *InteractiveModeAdapter) Name() string {
	return "interactive"
}

func (a *InteractiveModeAdapter) Description() string {
	return "Interactive mode clone finder with configurable length ranges"
}

func (a *InteractiveModeAdapter) FindClones(text string, config CloneFinderConfig) ([]CloneGroup, error) {
	// This would call the existing ProcessInteractiveMode function
	return nil, fmt.Errorf("not yet implemented - requires refactoring")
}

// NGramAdapter adapts the n-gram finder to CloneFinder interface
type NGramAdapter struct{}

func (a *NGramAdapter) Name() string {
	return "ngram"
}

func (a *NGramAdapter) Description() string {
	return "N-gram based duplicate finder using similarity metrics"
}

func (a *NGramAdapter) FindClones(text string, config CloneFinderConfig) ([]CloneGroup, error) {
	// This would call the existing FindDuplicatesByNGram function
	return nil, fmt.Errorf("not yet implemented - requires refactoring")
}

// JaccardSimilarityCalculator implements SimilarityCalculator using Jaccard similarity
type JaccardSimilarityCalculator struct{}

func (j *JaccardSimilarityCalculator) Name() string {
	return "jaccard"
}

func (j *JaccardSimilarityCalculator) CalculateSimilarity(text1, text2 string) float64 {
	tokens1 := strings.Fields(text1)
	tokens2 := strings.Fields(text2)
	
	set1 := make(map[string]struct{})
	for _, t := range tokens1 {
		set1[t] = struct{}{}
	}
	
	set2 := make(map[string]struct{})
	for _, t := range tokens2 {
		set2[t] = struct{}{}
	}
	
	intersection := 0
	union := 0
	seen := make(map[string]struct{})
	
	for t := range set1 {
		if _, ok := set2[t]; ok {
			intersection++
		}
		seen[t] = struct{}{}
		union++
	}
	
	for t := range set2 {
		if _, ok := seen[t]; !ok {
			union++
		}
	}
	
	if union == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(union)
}

// DocBookParserAdapter adapts existing DocBookParser to DocumentParser interface
type DocBookParserAdapter struct {
	parser interface{} // Will hold *DocBookParser from main package
}

func (d *DocBookParserAdapter) Name() string {
	return "docbook"
}

func (d *DocBookParserAdapter) SupportedFormats() []string {
	return []string{".xml", ".dbk", ".docbook"}
}

func (d *DocBookParserAdapter) Parse(reader io.Reader) ([]string, error) {
	// This would use the existing parser
	return nil, fmt.Errorf("not yet implemented - requires refactoring")
}

// PandocConverterAdapter adapts existing DocumentConverter to DocumentConverter interface
type PandocConverterAdapter struct {
	converter interface{} // Will hold *DocumentConverter from main package
}

func (p *PandocConverterAdapter) Name() string {
	return "pandoc"
}

func (p *PandocConverterAdapter) Convert(inputPath string, outputFormat string) (string, error) {
	// This would use the existing converter
	return "", fmt.Errorf("not yet implemented - requires refactoring")
}

func (p *PandocConverterAdapter) IsConversionNeeded(filePath string) bool {
	// This would use the existing converter
	return false
}

func (p *PandocConverterAdapter) SupportedInputFormats() []string {
	return []string{".doc", ".docx", ".odt", ".rtf", ".md", ".txt", ".html", ".htm"}
}

func (p *PandocConverterAdapter) SupportedOutputFormats() []string {
	return []string{".xml", ".dbk", ".docbook"}
}

// HTMLReportGenerator implements ReportGenerator for HTML reports
type HTMLReportGenerator struct{}

func (h *HTMLReportGenerator) Name() string {
	return "html"
}

func (h *HTMLReportGenerator) Format() string {
	return "html"
}

func (h *HTMLReportGenerator) Generate(groups []CloneGroup, config ReportConfig, outputPath string) error {
	// This would use the existing WriteResultsHTML function
	return fmt.Errorf("not yet implemented - requires refactoring")
}

// JSONReportGenerator implements ReportGenerator for JSON reports
type JSONReportGenerator struct{}

func (j *JSONReportGenerator) Name() string {
	return "json"
}

func (j *JSONReportGenerator) Format() string {
	return "json"
}

func (j *JSONReportGenerator) Generate(groups []CloneGroup, config ReportConfig, outputPath string) error {
	// This would generate JSON output
	return fmt.Errorf("not yet implemented - requires refactoring")
}

// SpaceTokenizer implements TextTokenizer using space-based tokenization
type SpaceTokenizer struct{}

func (s *SpaceTokenizer) Name() string {
	return "space"
}

func (s *SpaceTokenizer) Tokenize(text string) []string {
	return strings.Fields(text)
}

// StrictFilter implements Filter interface for strict filtering
type StrictFilter struct{}

func (s *StrictFilter) Name() string {
	return "strict"
}

func (s *StrictFilter) Filter(groups []CloneGroup, config FilterConfig) []CloneGroup {
	var filtered []CloneGroup
	
	for _, group := range groups {
		// Skip groups with too few fragments
		if len(group.Fragments) < 2 {
			continue
		}
		
		// Skip groups with archetype shorter than minimal length
		if config.MinArchetypeLength > 0 {
			archetypeTokens := len(strings.Fields(group.Archetype))
			if archetypeTokens < config.MinArchetypeLength {
				continue
			}
		}
		
		// Remove overlapping fragments if requested
		if config.RemoveOverlaps {
			var nonOverlapping []TextFragment
			for i, frag := range group.Fragments {
				overlaps := false
				for j := 0; j < i; j++ {
					if hasOverlap(frag, group.Fragments[j]) {
						overlaps = true
						break
					}
				}
				if !overlaps {
					nonOverlapping = append(nonOverlapping, frag)
				}
			}
			group.Fragments = nonOverlapping
			group.Power = len(nonOverlapping)
		}
		
		if len(group.Fragments) >= 2 {
			filtered = append(filtered, group)
		}
	}
	
	return filtered
}

// hasOverlap checks if two text fragments overlap
func hasOverlap(a, b TextFragment) bool {
	return (a.StartPos <= b.EndPos && b.StartPos <= a.EndPos)
}



