package internal

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/PavelMkr/docline-new/internal/framework"
)

type HeuristicModeAdapter struct{}

func (a *HeuristicModeAdapter) Name() string {
	return "heuristic"
}

func (a *HeuristicModeAdapter) Description() string {
	return "Heuristic NLP duplicate finder based on token overlap"
}

func (a *HeuristicModeAdapter) FindClones(text string, cfg framework.CloneFinderConfig) ([]framework.CloneGroup, error) {
	segments := splitTextIntoSentenceSegments(text)
	if len(segments) == 0 {
		return nil, nil
	}

	minCloneLength := defaultInt(cfg.MinCloneLength, 4)
	threshold := getFloat(cfg.CustomParams, "similarity_threshold", 0.72)
	if threshold <= 0 {
		threshold = 0.72
	}
	if threshold > 1 {
		threshold = 1
	}

	parent := make([]int, len(segments))
	for i := range parent {
		parent[i] = i
	}

	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			sim := jaccardSimilarity(segments[i].tokenFreq, segments[j].tokenFreq)
			if sim >= threshold {
				union(parent, i, j)
			}
		}
	}

	components := map[int][]int{}
	for i := range segments {
		root := find(parent, i)
		components[root] = append(components[root], i)
	}

	groups := make([]framework.CloneGroup, 0, len(components))
	for _, idxs := range components {
		if len(idxs) < maxInt(defaultInt(cfg.MinGroupPower, 2), 2) {
			continue
		}

		sort.Slice(idxs, func(i, j int) bool {
			return segments[idxs[i]].startToken < segments[idxs[j]].startToken
		})

		fragments := make([]framework.TextFragment, 0, len(idxs))
		for _, idx := range idxs {
			seg := segments[idx]
			if seg.tokenCount < minCloneLength {
				continue
			}
			fragments = append(fragments, framework.TextFragment{
				Content:  seg.text,
				StartPos: seg.startToken,
				EndPos:   seg.endToken,
			})
		}

		if len(fragments) < maxInt(defaultInt(cfg.MinGroupPower, 2), 2) {
			continue
		}

		groups = append(groups, framework.CloneGroup{
			Fragments: fragments,
			Power:     len(fragments),
			Archetype: fragments[0].Content,
			Metadata: map[string]interface{}{
				"finder":               "heuristic-nlp",
				"similarity_threshold": threshold,
			},
		})
	}

	return groups, nil
}

type sentenceSegment struct {
	text       string
	startToken int
	endToken   int
	tokenCount int
	tokenFreq  map[string]int
}

func splitTextIntoSentenceSegments(text string) []sentenceSegment {
	re := regexp.MustCompile(`[^.!?\n]+`)
	matches := re.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	segments := make([]sentenceSegment, 0, len(matches))
	searchFrom := 0
	for _, raw := range matches {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		localIdx := strings.Index(text[searchFrom:], raw)
		if localIdx < 0 {
			continue
		}
		absIdx := searchFrom + localIdx
		searchFrom = absIdx + len(raw)

		startToken := len(strings.Fields(text[:absIdx]))
		tokenCount := len(strings.Fields(trimmed))
		if tokenCount == 0 {
			continue
		}

		segments = append(segments, sentenceSegment{
			text:       trimmed,
			startToken: startToken,
			endToken:   startToken + tokenCount,
			tokenCount: tokenCount,
			tokenFreq:  normalizedTokenFrequency(trimmed),
		})
	}

	return segments
}

func normalizedTokenFrequency(s string) map[string]int {
	freq := map[string]int{}
	for _, tok := range strings.Fields(strings.ToLower(s)) {
		clean := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return -1
		}, tok)
		if len(clean) < 2 {
			continue
		}
		freq[clean]++
	}
	return freq
}

func jaccardSimilarity(a, b map[string]int) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	union := 0
	seen := map[string]struct{}{}

	for token, ac := range a {
		bc := b[token]
		inter += minInt(ac, bc)
		union += maxInt(ac, bc)
		seen[token] = struct{}{}
	}
	for token, bc := range b {
		if _, ok := seen[token]; ok {
			continue
		}
		union += bc
	}
	if union == 0 {
		return 0
	}
	return math.Min(1, float64(inter)/float64(union))
}

func find(parent []int, x int) int {
	if parent[x] != x {
		parent[x] = find(parent, parent[x])
	}
	return parent[x]
}

func union(parent []int, a, b int) {
	ra := find(parent, a)
	rb := find(parent, b)
	if ra != rb {
		parent[rb] = ra
	}
}

func getFloat(m map[string]interface{}, key string, def float64) float64 {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		switch vv := v.(type) {
		case float64:
			return vv
		case float32:
			return float64(vv)
		case int:
			return float64(vv)
		case int32:
			return float64(vv)
		case int64:
			return float64(vv)
		}
	}
	return def
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
