package framework

import "strings"

// JaccardSimilarityCalculator implements SimilarityCalculator using Jaccard similarity.
// It is kept in the framework package because it does not depend on any of the
// algorithm or report implementation packages and can be reused by plugins.
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

// SpaceTokenizer implements TextTokenizer using space-based tokenization.
type SpaceTokenizer struct{}

func (s *SpaceTokenizer) Name() string {
	return "space"
}

func (s *SpaceTokenizer) Tokenize(text string) []string {
	return strings.Fields(text)
}

// StrictFilter implements Filter interface for strict filtering.
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

// hasOverlap checks if two text fragments overlap.
func hasOverlap(a, b TextFragment) bool {
	return a.StartPos <= b.EndPos && b.StartPos <= a.EndPos
}
