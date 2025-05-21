package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	fmt.Println("Get request:", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := ensureResultsDir(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data HeuristicNgramFinderData
	var filePath string
	var uploadedTempFile bool = false

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		filePath = data.FilePath
		fmt.Println("Use file path from JSON:", filePath)
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}
		settingsStr := r.FormValue("settings")
		fmt.Println("Settings:", settingsStr)
		if err := json.Unmarshal([]byte(settingsStr), &data); err != nil {
			http.Error(w, "Failed to parse settings", http.StatusBadRequest)
			return
		}
		file, handler, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			filePath = filepath.Join("./results", handler.Filename)
			dst, err := os.Create(filePath)
			if err != nil {
				http.Error(w, "Failed to create file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
			fmt.Println("File uploaded:", filePath)
			uploadedTempFile = true
		} else {
			fmt.Println("Error getting file:", err)
			filePath = data.FilePath
		}
	} else {
		http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
		return
	}

	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	text, err := readFileContent(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", filePath, err), http.StatusInternalServerError)
		return
	}

	ngrams := HeuristicNgramAnalysis(data, text, 2)

	resultFilePath := "./results/heuristic_results.txt"
	resultData := fmt.Sprintf("Heuristic N-Grams: %v", ngrams)
	if err := writeToFile(resultFilePath, resultData); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// if multipart form, remove temp file
	if uploadedTempFile {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove temp file: %s, Err: %v\n", filePath, err)
		}
	}

	response := map[string]interface{}{
		"status":        "success",
		"message":       "Heuristic finder processed",
		"results_file":  resultFilePath,
		"analyzed_file": filePath,
		"ngrams":        ngrams,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ngramFinderHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get request:", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := ensureResultsDir(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data NgramDuplicateFinderData
	var filePath string
	var uploadedTempFile bool = false

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		filePath = data.FilePath
		fmt.Println("Use file path from JSON:", filePath)
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}
		settingsStr := r.FormValue("settings")
		fmt.Println("Settings:", settingsStr)
		if err := json.Unmarshal([]byte(settingsStr), &data); err != nil {
			http.Error(w, "Failed to parse settings", http.StatusBadRequest)
			return
		}
		file, handler, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			filePath = filepath.Join("./results", handler.Filename)
			dst, err := os.Create(filePath)
			if err != nil {
				http.Error(w, "Failed to create file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
			fmt.Println("File uploaded:", filePath)
			uploadedTempFile = true
		} else {
			fmt.Println("Error getting file:", err)
			filePath = data.FilePath
		}
	} else {
		http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
		return
	}

	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	content, err := readFileContent(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", filePath, err), http.StatusInternalServerError)
		return
	}

	parts := splitTextIntoParts(content)
	duplicates := FindDuplicatesByNGram(data, parts)
	resultFilePath := "./results/ngram_duplicates.txt"
	resultData := fmt.Sprintf("Duplicates Found: %v", duplicates)
	if err := writeToFile(resultFilePath, resultData); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// if multipart form, remove temp file
	if uploadedTempFile {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove temp file: %s, Err: %v\n", filePath, err)
		}
	}

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

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create temporary directory for uploaded files if it doesn't exist
	uploadDir := "./results"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	// Create file to save uploaded content
	filePath := filepath.Join(uploadDir, handler.Filename)
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

	// Parse settings
	var settings UploadSettings
	settingsStr := r.FormValue("settings")
	fmt.Println("Settings:", settingsStr)
	if err := json.Unmarshal([]byte(settingsStr), &settings); err != nil {
		http.Error(w, "Failed to parse settings", http.StatusBadRequest)
		return
	}

	// Read file content
	content, err := readFileContent(filePath)
	if err != nil {
		http.Error(w, "Failed to read file content", http.StatusInternalServerError)
		return
	}

	// Split text into parts
	parts := splitTextIntoParts(content)

	// Find duplicates
	data := NgramDuplicateFinderData{
		MinCloneSlider: settings.MinCloneSlider,
		MaxEditSlider:  settings.MaxEditSlider,
		MaxFuzzySlider: settings.MaxFuzzySlider,
		SourceLanguage: settings.SourceLanguage,
		FilePath:       filePath,
	}
	fmt.Printf("data: %+v\n", data)
	duplicates := FindDuplicatesByNGram(data, parts)

	// Form response
	response := FileUploadResponse{
		Status:     "success",
		Message:    "File processed successfully",
		Duplicates: duplicates,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func automaticModeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get request:", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := ensureResultsDir(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var settings AutomaticModeSettings
	var filePath string
	var uploadedTempFile bool = false

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err := json.NewDecoder(r.Body).Decode(&settings)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		filePath = settings.FilePath
		fmt.Println("Using file path from JSON:", filePath)
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		settingsStr := r.FormValue("settings")
		fmt.Println("Settings:", settingsStr)
		if err := json.Unmarshal([]byte(settingsStr), &settings); err != nil {
			http.Error(w, "Failed to parse settings", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			filePath = filepath.Join("./results", handler.Filename)
			dst, err := os.Create(filePath)
			if err != nil {
				http.Error(w, "Failed to create file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
			fmt.Println("File uploaded:", filePath)
			uploadedTempFile = true
		} else {
			fmt.Println("Error getting file:", err)
			return
		}
	} else {
		http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
		return
	}

	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// Read and process file content
	content, err := readFileContent(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", filePath, err), http.StatusInternalServerError)
		return
	}

	// Process content using automatic mode
	groups, err := ProcessAutomaticMode(content, settings)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing content: %v", err), http.StatusInternalServerError)
		return
	}

	// Format and save results
	resultFilePath := "./results/automatic_results.txt"
	resultData := FormatAutomaticModeResults(groups, settings)
	if err := writeToFile(resultFilePath, resultData); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert groups to response format
	responseGroups := make(map[string][]string)
	for i, group := range groups {
		groupKey := fmt.Sprintf("group%d", i+1)
		var fragments []string
		for _, frag := range group.Fragments {
			fragments = append(fragments, frag.Content)
		}
		responseGroups[groupKey] = fragments
	}

	// Clean up temporary file if needed
	if uploadedTempFile {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove temp file: %s, Err: %v\n", filePath, err)
		}
	}

	response := AutomaticModeResponse{
		Status:      "success",
		Message:     "Automatic mode analysis completed",
		Groups:      responseGroups,
		ResultsFile: resultFilePath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func interactiveModeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get request:", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := ensureResultsDir(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var settings InteractiveModeSettings
	var filePath string
	var uploadedTempFile bool = false

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err := json.NewDecoder(r.Body).Decode(&settings)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		filePath = settings.FilePath
		fmt.Println("Using file path from JSON:", filePath)
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		settingsStr := r.FormValue("settings")
		fmt.Println("Settings:", settingsStr)
		if err := json.Unmarshal([]byte(settingsStr), &settings); err != nil {
			http.Error(w, "Failed to parse settings", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			filePath = filepath.Join("./results", handler.Filename)
			dst, err := os.Create(filePath)
			if err != nil {
				http.Error(w, "Failed to create file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
			fmt.Println("File uploaded:", filePath)
			uploadedTempFile = true
		} else {
			fmt.Println("Error getting file:", err)
			return
		}
	} else {
		http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
		return
	}

	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// Read and process file content
	content, err := readFileContent(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", filePath, err), http.StatusInternalServerError)
		return
	}

	// Process content using interactive mode
	groups, err := ProcessInteractiveMode(content, settings)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing content: %v", err), http.StatusInternalServerError)
		return
	}

	// Format and save results
	resultFilePath := "./results/interactive_results.txt"
	resultData := FormatInteractiveModeResults(groups, settings)
	if err := writeToFile(resultFilePath, resultData); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert groups to response format
	responseGroups := make(map[string][]string)
	archetypes := make(map[string]string)
	for i, group := range groups {
		groupKey := fmt.Sprintf("group%d", i+1)
		var fragments []string
		for _, frag := range group.Fragments {
			fragments = append(fragments, frag.Content)
		}
		responseGroups[groupKey] = fragments
		if settings.UseArchetype {
			archetypes[groupKey] = group.Archetype
		}
	}

	// Clean up temporary file if needed
	if uploadedTempFile {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove temp file: %s, Err: %v\n", filePath, err)
		}
	}

	response := InteractiveModeResponse{
		Status:      "success",
		Message:     "Interactive mode analysis completed",
		Groups:      responseGroups,
		Archetypes:  archetypes,
		ResultsFile: resultFilePath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func SendSettings(e ui.Event) string {
	// json string
	jsonStr, err := ui.GetArg[string](e)
	if err != nil {
		return fmt.Sprintf("Error getting argument: %v", err)
	}

	// json to map
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &settings); err != nil {
		return fmt.Sprintf("Error parsing JSON: %v", err)
	}

	// print data
	fmt.Printf("Received settings: %+v\n", settings)

	// FIXME unnessesarry return
	return "Settings received and parsed successfully"
}

func main() {
	// Start HTTP server in a goroutine
	go func() {
		http.HandleFunc("/upload", uploadHandler)
		http.HandleFunc("/heuristic", heuristicFinderHandler)
		http.HandleFunc("/ngram", ngramFinderHandler)
		http.HandleFunc("/automatic", automaticModeHandler)
		http.HandleFunc("/interactive", interactiveModeHandler)

		fmt.Println("Starting server on :8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	// UI
	// Create a window.
	w := ui.NewWindow()
	// Bind a Go function.
	ui.Bind(w, "SendSettings", SendSettings)
	// Show frontend.
	w.Show("index.html")
	// Wait until all windows get closed.
	ui.Wait()
}
