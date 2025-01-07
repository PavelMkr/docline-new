package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type PageData struct {
	Title string
}

// Automatic Mode Data
type AutomaticModeData struct {
	LengthSlider           int  `json:"length_slider"`
	ConvertCheckbox        bool `json:"convert_checkbox"`
	ArchetypeSlider        int  `json:"archetype_slider"`
	StrictFilteringCheckbox bool `json:"strict_filtering_checkbox"`
}

// Interactive Mode Data
type InteractiveModeData struct {
	MinCloneSlider   int  `json:"min_clone_slider"`
	MaxCloneSlider   int  `json:"max_clone_slider"`
	MinGroupSlider   int  `json:"min_group_slider"`
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
	ExtensionPointCheckbox bool `json:"extention_point_checkbox"`
}

func main() {
	// Parse HTML templates
	templates := template.Must(template.ParseFiles("index.html"))
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
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

	// Start server
	fmt.Println("Server is running on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

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
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		return
	}

	fmt.Println("Heuristic Finder Data Received:", data)
	response := map[string]string{"status": "success", "message": "Heuristic finder processed"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
