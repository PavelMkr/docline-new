package main

import (
    "strings"
    "os"
    "regexp"
)

// GenerateNGrams creates n-grams from input text.
func GenerateNGrams(text string, n int) []string {
    words := strings.Fields(text)
    var ngrams []string
    for i := 0; i <= len(words)-n; i++ {
        ngrams = append(ngrams, strings.Join(words[i:i+n], " "))
    }
    return ngrams
}

// writeToFile writes data to file.
func writeToFile(filePath string, data string) error {
    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = file.WriteString(data)
    if err != nil {
        return err
    }
    return nil
}

// readFileContent reads file content.
func readFileContent(filePath string) (string, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }
    return string(content), nil
}

// splitTextIntoParts splits text into parts (for example, by sentences).
func splitTextIntoParts(text string) []string {
    // Use regex to find punctuation marks.
    re := regexp.MustCompile(`[.!?]`)
    parts := re.Split(text, -1)

    // Remove empty lines and whitespace.
    var result []string
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part != "" {
            result = append(result, part)
        }
    }
    return result
}