package main


// ApplyHeuristicRules применяет эвристические правила к n-граммам.
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

    // Генерация n-грамм из текста.
    ngrams := GenerateNGrams(text, n)

    // Применение эвристических правил.
    filteredNGrams := ApplyHeuristicRules(ngrams)
    return filteredNGrams
}