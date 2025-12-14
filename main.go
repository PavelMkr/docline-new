package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"Docline/framework"
)

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

// AnalysisResult represents HTTP response for analysis results
type AnalysisResult struct {
	Status      string              `json:"status"`
	Message     string              `json:"message"`
	Groups      map[string][]string `json:"groups,omitempty"`
	Archetypes  map[string]string   `json:"archetypes,omitempty"`
	ResultsFile string              `json:"results_file,omitempty"`
}

// setDefaultArchetypes ensures every group has Archetype; uses first fragment content if empty
func setDefaultArchetypes(groups []framework.CloneGroup) {
	for gi := range groups {
		if groups[gi].Archetype == "" && len(groups[gi].Fragments) > 0 {
			groups[gi].Archetype = groups[gi].Fragments[0].Content
		}
	}
}

func convertNGramResultsToGroups(ngramResults map[string][]string) []framework.CloneGroup {
	var groups []framework.CloneGroup
	for _, fragments := range ngramResults {
		group := framework.CloneGroup{
			Fragments: make([]framework.TextFragment, len(fragments)),
			Power:     len(fragments),
		}
		for i, frag := range fragments {
			fragTokens := strings.Fields(frag)
			start := findFirstTokenWindowIndex(fragTokens, fragTokens)
			end := start + len(fragTokens)
			group.Fragments[i] = framework.TextFragment{
				Content:  frag,
				StartPos: start,
				EndPos:   end,
			}
		}
		if len(group.Fragments) > 0 {
			group.Archetype = group.Fragments[0].Content
		}
		groups = append(groups, group)
	}
	return groups
}

// findFirstTokenWindowIndex returns the starting index of the first occurrence
// of needleTokens within hayTokens, or -1 if not found.
func findFirstTokenWindowIndex(hayTokens, needleTokens []string) int {
	if len(needleTokens) == 0 || len(hayTokens) < len(needleTokens) {
		return -1
	}
	lastStart := len(hayTokens) - len(needleTokens)
	for i := 0; i <= lastStart; i++ {
		match := true
		for j := 0; j < len(needleTokens); j++ {
			if hayTokens[i+j] != needleTokens[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// FormatAnalysisResults formats the analysis results for output
func FormatAnalysisResults(method string, groups []framework.CloneGroup, settings interface{}) string {
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
func ConvertGroupsToResponse(groups []framework.CloneGroup, useArchetypes bool) (map[string][]string, map[string]string) {
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

// generateResultsFileName create results file name based on input file and mode
func generateResultsFileName(inputPath, mode string) string {
	// Get file name without path and extension
	baseName := filepath.Base(inputPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]

	// Form new file name
	return fmt.Sprintf("%s_%s_results.html", nameWithoutExt, mode)
}

// getResultDir returns base results dir: ./results/<mode>/<filename-no-ext>
func getResultDir(inputPath, mode string) string {
	baseName := filepath.Base(inputPath)
	nameWithoutExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]
	return filepath.Join("./results", mode, nameWithoutExt)
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
		if os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("File not found: %s", filePath), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Error reading file '%s': %v", filePath, err), http.StatusInternalServerError)
		return
	}

	ngrams := HeuristicNgramAnalysis(data, text, 2)

	// Convert ngrams to clone groups format
	var groups []framework.CloneGroup
	for _, ngram := range ngrams {
		tokens := strings.Fields(ngram)
		group := framework.CloneGroup{
			Fragments: []framework.TextFragment{{
				Content:  ngram,
				StartPos: 0,
				EndPos:   len(tokens),
			}},
			Power: 1,
		}
		group.Archetype = ngram
		groups = append(groups, group)
	}
	setDefaultArchetypes(groups)

	// Format and save results text
	baseDir := getResultDir(filePath, "heuristic")
	_ = os.MkdirAll(baseDir, 0755)
	resultFilePath := filepath.Join(baseDir, generateResultsFileName(filePath, "heuristic"))
	// structured HTML
	heurSettings := fmt.Sprintf("<div><b>Extension Point Check:</b> %v</div>", data.ExtensionPointCheckbox)
	if err := WriteResultsHTML(resultFilePath, "Heuristic Analysis Results", groups, heurSettings, filePath); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Additional outputs per Heuristic Ngram Finder
	// heuristic_finder/out.json
	heurDir := filepath.Join(baseDir, "heuristic_finder")
	if err := os.MkdirAll(heurDir, 0755); err == nil {
		_ = writeJSON(filepath.Join(heurDir, "out.json"), map[string]any{
			"ngrams": groups,
		})
	}
	// <file>.neardups/pyvarelements.html
	base := filepath.Base(filePath)
	ndDir := filepath.Join(baseDir, base+".neardups")
	if err := os.MkdirAll(ndDir, 0755); err == nil {
		_ = WritePyVariativeElements(filepath.Join(ndDir, "pyvarelements.html"), groups)
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
	groups := convertNGramResultsToGroups(duplicates)
	setDefaultArchetypes(groups)
	for _, fragments := range duplicates {
		group := framework.CloneGroup{
			Fragments: make([]framework.TextFragment, len(fragments)),
			Power:     len(fragments),
		}
		for i, frag := range fragments {
			tokens := strings.Fields(frag)
			// Approximate the starting token index as 0 and compute length in tokens
			// If the caller passes the full token stream later, replace this with findFirstTokenWindowIndex
			//
			start := 0
			end := start + len(tokens)
			group.Fragments[i] = framework.TextFragment{
				Content:  frag,
				StartPos: start,
				EndPos:   end,
			}
		}
		groups = append(groups, group)
	}

	// Format and save results
	baseDir := getResultDir(filePath, "ngram")
	_ = os.MkdirAll(baseDir, 0755)
	resultFilePath := filepath.Join(baseDir, generateResultsFileName(filePath, "ngram"))
	ngramSettings := fmt.Sprintf("<div><b>Min clone length:</b> %d; <b>Max edit:</b> %d; <b>Max fuzzy:</b> %d; <b>Language:</b> %s</div>", data.MinCloneSlider, data.MaxEditSlider, data.MaxFuzzySlider, data.SourceLanguage)
	if err := WriteResultsHTML(resultFilePath, "N-Gram Analysis Results", groups, ngramSettings, filePath); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Additional outputs per Ngram Duplicate Finder
	// <file>.reformatted.result.txt (HTML now), <file>.reformatted.groups.json and pyvarelements.html
	base := filepath.Base(filePath)
	reform := filepath.Join(baseDir, base+".reformatted.result.html")
	reformText := FormatAnalysisResults("ngram", groups, data)
	_ = writeTextAsHTML(reform, "N-Gram Reformatted Result", reformText)
	groupsJSON := filepath.Join(baseDir, base+".reformatted.groups.json")
	_ = writeJSON(groupsJSON, groups)
	_ = WritePyVariativeElements(filepath.Join(baseDir, "pyvarelements.html"), groups)

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

	// Format and save results text
	baseDir := getResultDir(filePath, "automatic")
	_ = os.MkdirAll(baseDir, 0755)
	resultFilePath := filepath.Join(baseDir, generateResultsFileName(filePath, "automatic"))
	autoSettings := fmt.Sprintf("<div><b>Min clone length:</b> %d; <b>Convert to DRL:</b> %v; <b>Min archetype len:</b> %d; <b>Strict filter:</b> %v</div>", settings.MinCloneLength, settings.ConvertToDRL, settings.ArchetypeLength, settings.StrictFilter)
	if err := WriteResultsHTML(resultFilePath, "Automatic Mode Analysis Results", groups, autoSettings, filePath); err != nil {
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Console stats approximation
	fmt.Printf("=> Ntok, Total groups, E(cl. length)\n")
	fmt.Printf("%d, %5d, %.3f\n", settings.MinCloneLength, len(groups), AverageTokensInGroup(groups))

	// Generate Automatic mode outputs
	// Output reports under automatic/<file>/Output
	outDir := filepath.Join(baseDir, "Output")
	if err := os.MkdirAll(outDir, 0755); err == nil {
		_ = WritePygroupsHTML(filepath.Join(outDir, "pygroups.html"), groups, []string{filepath.Base(filePath)}, AverageTokensInGroup(groups), len(groups))
		_ = WritePyVariativeElements(filepath.Join(outDir, "pyvarelements.html"), groups)
		tokens := strings.Fields(content)
		_ = WriteDensityReports(outDir, len(tokens), groups)
		_ = WriteShortTermsCSV(filepath.Join(outDir, "shortterms.csv"), groups, 3, 2)
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
	if !settings.UseArchetype {
		setDefaultArchetypes(groups)
	}

	// Pre-generate heatmap HTML for server
	tokens := strings.Fields(content)
	preHTML := buildHeatmapHTML(len(tokens), groups)
	heatmapMu.Lock()
	heatmapHTML = preHTML
	heatmapMu.Unlock()

	// Prepare baseDir for interactive and save heatmap there
	baseDir := getResultDir(filePath, "interactive")
	_ = os.MkdirAll(baseDir, 0755)
	_ = writeSimpleHTML(filepath.Join(baseDir, "interactive_heatmap.html"), "Interactive Heatmap", preHTML)
	// Persist current interactive dir for /select endpoint
	interactiveDirMu.Lock()
	interactiveCurrentDir = baseDir
	interactiveDirMu.Unlock()

	// Format and save results
	resultFilePath := filepath.Join(baseDir, generateResultsFileName(filePath, "interactive"))
	fmt.Printf("Formatting results for file: %s\n", resultFilePath)
	interSettings := fmt.Sprintf("<div><b>Min clone:</b> %d; <b>Max clone:</b> %d; <b>Min group power:</b> %d; <b>Archetype:</b> %v</div>", settings.MinCloneLength, settings.MaxCloneLength, settings.MinGroupPower, settings.UseArchetype)
	fmt.Printf("Attempting to write results to file...\n")
	if err := WriteResultsHTML(resultFilePath, "Interactive Mode Analysis Results", groups, interSettings, filePath); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		http.Error(w, fmt.Sprintf("Error writing to file: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Successfully wrote results to file\n")

	// Convert groups to response format
	responseGroups, archetypes := ConvertGroupsToResponse(groups, settings.UseArchetype)

	// Start interactive heatmap server on first interactive request
	startInteractiveOnce.Do(func() { go startInteractiveHeatmapServer() })

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

// RegisterRoutes registers all HTTP routes on the given ServeMux
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/heuristic", heuristicFinderHandler)
	mux.HandleFunc("/ngram", ngramFinderHandler)
	mux.HandleFunc("/automatic", automaticModeHandler)
	mux.HandleFunc("/interactive", interactiveModeHandler)
}

var startInteractiveOnce sync.Once
var openInteractiveOnce sync.Once
var heatmapMu sync.RWMutex
var heatmapHTML string
var interactiveDirMu sync.RWMutex
var interactiveCurrentDir string

// buildHeatmapHTML builds a simple token-density heatmap based on groups
func buildHeatmapHTML(totalTokens int, groups []framework.CloneGroup) string {
	if totalTokens <= 0 {
		return "<p>No data to display</p>"
	}
	density := make([]int, totalTokens)
	maxv := 0
	for _, g := range groups {
		for _, f := range g.Fragments {
			b := f.StartPos
			e := f.EndPos
			if b < 0 {
				b = 0
			}
			if e > totalTokens {
				e = totalTokens
			}
			for i := b; i < e; i++ {
				density[i]++
				if density[i] > maxv {
					maxv = density[i]
				}
			}
		}
	}
	if maxv == 0 {
		maxv = 1
	}
	// bucketize into up to 600 bins to avoid huge DOM
	bins := 600
	if totalTokens < bins {
		bins = totalTokens
	}
	binSize := (totalTokens + bins - 1) / bins
	var sb strings.Builder
	sb.WriteString("<div style=\"font-family:sans-serif\"><div>Density heatmap (" + fmt.Sprintf("%d", totalTokens) + " tokens)</div>")
	sb.WriteString("<div style=\"white-space:nowrap;border:1px solid #ccc;height:24px\">")
	for i := 0; i < totalTokens; i += binSize {
		end := i + binSize
		if end > totalTokens {
			end = totalTokens
		}
		sum := 0
		for j := i; j < end; j++ {
			sum += density[j]
		}
		avg := float64(sum) / float64(end-i)
		alpha := avg / float64(maxv)
		if alpha > 1 {
			alpha = 1
		}
		widthPct := float64(end-i) / float64(totalTokens) * 100.0
		sb.WriteString(fmt.Sprintf("<span title=\"tokens %d-%d, avg=%.2f\" style=\"display:inline-block;height:24px;width:%.4f%%;background:rgba(255,0,0,%.3f)\"></span>", i, end, avg, widthPct, alpha))
	}
	sb.WriteString("</div>")
	sb.WriteString("<p style=\"margin-top:8px;color:#666\">Red intensity = clone density. Hover to see bin stats.</p></div>")
	return sb.String()
}

// startInteractiveHeatmapServer launches a minimal server at 127.0.0.1:49999
// providing a heatmap and an endpoint to generate near-duplicates html
func startInteractiveHeatmapServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		heatmapMu.RLock()
		htmlBody := heatmapHTML
		heatmapMu.RUnlock()
		if htmlBody == "" {
			htmlBody = "<p>No heatmap yet. Run interactive analysis to populate.</p>"
		}
		_ = writeSimpleHTMLToWriter(w, "Doc Clone Miner Heatmap", htmlBody)
	})
	mux.HandleFunc("/select", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// For simplicity, accept JSON {"fragments":["...","..."]}
		var body struct {
			Fragments []string `json:"fragments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		// produce pyvarelements.html in results/interactive/<file>/
		var groups []framework.CloneGroup
		if len(body.Fragments) > 0 {
			g := framework.CloneGroup{Power: len(body.Fragments)}
			for _, f := range body.Fragments {
				g.Fragments = append(g.Fragments, framework.TextFragment{Content: f})
			}
			groups = append(groups, g)
		}
		interactiveDirMu.RLock()
		outDir := interactiveCurrentDir
		interactiveDirMu.RUnlock()
		if outDir == "" {
			outDir = filepath.Join("./results", "interactive", "current")
		}
		_ = os.MkdirAll(outDir, 0755)
		target := filepath.Join(outDir, "pyvarelements.html")
		if err := WritePyVariativeElements(target, groups); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "report": target})
	})
	srv := &http.Server{Addr: "127.0.0.1:49999", Handler: mux}
	fmt.Println("Interactive heatmap server on http://127.0.0.1:49999/")
	// Attempt to open browser once when server starts the first time
	openInteractiveOnce.Do(func() { go openDefaultBrowser("http://127.0.0.1:49999/") })
	if err := srv.ListenAndServe(); err != nil {
		fmt.Println("interactive server stopped:", err)
	}
}

// writeSimpleHTMLToWriter mirrors writeSimpleHTML but writes to ResponseWriter
func writeSimpleHTMLToWriter(w http.ResponseWriter, title, bodyHTML string) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte("<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>" + title + "</title></head><body><h2>" + title + "</h2>" + bodyHTML + "</body></html>"))
	return err
}

// openDefaultBrowser tries to open the given URL in the user's default browser.
func openDefaultBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}
	_ = cmd.Start()
}

func main() {
	// CLI flags
	cliAuto := flag.Bool("cli-auto", false, "Run in automatic mode (CLI)")
	cliInter := flag.Bool("cli-interactive", false, "Run in interactive mode (CLI)")
	cliNGram := flag.Bool("cli-ngram", false, "Run in ngram duplicate mode (CLI)")
	cliHeur := flag.Bool("cli-heuristic", false, "Run in heuristic ngram mode (CLI)")

	input := flag.String("input", "", "Input file path")
	minClone := flag.Int("minClone", 20, "Minimal clone length (tokens)")
	maxClone := flag.Int("maxClone", 50, "Maximal clone length (tokens)")
	maxEdit := flag.Int("maxEdit", 9, "Maximal edit distance (Levenshtein)")
	maxDist := flag.Int("maxDist", 2, "Maximal fuzzy hash distance")
	minGroup := flag.Int("minGroup", 2, "Minimal Group Power (number of clones)")
	sourceLang := flag.String("source-language", "english", "Source document language")
	useArch := flag.Bool("use-archetype", false, "Archetype calculation")
	archetype := flag.Int("archetype", 5, "Minimal archetype length (tokens) [auto mode]")
	strict := flag.Bool("strict", true, "Strict filtering [auto mode]")
	convertToDRL := flag.Bool("drl", true, "Convert to DRL [auto mode]")
	extention := flag.Bool("extension", true, "Extension point values")
	flag.Parse()

	if *cliAuto || *cliInter || *cliNGram || *cliHeur {
		if *input == "" {
			fmt.Println("Error: --input is required")
			os.Exit(1)
		}
		if err := ensureResultsDir(); err != nil {
			fmt.Println("Failed to create results dir:", err)
			os.Exit(1)
		}
		if err := validateFileFormat(*input); err != nil {
			fmt.Println("File format error:", err)
			os.Exit(1)
		}
		content, err := readFileContent(*input)
		if err != nil {
			fmt.Println("Failed to read file:", err)
			os.Exit(1)
		}
		if *cliAuto {
			settings := AutomaticModeSettings{
				MinCloneLength:  *minClone,
				ConvertToDRL:    *convertToDRL,
				ArchetypeLength: *archetype,
				StrictFilter:    *strict,
				FilePath:        *input,
			}
			groups, err := ProcessAutomaticMode(content, settings)
			if err != nil {
				fmt.Println("Analysis error:", err)
				os.Exit(1)
			}
			baseDir := getResultDir(*input, "automatic")
			_ = os.MkdirAll(baseDir, 0755)
			resultFile := filepath.Join(baseDir, generateResultsFileName(*input, "automatic"))
			autoSettings := fmt.Sprintf("<div><b>Min clone length:</b> %d; <b>Convert to DRL:</b> %v; <b>Min archetype len:</b> %d; <b>Strict filter:</b> %v</div>", settings.MinCloneLength, settings.ConvertToDRL, settings.ArchetypeLength, settings.StrictFilter)
			if err := WriteResultsHTML(resultFile, "Automatic Mode Analysis Results", groups, autoSettings, *input); err != nil {
				fmt.Println("Failed to write result:", err)
				os.Exit(1)
			}
			fmt.Println("Analysis complete. Results saved to:", resultFile)
			return
		}
		if *cliInter {
			settings := InteractiveModeSettings{
				MinCloneLength: *minClone,
				MaxCloneLength: *maxClone,
				MinGroupPower:  *minGroup,
				UseArchetype:   *useArch,
				FilePath:       *input,
			}
			groups, err := ProcessInteractiveMode(content, settings)
			if err != nil {
				fmt.Println("Analysis error:", err)
				os.Exit(1)
			}
			if !settings.UseArchetype {
				setDefaultArchetypes(groups)
			}
			baseDir := getResultDir(*input, "interactive")
			_ = os.MkdirAll(baseDir, 0755)
			resultFile := filepath.Join(baseDir, generateResultsFileName(*input, "interactive"))
			interSettings := fmt.Sprintf("<div><b>Min clone:</b> %d; <b>Max clone:</b> %d; <b>Min group power:</b> %d; <b>Archetype:</b> %v</div>", settings.MinCloneLength, settings.MaxCloneLength, settings.MinGroupPower, settings.UseArchetype)
			if err := WriteResultsHTML(resultFile, "Interactive Mode Analysis Results", groups, interSettings, *input); err != nil {
				fmt.Println("Failed to write result:", err)
				os.Exit(1)
			}
			fmt.Println("Analysis complete. Results saved to:", resultFile)
			// Prepare heatmap before opening server
			tokens := strings.Fields(content)
			preHTML := buildHeatmapHTML(len(tokens), groups)
			heatmapMu.Lock()
			heatmapHTML = preHTML
			heatmapMu.Unlock()
			// Save interactive heatmap to results/interactive/<file>
			_ = writeSimpleHTML(filepath.Join(baseDir, "interactive_heatmap.html"), "Interactive Heatmap", preHTML)
			interactiveDirMu.Lock()
			interactiveCurrentDir = baseDir
			interactiveDirMu.Unlock()
			// Start interactive UI server and open browser, then block to keep it running
			go startInteractiveHeatmapServer()
			go openDefaultBrowser("http://127.0.0.1:49999/")
			fmt.Println("Interactive heatmap available at http://127.0.0.1:49999/ (press Ctrl+C to exit)")
			select {}
		}
		if *cliNGram {
			settings := NgramDuplicateFinderData{
				MinCloneSlider: *minClone,
				MaxEditSlider:  *maxDist,
				MaxFuzzySlider: *maxEdit,
				SourceLanguage: *sourceLang,
				FilePath:       *input,
			}
			parts := splitTextIntoParts(content)
			duplicates := FindDuplicatesByNGram(settings, parts)
			groups := convertNGramResultsToGroups(duplicates)
			baseDir := getResultDir(*input, "ngram")
			_ = os.MkdirAll(baseDir, 0755)
			resultFile := filepath.Join(baseDir, generateResultsFileName(*input, "ngram"))
			ngramSettings := fmt.Sprintf("<div><b>Min clone length:</b> %d; <b>Max edit:</b> %d; <b>Max fuzzy:</b> %d; <b>Language:</b> %s</div>", settings.MinCloneSlider, settings.MaxEditSlider, settings.MaxFuzzySlider, settings.SourceLanguage)
			if err := WriteResultsHTML(resultFile, "N-Gram Analysis Results", groups, ngramSettings, *input); err != nil {
				fmt.Println("Failed to write result:", err)
				os.Exit(1)
			}
			fmt.Println("Analysis complete. Results saved to:", resultFile)
			return
		}
		if *cliHeur {
			settings := HeuristicNgramFinderData{
				ExtensionPointCheckbox: *extention,
				FilePath:               *input,
			}
			ngrams := HeuristicNgramAnalysis(settings, content, 2)
			var groups []framework.CloneGroup
			for _, ngram := range ngrams {
				group := framework.CloneGroup{
					Fragments: []framework.TextFragment{{
						Content:  ngram,
						StartPos: 0, // TODO: Calculate actual positions
						EndPos:   1,
					}},
					Power: 1,
				}
				groups = append(groups, group)
			}
			baseDir := getResultDir(*input, "heuristic")
			_ = os.MkdirAll(baseDir, 0755)
			resultFile := filepath.Join(baseDir, generateResultsFileName(*input, "heuristic"))
			heurSettings := fmt.Sprintf("<div><b>Extension Point Check:</b> %v</div>", settings.ExtensionPointCheckbox)
			if err := WriteResultsHTML(resultFile, "Heuristic Analysis Results", groups, heurSettings, *input); err != nil {
				fmt.Println("Failed to write result:", err)
				os.Exit(1)
			}
			fmt.Println("Analysis complete. Results saved to:", resultFile)
			return
		}
	}

	// Start HTTP server in a goroutine
	go func() {
		mux := http.NewServeMux()
		RegisterRoutes(mux)
		fmt.Println("Starting server on :8080...")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	// Start UI after server
	fmt.Println("Starting UI...")
	startUI()
}
