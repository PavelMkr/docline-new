package main

import (
	"fmt"
	"strings"
)

// AutomaticModeSettings represents the settings for automatic mode
type AutomaticModeSettings struct {
	MinCloneLength  int    `json:"minCloneLength"`
	ConvertToDRL    bool   `json:"convertToDRL"`
	ArchetypeLength int    `json:"archetypeLength"`
	StrictFilter    bool   `json:"strictFilter"`
	FilePath        string `json:"filePath,omitempty"`
}

// AutomaticModeResponse represents the response for automatic mode analysis
type AutomaticModeResponse struct {
	Status      string              `json:"status"`
	Message     string              `json:"message"`
	Groups      map[string][]string `json:"groups,omitempty"`
	ResultsFile string              `json:"results_file,omitempty"`
}

// TextFragment represents a fragment of text with its position
type TextFragment struct {
	Content  string
	StartPos int
	EndPos   int
}

// CloneGroup represents a group of similar text fragments
type CloneGroup struct {
	Fragments []TextFragment
	Archetype string
	Power     int // Number of fragments in the group
}

// convertToDRL converts text to DRL format
func convertToDRL(text string) string {
	// TODO: Implement actual DRL conversion
	// For now, just return the text with some basic preprocessing
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return text
}

// findClones finds similar text fragments using the Clone Miner algorithm
func findClones(text string, settings AutomaticModeSettings) []CloneGroup {
	// Split text into tokens
	tokens := strings.Fields(text)
	if len(tokens) < settings.MinCloneLength {
		return nil
	}

	var groups []CloneGroup
	// Create a sliding window of tokens
	for i := 0; i <= len(tokens)-settings.MinCloneLength; i++ {
		window := tokens[i : i+settings.MinCloneLength]
		windowText := strings.Join(window, " ")

		// Check if this window is similar to any existing group
		found := false
		for j := range groups {
			if isSimilar(windowText, groups[j].Archetype) {
				groups[j].Fragments = append(groups[j].Fragments, TextFragment{
					Content:  windowText,
					StartPos: i,
					EndPos:   i + settings.MinCloneLength,
				})
				groups[j].Power++
				found = true
				break
			}
		}

		// If no similar group found, create a new one
		if !found {
			groups = append(groups, CloneGroup{
				Fragments: []TextFragment{{
					Content:  windowText,
					StartPos: i,
					EndPos:   i + settings.MinCloneLength,
				}},
				Archetype: windowText,
				Power:     1,
			})
		}
	}

	// Apply strict filtering if enabled
	if settings.StrictFilter {
		groups = filterCloneGroups(groups, settings)
	}

	return groups
}

// isSimilar checks if two text fragments are similar enough
func isSimilar(a, b string) bool {
	// TODO: Implement actual similarity check
	// For now, use simple string comparison
	return a == b
}

// filterCloneGroups applies strict filtering to clone groups
func filterCloneGroups(groups []CloneGroup, settings AutomaticModeSettings) []CloneGroup {
	var filtered []CloneGroup

	for _, group := range groups {
		// Skip groups with too few fragments
		if len(group.Fragments) < 2 {
			continue
		}

		// Skip groups with archetype shorter than minimal length
		if len(strings.Fields(group.Archetype)) < settings.ArchetypeLength {
			continue
		}

		// Remove overlapping fragments
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

		if len(nonOverlapping) >= 2 {
			group.Fragments = nonOverlapping
			group.Power = len(nonOverlapping)
			filtered = append(filtered, group)
		}
	}

	return filtered
}

// hasOverlap checks if two text fragments overlap
func hasOverlap(a, b TextFragment) bool {
	return (a.StartPos <= b.EndPos && b.StartPos <= a.EndPos)
}

// ProcessAutomaticMode processes the text using automatic mode settings
func ProcessAutomaticMode(text string, settings AutomaticModeSettings) ([]CloneGroup, error) {
	// Convert to DRL if needed
	if settings.ConvertToDRL {
		text = convertToDRL(text)
	}

	// Find clones
	groups := findClones(text, settings)

	// Convert groups to response format
	responseGroups := make(map[string][]string)
	for i, group := range groups {
		groupKey := fmt.Sprintf("group%d", i+1)
		var fragments []string
		for _, frag := range group.Fragments {
			fragments = append(fragments, frag.Content)
		}
		responseGroups[groupKey] = fragments
	}

	return groups, nil
}

// FormatAutomaticModeResults formats the analysis results for output
func FormatAutomaticModeResults(groups []CloneGroup, settings AutomaticModeSettings) string {
	var sb strings.Builder

	sb.WriteString("Automatic Mode Analysis Results\n")
	sb.WriteString("=============================\n\n")
	sb.WriteString(fmt.Sprintf("Settings:\n"))
	sb.WriteString(fmt.Sprintf("- Minimal Clone Length: %d tokens\n", settings.MinCloneLength))
	sb.WriteString(fmt.Sprintf("- Convert to DRL: %v\n", settings.ConvertToDRL))
	sb.WriteString(fmt.Sprintf("- Minimal Archetype Length: %d tokens\n", settings.ArchetypeLength))
	sb.WriteString(fmt.Sprintf("- Strict Filtering: %v\n\n", settings.StrictFilter))

	sb.WriteString(fmt.Sprintf("Found %d clone groups:\n\n", len(groups)))
	for i, group := range groups {
		sb.WriteString(fmt.Sprintf("Group %d (Power: %d):\n", i+1, group.Power))
		sb.WriteString(fmt.Sprintf("Archetype: %s\n", group.Archetype))
		sb.WriteString("Fragments:\n")
		for j, frag := range group.Fragments {
			sb.WriteString(fmt.Sprintf("  %d. [%d-%d] %s\n", j+1, frag.StartPos, frag.EndPos, frag.Content))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
