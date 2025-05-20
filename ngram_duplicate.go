package main

import "fmt"

func CalculateNGramSimilarity(map1, map2 map[string]int) float64 {
    intersection := 0
    union := 0

    allNGrams := make(map[string]bool)
    for ngram := range map1 {
        allNGrams[ngram] = true
    }
    for ngram := range map2 {
        allNGrams[ngram] = true
    }

    for ngram := range allNGrams {
        if map1[ngram] > 0 && map2[ngram] > 0 {
            intersection++
        }
        union++
    }

    if union == 0 {
        return 0
    }
    return float64(intersection) / float64(union)
}

// BuildNGramMap создает карту n-грамм для текста.
func BuildNGramMap(text string, n int) map[string]int {
    ngrams := GenerateNGrams(text, n)
    ngramMap := make(map[string]int)
    for _, ngram := range ngrams {
        ngramMap[ngram]++
    }
    return ngramMap
}

func FindDuplicatesByNGram(data NgramDuplicateFinderData, texts []string) map[string][]string {
    duplicates := make(map[string][]string)
    ngramMaps := make([]map[string]int, len(texts))

    // Используем MinCloneSlider как размер n-граммы.
    n := data.MinCloneSlider

    for i, text := range texts {
        ngramMaps[i] = BuildNGramMap(text, n)
    }

	for i := 0; i < len(texts); i++ {
		for j := i + 1; j < len(texts); j++ {
			similarity := CalculateNGramSimilarity(ngramMaps[i], ngramMaps[j])
			fmt.Printf("Similarity between text %d and text %d: %.2f\n", i, j, similarity)
	
			if similarity >= float64(data.MaxFuzzySlider)/100 {
				duplicates[texts[i]] = append(duplicates[texts[i]], texts[j])
			}
		}
	}

    return duplicates
}