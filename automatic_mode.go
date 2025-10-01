package main

import (
	"fmt"
	"regexp"
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

// convertToDRL converts text to DRL format
func convertToDRL(text string) string {
	// Basic deterministic representation of DRL
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ToLower(text)
	// Replace any non letter/digit whitespace combo with space
	// Keep unicode letters and digits roughly by removing common punctuation.
	// Go's regex character classes are ASCII-focused; simplify to strip punctuation symbols.
	punctuation := regexp.MustCompile(`[\p{P}\p{S}]+`)
	text = punctuation.ReplaceAllString(text, " ")
	spaceCollapse := regexp.MustCompile(`\s+`)
	text = spaceCollapse.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	return text
}

// findClones finds similar text fragments using the Clone Miner algorithm
func findClones(text string, settings AutomaticModeSettings) []CloneGroup {
	// Split text into tokens
	tokens := strings.Fields(text)
	if len(tokens) < settings.MinCloneLength {
		return nil
	}

	windowSize := settings.MinCloneLength
	total := len(tokens) - windowSize + 1
	if total <= 0 {
		return nil
	}

	// Two-pass exact-window grouping for performance
	// Pass 1: frequency count
	freq := make(map[string]int, total)
	for i := 0; i <= len(tokens)-windowSize; i++ {
		w := strings.Join(tokens[i:i+windowSize], " ")
		freq[w]++
	}

	// Pass 2: collect positions for windows with freq >= 2
	candidates := make(map[string][]TextFragment)
	for k, c := range freq {
		if c >= 2 {
			candidates[k] = nil
		}
	}
	for i := 0; i <= len(tokens)-windowSize; i++ {
		w := strings.Join(tokens[i:i+windowSize], " ")
		if _, ok := candidates[w]; ok {
			candidates[w] = append(candidates[w], TextFragment{
				Content:  w,
				StartPos: i,
				EndPos:   i + windowSize,
			})
		}
	}

	// Build groups
	var groups []CloneGroup
	for archetype, frags := range candidates {
		if len(frags) < 2 {
			continue
		}
		groups = append(groups, CloneGroup{
			Fragments: frags,
			Power:     len(frags),
			Archetype: archetype,
		})
	}

	// Merge groups with similar archetypes using isSimilar
	var merged []CloneGroup
	for _, g := range groups {
		mergedIntoExisting := false
		for mi := range merged {
			if isSimilar(g.Archetype, merged[mi].Archetype) {
				merged[mi].Fragments = append(merged[mi].Fragments, g.Fragments...)
				merged[mi].Power = len(merged[mi].Fragments)
				mergedIntoExisting = true
				break
			}
		}
		if !mergedIntoExisting {
			merged = append(merged, g)
		}
	}
	groups = merged

	// Apply strict filtering if enabled
	if settings.StrictFilter {
		groups = filterCloneGroups(groups, settings)
	}

	return groups
}

// isSimilar checks if two text fragments are similar enough
func isSimilar(a, b string) bool {
	// Token-level Jaccard similarity of unigrams
	aTokens := strings.Fields(a)
	bTokens := strings.Fields(b)
	if len(aTokens) == 0 && len(bTokens) == 0 {
		return true
	}
	aSet := make(map[string]struct{})
	for _, t := range aTokens {
		aSet[t] = struct{}{}
	}
	bSet := make(map[string]struct{})
	for _, t := range bTokens {
		bSet[t] = struct{}{}
	}
	intersection := 0
	union := 0
	seen := make(map[string]struct{})
	for t := range aSet {
		if _, ok := bSet[t]; ok {
			intersection++
		}
		seen[t] = struct{}{}
		union++
	}
	for t := range bSet {
		if _, ok := seen[t]; ok {
			continue
		}
		union++
	}
	if union == 0 {
		return false
	}
	jaccard := float64(intersection) / float64(union)
	// Conservative threshold since windows are exact-length token slices
	return jaccard >= 0.9
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
		sb.WriteString("Fragments (in tokens):\n")
		for j, frag := range group.Fragments {
			sb.WriteString(fmt.Sprintf("  %d. [%d-%d] %s\n", j+1, frag.StartPos, frag.EndPos, frag.Content))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
