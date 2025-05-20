package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	ui "github.com/webui-dev/go-webui/v2"
)

// HeuristicNgram
type HeuristicNgramFinderData struct {
	ExtensionPointCheckbox bool   `json:"extension_point_checkbox"`
	FilePath               string `json:"file_path"`
}

// NgramDuplicate
type NgramDuplicateFinderData struct {
	MinCloneSlider int    `json:"min_clone_slider"`
	MaxEditSlider  int    `json:"max_edit_slider"`
	MaxFuzzySlider int    `json:"max_fuzzy_slider"`
	SourceLanguage string `json:"source_language"`
	FilePath       string `json:"file_path"`
}

// FileUploadResponse represents the response on file upload
type FileUploadResponse struct {
	Status     string              `json:"status"`
	Message    string              `json:"message"`
	Duplicates map[string][]string `json:"duplicates,omitempty"`
}

// UploadSettings represents the analysis settings
type UploadSettings struct {
	MinCloneSlider int    `json:"min_clone_slider"`
	MaxEditSlider  int    `json:"max_edit_slider"`
	MaxFuzzySlider int    `json:"max_fuzzy_slider"`
	SourceLanguage string `json:"source_language"`
}

// ensureResultsDir ensures that the results directory exists
func ensureResultsDir() error {
	uploadDir := "./results"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %v", err)
	}
	return nil
}

func heuristicFinderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Ensure results directory exists
	if err := ensureResultsDir(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data HeuristicNgramFinderData

	// Try to parse as JSON first
	err := json.NewDecoder(r.Body).Decode(&data)
	if err == nil {
		// JSON parsed successfully
	} else {
		// If JSON parsing failed, try multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			http.Error(w, "Failed to parse request", http.StatusBadRequest)
			return
		}

		// Get settings from form
		settingsStr := r.FormValue("settings")
		if err := json.Unmarshal([]byte(settingsStr), &data); err != nil {
			http.Error(w, "Failed to parse settings", http.StatusBadRequest)
			return
		}

		// Get uploaded file if present
		file, handler, err := r.FormFile("file")
		if err == nil {
			defer file.Close()

			// Create file to save uploaded content
			filePath := filepath.Join("./results", handler.Filename)
			dst, err := os.Create(filePath)
			if err != nil {
				http.Error(w, "Failed to create file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// Copy uploaded file content
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}

			// Set the file path in the data
			data.FilePath = filePath
		}
	}

	// Check if file path is provided
	if data.FilePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// Read text from file
	text, err := readFileContent(data.FilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", data.FilePath, err), http.StatusInternalServerError)
		return
	}

	// Use heuristic analysis
	ngrams := HeuristicNgramAnalysis(data, text, 2)
	//fmt.Println("Heuristic N-Grams:", ngrams)

	// Save result to file
	resultFilePath := "./results/heuristic_results.txt"
	resultData := fmt.Sprintf("Heuristic N-Grams: %v\nAnalyzed file: %s", ngrams, data.FilePath)
	if err := writeToFile(resultFilePath, resultData); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Return results directly in the response
	response := map[string]interface{}{
		"status":        "success",
		"message":       "Heuristic finder processed",
		"results_file":  resultFilePath,
		"analyzed_file": data.FilePath,
		"ngrams":        ngrams,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ngramFinderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Ensure results directory exists
	if err := ensureResultsDir(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data NgramDuplicateFinderData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		// Try to parse as multipart form if JSON parsing failed
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Failed to parse request", http.StatusBadRequest)
			return
		}

		// Get settings from form
		settingsStr := r.FormValue("settings")
		if err := json.Unmarshal([]byte(settingsStr), &data); err != nil {
			http.Error(w, "Failed to parse settings", http.StatusBadRequest)
			return
		}

		// Get uploaded file if present
		file, handler, err := r.FormFile("file")
		if err == nil {
			defer file.Close()

			// Create file to save uploaded content
			filePath := filepath.Join("./results", handler.Filename)
			dst, err := os.Create(filePath)
			if err != nil {
				http.Error(w, "Failed to create file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// Copy uploaded file content
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}

			// Set the file path in the data
			data.FilePath = filePath
		}
	}

	// Check if file path is provided
	if data.FilePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// Read text from file
	content, err := readFileContent(data.FilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", data.FilePath, err), http.StatusInternalServerError)
		return
	}

	// Split text into parts
	parts := splitTextIntoParts(content)

	// Find duplicates
	duplicates := FindDuplicatesByNGram(data, parts)

	// Save result to file
	resultFilePath := "./results/ngram_duplicates.txt"
	resultData := fmt.Sprintf("Duplicates Found: %v", duplicates)
	if err := writeToFile(resultFilePath, resultData); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Return results directly in the response
	response := map[string]interface{}{
		"status":       "success",
		"message":      "Ngram finder processed",
		"results_file": resultFilePath,
		"duplicates":   duplicates,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Получаем загруженный файл
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Создаем временную директорию для загруженных файлов, если её нет
	uploadDir := "./results"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	// Создаем файл для сохранения загруженного содержимого
	filePath := filepath.Join(uploadDir, handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Копируем содержимое загруженного файла
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Парсим настройки
	var settings UploadSettings
	settingsStr := r.FormValue("settings")
	if err := json.Unmarshal([]byte(settingsStr), &settings); err != nil {
		http.Error(w, "Failed to parse settings", http.StatusBadRequest)
		return
	}

	// Читаем содержимое файла
	content, err := readFileContent(filePath)
	if err != nil {
		http.Error(w, "Failed to read file content", http.StatusInternalServerError)
		return
	}

	// Разделяем текст на части
	parts := splitTextIntoParts(content)

	// Ищем дубликаты
	data := NgramDuplicateFinderData{
		MinCloneSlider: settings.MinCloneSlider,
		MaxEditSlider:  settings.MaxEditSlider,
		MaxFuzzySlider: settings.MaxFuzzySlider,
		SourceLanguage: settings.SourceLanguage,
		FilePath:       filePath,
	}
	duplicates := FindDuplicatesByNGram(data, parts)

	// Формируем ответ
	response := FileUploadResponse{
		Status:     "success",
		Message:    "File processed successfully",
		Duplicates: duplicates,
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Start HTTP server in a goroutine
	go func() {
		http.HandleFunc("/upload", uploadHandler)
		http.HandleFunc("/heuristic", heuristicFinderHandler)
		http.HandleFunc("/ngram", ngramFinderHandler)

		fmt.Println("Starting server on :8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	// UI
	// Create a window.
	w := ui.NewWindow()
	// Show frontend.
	w.Show("index.html")
	// Wait until all windows get closed.
	ui.Wait()
}
