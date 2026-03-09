package framework

// CloneGroup represents a group of similar text fragments
type CloneGroup struct {
	Fragments []TextFragment
	Power     int    // Number of fragments in the group
	Archetype string // Representative text for the group
	Metadata  map[string]interface{} // Additional metadata
}

// TextFragment represents a single text fragment with position information
type TextFragment struct {
	Content  string // The actual text content
	StartPos int    // Starting position (token index)
	EndPos   int    // Ending position (token index)
	Metadata map[string]interface{} // Additional metadata
}

// CloneFinderConfig holds configuration for clone finders
type CloneFinderConfig struct {
	MinCloneLength int     // Minimum clone length in tokens
	MaxCloneLength int     // Maximum clone length in tokens (0 = unlimited)
	MinGroupPower  int     // Minimum number of fragments in a group
	SimilarityThreshold float64 // Minimum similarity score (0.0-1.0)
	CustomParams   map[string]interface{} // Algorithm-specific parameters
}

// ReportConfig holds configuration for report generation
type ReportConfig struct {
	Title       string                 // Report title
	SourceFile  string                 // Source file path
	Settings    map[string]interface{} // Analysis settings
	OutputDir   string                 // Output directory
	CustomParams map[string]interface{} // Format-specific parameters
}

// FilterConfig holds configuration for filters
type FilterConfig struct {
	MinArchetypeLength int     // Minimum archetype length
	RemoveOverlaps     bool    // Remove overlapping fragments
	StrictFiltering    bool    // Enable strict filtering
	CustomParams       map[string]interface{} // Filter-specific parameters
}

// AnalysisResult represents the complete result of an analysis
type AnalysisResult struct {
	Groups      []CloneGroup              // Found clone groups
	Statistics  AnalysisStatistics        // Analysis statistics
	Metadata    map[string]interface{}   // Additional metadata
	Config      CloneFinderConfig         // Configuration used
}

// AnalysisStatistics holds statistical information about the analysis
type AnalysisStatistics struct {
	TotalGroups    int     // Total number of clone groups
	TotalFragments int     // Total number of fragments
	AvgTokens      float64 // Average tokens per fragment
	MaxTokens      int     // Maximum tokens in a fragment
	MinTokens      int     // Minimum tokens in a fragment
}



