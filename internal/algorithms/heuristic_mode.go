package internal

import (
	"math"
	"regexp"
	"strings"
	"unicode"
)



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
