package internal

import (
	"runtime"
	"strings"
	"sync"
)

// GenerateNGrams returns all word-level n-grams (joined with spaces). Capacity is
// pre-reserved to avoid repeated slice growth.
func GenerateNGrams(text string, n int) []string {
	words := strings.Fields(text)
	if n < 1 || len(words) < n {
		return nil
	}
	win := len(words) - n + 1
	out := make([]string, 0, win)
	for i := 0; i < win; i++ {
		out = append(out, strings.Join(words[i:i+n], " "))
	}
	return out
}

// NgramDuplicate
type NgramDuplicateFinderData struct {
	MinCloneSlider int `json:"min_clone_slider"`
	MaxEditSlider  int `json:"max_edit_slider"`
	MaxFuzzySlider int `json:"max_fuzzy_slider"`
	// SourceLanguage string `json:"source_language"`
	FilePath string `json:"file_path"`
}

// CalculateNGramSimilarity returns Jaccard similarity over the *sets* of n-gram types
// (keys with positive count): |M1 ∩ M2| / |M1 ∪ M2|.
func CalculateNGramSimilarity(map1, map2 map[string]int) float64 {
	if len(map1) == 0 && len(map2) == 0 {
		return 0
	}

	intersection := 0
	for k := range map1 {
		if map2[k] > 0 {
			intersection++
		}
	}
	// |M1 ∪ M2| = |M1| + |{k ∈ M2 : k ∉ M1}|
	union := len(map1)
	for k := range map2 {
		if map1[k] == 0 {
			union++
		}
	}
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// BuildNGramMap creates a multiset map of n-grams (word windows) for text.
// One pass, no intermediate slice of all n-gram strings.
func BuildNGramMap(text string, n int) map[string]int {
	return BuildNGramMapFromWords(strings.Fields(text), n)
}

// BuildNGramMapFromWords counts n-grams from an already tokenized word slice
// (avoids repeated Fields when the same text is processed elsewhere).
func BuildNGramMapFromWords(words []string, n int) map[string]int {
	if n < 1 || len(words) < n {
		return map[string]int{}
	}
	win := len(words) - n + 1
	m := make(map[string]int, win)
	for i := 0; i < win; i++ {
		ng := strings.Join(words[i:i+n], " ")
		m[ng]++
	}
	return m
}

// ngramParallelPairThreshold: below this many pairs, sequential comparison is cheaper
// than goroutine scheduling overhead.
const ngramParallelPairThreshold = 256

func FindDuplicatesByNGram(data NgramDuplicateFinderData, texts []string) map[string][]string {
	duplicates := make(map[string][]string)
	T := len(texts)
	if T < 2 {
		return duplicates
	}

	n := data.MinCloneSlider
	if n < 1 {
		n = 1
	}

	wordsPer := make([][]string, T)
	for i, t := range texts {
		wordsPer[i] = strings.Fields(t)
	}

	ngramMaps := make([]map[string]int, T)
	for i := range texts {
		ngramMaps[i] = BuildNGramMapFromWords(wordsPer[i], n)
	}

	threshold := float64(data.MaxFuzzySlider) / 100
	pairCount := T * (T - 1) / 2

	if pairCount < ngramParallelPairThreshold {
		for i := 0; i < T; i++ {
			for j := i + 1; j < T; j++ {
				if CalculateNGramSimilarity(ngramMaps[i], ngramMaps[j]) >= threshold {
					duplicates[texts[i]] = append(duplicates[texts[i]], texts[j])
				}
			}
		}
		return duplicates
	}

	// Parallel path: compute all pairwise similarities, then merge in fixed (i,j) order.
	sims := make([]float64, pairCount)
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	if workers > pairCount {
		workers = pairCount
	}
	chunk := (pairCount + workers - 1) / workers

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		start := w * chunk
		if start >= pairCount {
			break
		}
		end := min(start+chunk, pairCount)
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for idx := start; idx < end; idx++ {
				i, j := ngramPairIndexToIJ(T, idx)
				sims[idx] = CalculateNGramSimilarity(ngramMaps[i], ngramMaps[j])
			}
		}(start, end)
	}
	wg.Wait()

	for idx := 0; idx < pairCount; idx++ {
		if sims[idx] < threshold {
			continue
		}
		i, j := ngramPairIndexToIJ(T, idx)
		duplicates[texts[i]] = append(duplicates[texts[i]], texts[j])
	}

	return duplicates
}

// ngramPairIndexToIJ maps a linear pair index in row-major upper triangle to (i, j), i < j.
// Order matches nested loops: (0,1),(0,2),…,(0,T-1),(1,2),…
func ngramPairIndexToIJ(T, idx int) (i, j int) {
	// Largest i such that rowStart(i) <= idx.
	lo, hi := 0, T-2
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if ngramPairRowStart(T, mid) <= idx {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	i = lo
	j = i + 1 + (idx - ngramPairRowStart(T, i))
	return i, j
}

// ngramPairRowStart returns the global index of pair (i, i+1).
func ngramPairRowStart(T, i int) int {
	if i <= 0 {
		return 0
	}
	return i*(T-1) - i*(i-1)/2
}
