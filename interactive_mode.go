package main

import (
	"fmt"
	"strings"

	"Docline/framework"
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
func findInteractiveClones(text string, settings InteractiveModeSettings) []framework.CloneGroup {
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

	// Fix to avoid memory explosion: 1) Count frequencies; 2) Collect positions for frequent windows only.
	length := settings.MinCloneLength
	if length <= 0 {
		length = 1
	}
	windowCount := tokenCount - length + 1
	if windowCount <= 0 {
		return nil
	}

	fmt.Printf("Processing windows of size %d (two-pass) ...\n", length)

	// frequency count
	freq := make(map[string]int, windowCount)
	for i := 0; i <= tokenCount-length; i++ {
		window := tokens[i : i+length]
		windowText := strings.Join(window, " ")
		freq[windowText]++
		if i%5000 == 0 && i > 0 {
			fmt.Printf("Counted %d/%d windows...\n", i, windowCount)
		}
	}

	// Prepare container only for candidates meeting min power
	potentialClones := make(map[string][]framework.TextFragment)
	candidates := 0
	for k, c := range freq {
		if c >= settings.MinGroupPower {
			potentialClones[k] = nil // mark as candidate
			candidates++
		}
	}
	fmt.Printf("Candidates meeting MinGroupPower=%d: %d (of %d uniques)\n", settings.MinGroupPower, candidates, len(freq))

	// collect positions only for candidates
	processed := 0
	for i := 0; i <= tokenCount-length; i++ {
		window := tokens[i : i+length]
		windowText := strings.Join(window, " ")
		if _, ok := potentialClones[windowText]; ok {
			potentialClones[windowText] = append(potentialClones[windowText], framework.TextFragment{
				Content:  windowText,
				StartPos: i,
				EndPos:   i + length,
			})
		}
		processed++
		if processed%5000 == 0 {
			fmt.Printf("Collected %d/%d windows...\n", processed, windowCount)
		}
	}

	fmt.Printf("Collected positions for %d candidate fragments\n", len(potentialClones))

	// Merge potential clones into groups using fuzzy similarity
	var groups []framework.CloneGroup
	for text, fragments := range potentialClones {
		// Try to find an existing group with similar archetype
		placed := false
		for gi := range groups {
			if isSimilarInteractive(text, groups[gi].Archetype) {
				groups[gi].Fragments = append(groups[gi].Fragments, fragments...)
				groups[gi].Power = len(groups[gi].Fragments)
				placed = true
				break
			}
		}
		if !placed {
			groups = append(groups, framework.CloneGroup{
				Fragments: append([]framework.TextFragment(nil), fragments...),
				Archetype: text,
				Power:     len(fragments),
			})
		}
	}

	// Apply standard interactive filtering
	groups = filterInteractiveGroups(groups, settings)

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
	// Use token-level Jaccard similarity with a slightly lower threshold for
	// interactive exploration
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
	return jaccard >= 0.8
}

// filterInteractiveGroups filters clone groups based on interactive mode settings
func filterInteractiveGroups(groups []framework.CloneGroup, settings InteractiveModeSettings) []framework.CloneGroup {
	var filtered []framework.CloneGroup

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
		var nonOverlapping []framework.TextFragment
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
func calculateArchetypes(groups *[]framework.CloneGroup) {
	for i := range *groups {
		group := &(*groups)[i]
		if len(group.Fragments) == 0 {
			continue
		}

		// Choose fragment with highest average Jaccard similarity to others
		bestIdx := 0
		bestScore := -1.0
		for idx, candidate := range group.Fragments {
			total := 0.0
			count := 0.0
			for j, other := range group.Fragments {
				if j == idx {
					continue
				}
				// Reuse interactive similarity as a proxy score
				// But compute raw Jaccard score to average
				aTokens := strings.Fields(candidate.Content)
				bTokens := strings.Fields(other.Content)
				aSet := make(map[string]struct{})
				for _, t := range aTokens {
					aSet[t] = struct{}{}
				}
				bSet := make(map[string]struct{})
				for _, t := range bTokens {
					bSet[t] = struct{}{}
				}
				inter := 0
				uni := 0
				seen := make(map[string]struct{})
				for t := range aSet {
					if _, ok := bSet[t]; ok {
						inter++
					}
					seen[t] = struct{}{}
					uni++
				}
				for t := range bSet {
					if _, ok := seen[t]; ok {
						continue
					}
					uni++
				}
				if uni > 0 {
					total += float64(inter) / float64(uni)
					count++
				}
			}
			score := 0.0
			if count > 0 {
				score = total / count
			}
			if score > bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		group.Archetype = group.Fragments[bestIdx].Content
	}
}

// ProcessInteractiveMode processes the text using interactive mode settings
func ProcessInteractiveMode(text string, settings InteractiveModeSettings) ([]framework.CloneGroup, error) {
	// Find clones
	groups := findInteractiveClones(text, settings)

	return groups, nil
}

// FormatInteractiveModeResults formats the analysis results for output
func FormatInteractiveModeResults(groups []framework.CloneGroup, settings InteractiveModeSettings) string {
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
