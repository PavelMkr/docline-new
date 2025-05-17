package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	ui "github.com/webui-dev/go-webui/v2"
)

type PageData struct {
	Title string
}

// Automatic Mode Data
type AutomaticModeData struct {
	LengthSlider            int  `json:"length_slider"`
	ConvertCheckbox         bool `json:"convert_checkbox"`
	ArchetypeSlider         int  `json:"archetype_slider"`
	StrictFilteringCheckbox bool `json:"strict_filtering_checkbox"`
}

// Interactive Mode Data
type InteractiveModeData struct {
	MinCloneSlider    int  `json:"min_clone_slider"`
	MaxCloneSlider    int  `json:"max_clone_slider"`
	MinGroupSlider    int  `json:"min_group_slider"`
	ExtensionCheckbox bool `json:"extension_checkbox"`
}

// Ngram Duplicate Finder Data
type NgramDuplicateFinderData struct {
	MinCloneSlider int    `json:"min_clone_slider"`
	MaxEditSlider  int    `json:"max_edit_slider"`
	MaxFuzzySlider int    `json:"max_fuzzy_slider"`
	SourceLanguage string `json:"source_language"`
}

// Heuristic Ngram Finder Data
type HeuristicNgramFinderData struct {
	ExtensionPointCheckbox bool `json:"extension_point_checkbox"`
}

// Structure for analysis settings
type AnalysisSettings struct {
	MinCloneSlider int    `json:"min_clone_slider"`
	MaxEditSlider  int    `json:"max_edit_slider"`
	MaxFuzzySlider int    `json:"max_fuzzy_slider"`
	SourceLanguage string `json:"source_language"`
}

// Structure for storing information about duplicates
type Duplicate struct {
	Text      string `json:"text"`
	Count     int    `json:"count"`
	Positions []int  `json:"positions"`
}

//
// func
//
func analyzeDuplicates(content string, settings AnalysisSettings) ([]Duplicate, error) {
	// Split text into words
	words := strings.Fields(content)

	// Create a map for counting duplicates
	wordCount := make(map[string]*Duplicate)

	// Analyze each word
	for pos, word := range words {
		// Normalize word (convert to lowercase and remove punctuation)
		word = strings.ToLower(strings.Trim(word, ".,!?()[]{}\"':;"))

		// Skip words shorter than the minimum length
		if len(word) < settings.MinCloneSlider {
			continue
		}

		// Add or update information about duplicates
		if dup, exists := wordCount[word]; exists {
			dup.Count++
			dup.Positions = append(dup.Positions, pos)
		} else {
			wordCount[word] = &Duplicate{
				Text:      word,
				Count:     1,
				Positions: []int{pos},
			}
		}
	}

	// Form a list of duplicates
	var duplicates []Duplicate
	for _, dup := range wordCount {
		if dup.Count > 1 { // Add only if the word occurs more than once
			duplicates = append(duplicates, *dup)
		}
	}

	// Sort by the number of repetitions (desc)
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].Count > duplicates[j].Count
	})

	return duplicates, nil
}

//
// Handlers
//
func automaticModeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var data AutomaticModeData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("Automatic Mode Data Received:", data)
	response := map[string]string{"status": "success", "message": "Automatic mode processed"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func interactiveModeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var data InteractiveModeData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("Interactive Mode Data Received:", data)
	response := map[string]string{"status": "success", "message": "Interactive mode processed"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ngramFinderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var data NgramDuplicateFinderData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("Ngram Finder Data Received:", data)
	response := map[string]string{"status": "success", "message": "Ngram finder processed"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func heuristicFinderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var data HeuristicNgramFinderData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("Heuristic Finder Data Received:", data)
	response := map[string]string{"status": "success", "message": "Heuristic finder processed"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// File size limit (10 MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file type
	// FIXME: add more file types
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".txt":  true,
		".md":   true,
		".doc":  true,
		".docx": true,
	}

	if !allowedExts[ext] {
		http.Error(w, "Invalid file type. Allowed types: .txt, .md, .doc, .docx", http.StatusBadRequest)
		return
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Get analysis settings
	settingsStr := r.FormValue("settings")
	var settings AnalysisSettings
	if err := json.Unmarshal([]byte(settingsStr), &settings); err != nil {
		http.Error(w, "Invalid settings format", http.StatusBadRequest)
		return
	}

	// Check if settings are valid
	if settings.MinCloneSlider < 1 {
		settings.MinCloneSlider = 1 // Set minimum value
	}

	// Analyze text for duplicates
	duplicates, err := analyzeDuplicates(string(content), settings)
	if err != nil {
		http.Error(w, "Error analyzing duplicates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send result
	response := map[string]interface{}{
		"status":   "success",
		"result":   duplicates,
		"filename": header.Filename,
		"filesize": header.Size,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}


//
// main
//
func main() {
	// Parse HTML templates
	templates := template.Must(template.ParseFiles("index.html"))

	// Serve static files (CSS, JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{Title: "Mode Selector"}
		if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	http.HandleFunc("/automatic_mode", automaticModeHandler)
	http.HandleFunc("/interactive_mode", interactiveModeHandler)
	http.HandleFunc("/ngram_finder", ngramFinderHandler)
	http.HandleFunc("/heuristic_finder", heuristicFinderHandler)
	http.HandleFunc("/upload", uploadHandler)

	// Create channel for graceful shutdown
	done := make(chan bool)

	// Start HTTP server
	server := &http.Server{Addr: ":8000"}
	go func() {
		fmt.Println("HTTP Server running on port 8000...")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP Server error: %v", err)
		}
		done <- true
	}()

	// Launch the browser window via the HTTP server
	window := ui.NewWindow()
	window.Bind("onClose", func(e ui.Event) interface{} {
		log.Println("Window closing...")
		server.Close()
		window.Close()
		return nil
	})

	if err := window.Show("http://localhost:8000/"); err != nil {
		log.Printf("Error showing window: %v", err)
		server.Close()
		return
	}

	ui.Wait()
	<-done
}
