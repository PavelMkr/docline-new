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

// Common structures for analysis results
type AnalysisResult struct {
	Status      string              `json:"status"`
	Message     string              `json:"message"`
	Groups      map[string][]string `json:"groups,omitempty"`
	Archetypes  map[string]string   `json:"archetypes,omitempty"`
	ResultsFile string              `json:"results_file,omitempty"`
}

// FormatAnalysisResults formats the analysis results for output
func FormatAnalysisResults(method string, groups []CloneGroup, settings interface{}) string {
	var sb strings.Builder

	// Write header based on method
	switch method {
	case "automatic":
		sb.WriteString("Automatic Mode Analysis Results\n")
		sb.WriteString("=============================\n\n")
		if s, ok := settings.(AutomaticModeSettings); ok {
			sb.WriteString(fmt.Sprintf("Settings:\n"))
			sb.WriteString(fmt.Sprintf("- Minimal Clone Length: %d tokens\n", s.MinCloneLength))
			sb.WriteString(fmt.Sprintf("- Convert to DRL: %v\n", s.ConvertToDRL))
			sb.WriteString(fmt.Sprintf("- Minimal Archetype Length: %d tokens\n", s.ArchetypeLength))
			sb.WriteString(fmt.Sprintf("- Strict Filtering: %v\n\n", s.StrictFilter))
		}
	case "interactive":
		sb.WriteString("Interactive Mode Analysis Results\n")
		sb.WriteString("===============================\n\n")
		if s, ok := settings.(InteractiveModeSettings); ok {
			sb.WriteString(fmt.Sprintf("Settings:\n"))
			sb.WriteString(fmt.Sprintf("- Minimal Clone Length: %d tokens\n", s.MinCloneLength))
			sb.WriteString(fmt.Sprintf("- Maximal Clone Length: %d tokens\n", s.MaxCloneLength))
			sb.WriteString(fmt.Sprintf("- Minimal Group Power: %d clones\n", s.MinGroupPower))
			sb.WriteString(fmt.Sprintf("- Archetype Calculation: %v\n\n", s.UseArchetype))
		}
	case "ngram":
		sb.WriteString("N-Gram Analysis Results\n")
		sb.WriteString("======================\n\n")
		if s, ok := settings.(NgramDuplicateFinderData); ok {
			sb.WriteString(fmt.Sprintf("Settings:\n"))
			sb.WriteString(fmt.Sprintf("- Minimal Clone Length: %d tokens\n", s.MinCloneSlider))
			sb.WriteString(fmt.Sprintf("- Max Edit Distance: %d\n", s.MaxEditSlider))
			sb.WriteString(fmt.Sprintf("- Max Fuzzy Hash Distance: %d\n", s.MaxFuzzySlider))
			sb.WriteString(fmt.Sprintf("- Source Language: %s\n\n", s.SourceLanguage))
		}
	case "heuristic":
		sb.WriteString("Heuristic Analysis Results\n")
		sb.WriteString("========================\n\n")
		if s, ok := settings.(HeuristicNgramFinderData); ok {
			sb.WriteString(fmt.Sprintf("Settings:\n"))
			sb.WriteString(fmt.Sprintf("- Extension Point Check: %v\n\n", s.ExtensionPointCheckbox))
		}
	}

	// Write groups
	sb.WriteString(fmt.Sprintf("Found %d clone groups:\n\n", len(groups)))
	for i, group := range groups {
		sb.WriteString(fmt.Sprintf("Group %d (Power: %d):\n", i+1, group.Power))
		if group.Archetype != "" {
			sb.WriteString(fmt.Sprintf("Archetype: %s\n", group.Archetype))
		}
		sb.WriteString("Fragments:\n")
		for j, frag := range group.Fragments {
			sb.WriteString(fmt.Sprintf("  %d. [%d-%d] %s\n", j+1, frag.StartPos, frag.EndPos, frag.Content))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ConvertGroupsToResponse converts clone groups to response format
func ConvertGroupsToResponse(groups []CloneGroup, useArchetypes bool) (map[string][]string, map[string]string) {
	responseGroups := make(map[string][]string)
	archetypes := make(map[string]string)

	for i, group := range groups {
		groupKey := fmt.Sprintf("group%d", i+1)
		var fragments []string
		for _, frag := range group.Fragments {
			fragments = append(fragments, frag.Content)
		}
		responseGroups[groupKey] = fragments
		if useArchetypes && group.Archetype != "" {
			archetypes[groupKey] = group.Archetype
		}
	}

	return responseGroups, archetypes
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
			fmt.Printf("Uploaded file: %s (original filename)\n", handler.Filename)
			filePath = filepath.Join("./results", handler.Filename)
			fmt.Printf("Full file path: %s\n", filePath)
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

	// check file format
	if err := validateFileFormat(filePath); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	text, err := readFileContent(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", filePath, err), http.StatusInternalServerError)
		return
	}

	ngrams := HeuristicNgramAnalysis(data, text, 2)

	// Convert ngrams to clone groups format
	var groups []CloneGroup
	for _, ngram := range ngrams {
		group := CloneGroup{
			Fragments: []TextFragment{{
				Content:  ngram,
				StartPos: 0, // TODO: Calculate actual positions
				EndPos:   1,
			}},
			Power: 1,
		}
		groups = append(groups, group)
	}

	// Format and save results
	resultFilePath := "./results/heuristic_results.txt"
	resultData := FormatAnalysisResults("heuristic", groups, data)
	if err := writeToFile(resultFilePath, resultData); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert groups to response format
	responseGroups, _ := ConvertGroupsToResponse(groups, false)

	// Clean up temporary file if needed
	if uploadedTempFile {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove temp file: %s, Err: %v\n", filePath, err)
		}
	}

	response := AnalysisResult{
		Status:      "success",
		Message:     "Heuristic analysis completed",
		Groups:      responseGroups,
		ResultsFile: resultFilePath,
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
			fmt.Printf("Uploaded file: %s (original filename)\n", handler.Filename)
			filePath = filepath.Join("./results", handler.Filename)
			fmt.Printf("Full file path: %s\n", filePath)
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

	// check file format
	if err := validateFileFormat(filePath); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content, err := readFileContent(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", filePath, err), http.StatusInternalServerError)
		return
	}

	parts := splitTextIntoParts(content)
	duplicates := FindDuplicatesByNGram(data, parts)

	// Convert duplicates to clone groups format
	var groups []CloneGroup
	for _, fragments := range duplicates {
		group := CloneGroup{
			Fragments: make([]TextFragment, len(fragments)),
			Power:     len(fragments),
		}
		for i, frag := range fragments {
			group.Fragments[i] = TextFragment{
				Content:  frag,
				StartPos: i, // TODO: Calculate actual positions
				EndPos:   i + 1,
			}
		}
		groups = append(groups, group)
	}

	// Format and save results
	resultFilePath := "./results/ngram_results.txt"
	resultData := FormatAnalysisResults("ngram", groups, data)
	if err := writeToFile(resultFilePath, resultData); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert groups to response format
	responseGroups, _ := ConvertGroupsToResponse(groups, false)

	// Clean up temporary file if needed
	if uploadedTempFile {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove temp file: %s, Err: %v\n", filePath, err)
		}
	}

	response := AnalysisResult{
		Status:      "success",
		Message:     "N-gram analysis completed",
		Groups:      responseGroups,
		ResultsFile: resultFilePath,
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

	// check file format
	if err := validateFileFormat(filePath); err != nil {
		// remove uploaded file if the format is not supported
		os.Remove(filePath)
		http.Error(w, err.Error(), http.StatusBadRequest)
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
			fmt.Printf("Uploaded file: %s (original filename)\n", handler.Filename)
			filePath = filepath.Join("./results", handler.Filename)
			fmt.Printf("Full file path: %s\n", filePath)
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

	// check file format
	if err := validateFileFormat(filePath); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	responseGroups, _ := ConvertGroupsToResponse(groups, false)

	// Clean up temporary file if needed
	if uploadedTempFile {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove temp file: %s, Err: %v\n", filePath, err)
		}
	}

	response := AnalysisResult{
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
			fmt.Printf("Uploaded file: %s (original filename)\n", handler.Filename)
			filePath = filepath.Join("./results", handler.Filename)
			fmt.Printf("Full file path: %s\n", filePath)
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

	// check file format
	if err := validateFileFormat(filePath); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Read and process file content
	content, err := readFileContent(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", filePath, err), http.StatusInternalServerError)
		return
	}

	// Process content using interactive mode
	fmt.Printf("Starting interactive mode processing with settings: %+v\n", settings)
	groups, err := ProcessInteractiveMode(content, settings)
	if err != nil {
		fmt.Printf("Error in ProcessInteractiveMode: %v\n", err)
		http.Error(w, fmt.Sprintf("Error processing content: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Interactive mode processing completed, found %d groups\n", len(groups))

	// Format and save results
	resultFilePath := "./results/interactive_results.txt"
	fmt.Printf("Formatting results for file: %s\n", resultFilePath)
	resultData := FormatInteractiveModeResults(groups, settings)
	fmt.Printf("Results formatted, data length: %d bytes\n", len(resultData))

	fmt.Printf("Attempting to write results to file...\n")
	if err := writeToFile(resultFilePath, resultData); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Successfully wrote results to file\n")

	// Convert groups to response format
	responseGroups, archetypes := ConvertGroupsToResponse(groups, settings.UseArchetype)

	// Clean up temporary file if needed
	if uploadedTempFile {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove temp file: %s, Err: %v\n", filePath, err)
		}
	}

	response := AnalysisResult{
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
