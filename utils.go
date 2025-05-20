package main

import (
    "strings"
    "os"
    "regexp"
)

// GenerateNGrams создает n-граммы из входного текста.
func GenerateNGrams(text string, n int) []string {
    words := strings.Fields(text)
    var ngrams []string
    for i := 0; i <= len(words)-n; i++ {
        ngrams = append(ngrams, strings.Join(words[i:i+n], " "))
    }
    return ngrams
}

// writeToFile записывает данные в файл.
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

// readFileContent читает содержимое файла.
func readFileContent(filePath string) (string, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }
    return string(content), nil
}

// splitTextIntoParts разделяет текст на части (например, по предложениям).
func splitTextIntoParts(text string) []string {
    // Используем регулярное выражение для поиска знаков препинания.
    re := regexp.MustCompile(`[.!?]`)
    parts := re.Split(text, -1)

    // Удаляем пустые строки и пробельные символы.
    var result []string
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part != "" {
            result = append(result, part)
        }
    }
    return result
}