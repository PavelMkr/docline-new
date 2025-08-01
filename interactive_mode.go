package main

import (
	"fmt"
	"strings"
)

// InteractiveModeSettings represents the settings for interactive mode
type InteractiveModeSettings struct {
	MinCloneLength int    `json:"minCloneLength"`
	MaxCloneLength int    `json:"maxCloneLength"`
	MinGroupPower  int    `json:"minGroupPower"`
	UseArchetype   bool   `json:"useArchetype"`
	FilePath       string `json:"filePath,omitempty"`
}

// InteractiveModeResponse represents the response for interactive mode analysis
type InteractiveModeResponse struct {
	Status      string              `json:"status"`
	Message     string              `json:"message"`
	Groups      map[string][]string `json:"groups,omitempty"`
	Archetypes  map[string]string   `json:"archetypes,omitempty"`
	ResultsFile string              `json:"results_file,omitempty"`
}

// findInteractiveClones finds similar text fragments using interactive mode settings
func findInteractiveClones(text string, settings InteractiveModeSettings) []CloneGroup {
	fmt.Printf("Starting clone search with settings: minLength=%d, maxLength=%d, minPower=%d\n",
		settings.MinCloneLength, settings.MaxCloneLength, settings.MinGroupPower)

	// Split text into tokens
	tokens := strings.Fields(text)
	tokenCount := len(tokens)
	fmt.Printf("Text split into %d tokens\n", tokenCount)

	if tokenCount < settings.MinCloneLength {
		fmt.Printf("Text is too short (less than %d tokens)\n", settings.MinCloneLength)
		return nil
	}

	// Use a map to store potential clones by their content
	potentialClones := make(map[string][]TextFragment)
	maxLength := settings.MaxCloneLength
	if maxLength <= 0 {
		maxLength = tokenCount
	}

	// Process each possible window size
	for length := settings.MinCloneLength; length <= maxLength; length++ {
		fmt.Printf("Processing windows of size %d...\n", length)
		windowCount := tokenCount - length + 1
		processedWindows := 0

		// Create sliding windows of current size
		for i := 0; i <= tokenCount-length; i++ {
			window := tokens[i : i+length]
			windowText := strings.Join(window, " ")

			// Add to potential clones
			potentialClones[windowText] = append(potentialClones[windowText], TextFragment{
				Content:  windowText,
				StartPos: i,
				EndPos:   i + length,
			})

			processedWindows++
			if processedWindows%1000 == 0 {
				fmt.Printf("Processed %d/%d windows of size %d\n", processedWindows, windowCount, length)
			}
		}
	}

	fmt.Printf("Found %d unique text fragments\n", len(potentialClones))

	// Convert potential clones to groups
	var groups []CloneGroup
	for text, fragments := range potentialClones {
		if len(fragments) >= settings.MinGroupPower {
			// Remove overlapping fragments
			var nonOverlapping []TextFragment
			for i, frag := range fragments {
				overlaps := false
				for j := 0; j < i; j++ {
					if hasOverlap(frag, fragments[j]) {
						overlaps = true
						break
					}
				}
				if !overlaps {
					nonOverlapping = append(nonOverlapping, frag)
				}
			}

			if len(nonOverlapping) >= settings.MinGroupPower {
				groups = append(groups, CloneGroup{
					Fragments: nonOverlapping,
					Archetype: text,
					Power:     len(nonOverlapping),
				})
			}
		}
	}

	fmt.Printf("Found %d clone groups after filtering\n", len(groups))

	// Calculate archetypes if enabled
	if settings.UseArchetype {
		fmt.Printf("Calculating archetypes...\n")
		calculateArchetypes(&groups)
	}

	return groups
}

// isSimilarInteractive checks if two text fragments are similar enough for interactive mode
func isSimilarInteractive(a, b string) bool {
	// For now, use exact string comparison
	// TODO: Implement fuzzy matching for interactive mode
	return a == b
}

// filterInteractiveGroups filters clone groups based on interactive mode settings
func filterInteractiveGroups(groups []CloneGroup, settings InteractiveModeSettings) []CloneGroup {
	var filtered []CloneGroup

	for _, group := range groups {
		// Skip groups with insufficient power
		if group.Power < settings.MinGroupPower {
			continue
		}

		// Skip groups with fragments shorter than minimal length
		if len(strings.Fields(group.Fragments[0].Content)) < settings.MinCloneLength {
			continue
		}

		// Skip groups with fragments longer than maximal length
		if settings.MaxCloneLength > 0 && len(strings.Fields(group.Fragments[0].Content)) > settings.MaxCloneLength {
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

		if len(nonOverlapping) >= settings.MinGroupPower {
			group.Fragments = nonOverlapping
			group.Power = len(nonOverlapping)
			filtered = append(filtered, group)
		}
	}

	return filtered
}

// calculateArchetypes calculates archetypes for clone groups
func calculateArchetypes(groups *[]CloneGroup) {
	for i := range *groups {
		group := &(*groups)[i]
		if len(group.Fragments) == 0 {
			continue
		}

		// For now, use the longest fragment as archetype
		// TODO: Implement more sophisticated archetype calculation
		longestFrag := group.Fragments[0]
		for _, frag := range group.Fragments {
			if len(frag.Content) > len(longestFrag.Content) {
				longestFrag = frag
			}
		}
		group.Archetype = longestFrag.Content
	}
}

// ProcessInteractiveMode processes the text using interactive mode settings
func ProcessInteractiveMode(text string, settings InteractiveModeSettings) ([]CloneGroup, error) {
	// Find clones
	groups := findInteractiveClones(text, settings)

	return groups, nil
}

// FormatInteractiveModeResults formats the analysis results for output
func FormatInteractiveModeResults(groups []CloneGroup, settings InteractiveModeSettings) string {
	var sb strings.Builder

	sb.WriteString("Interactive Mode Analysis Results\n")
	sb.WriteString("===============================\n\n")
	sb.WriteString(fmt.Sprintf("Settings:\n"))
	sb.WriteString(fmt.Sprintf("- Minimal Clone Length: %d tokens\n", settings.MinCloneLength))
	sb.WriteString(fmt.Sprintf("- Maximal Clone Length: %d tokens\n", settings.MaxCloneLength))
	sb.WriteString(fmt.Sprintf("- Minimal Group Power: %d clones\n", settings.MinGroupPower))
	sb.WriteString(fmt.Sprintf("- Archetype Calculation: %v\n\n", settings.UseArchetype))

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
