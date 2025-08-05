// FIXME find every 2 words in text
package main

// HeuristicNgram
type HeuristicNgramFinderData struct {
	ExtensionPointCheckbox bool   `json:"extension_point_checkbox"`
	FilePath               string `json:"file_path"`
}

// applies heuristic rules to n-grams.
func ApplyHeuristicRules(ngrams []string) []string {
    unique := make(map[string]bool)
    var filtered []string

    for _, ngram := range ngrams {
        if !containsNumbers(ngram) && !unique[ngram] {
            filtered = append(filtered, ngram)
            unique[ngram] = true
        }
    }
    return filtered
}

// check numbers in string
func containsNumbers(s string) bool {
    for _, char := range s {
        if char >= '0' && char <= '9' {
            return true
        }
    }
    return false
}

// start analyzis
func HeuristicNgramAnalysis(data HeuristicNgramFinderData, text string, n int) []string {
    if !data.ExtensionPointCheckbox {
        return nil // if false - dont analyze
    }

    // Generate n-grams from text.
    ngrams := GenerateNGrams(text, n)

    // Applying heuristic rules.
    filteredNGrams := ApplyHeuristicRules(ngrams)
    return filteredNGrams
}