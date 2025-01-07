package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Page 1
type AutomaticModeData struct {
	LengthSlider           int  `json:"length_slider"`
	ConvertCheckbox        bool `json:"convert_checkbox"`
	ArchetypeSlider        int  `json:"archetype_slider"`
	StrictFilteringCheckbox bool `json:"strict_filtering_checkbox"`
}

func AutomaticModeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data AutomaticModeData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	fmt.Println("Automatic Mode Data Received:")
	fmt.Printf("Length Slider: %d\n", data.LengthSlider)
	fmt.Printf("Convert Checkbox: %t\n", data.ConvertCheckbox)
	fmt.Printf("Archetype Slider: %d\n", data.ArchetypeSlider)
	fmt.Printf("Strict Filtering Checkbox: %t\n", data.StrictFilteringCheckbox)

	// Пример ответа в формате JSON
	response := map[string]string{"status": "success", "message": "Data received successfully"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Page 2
type InteractiveModeData struct {
	MinCloneSlider   int  `json:"min_clone_slider"`
	MaxCloneSlider   int  `json:"max_clone_slider"`
	MinGroupSlider   int  `json:"min_group_slider"`
	ExtensionCheckbox bool `json:"extension_checkbox"`
}

func InteractiveModeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data InteractiveModeData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	fmt.Println("Interactive Mode Data Received:")
	fmt.Printf("Min Clone Slider: %d\n", data.MinCloneSlider)
	fmt.Printf("Max Clone Slider: %d\n", data.MaxCloneSlider)
	fmt.Printf("Min Group Slider: %d\n", data.MinGroupSlider)
	fmt.Printf("Extension Checkbox: %t\n", data.ExtensionCheckbox)

	response := map[string]string{"status": "success", "message": "Data received successfully"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Page 3
type NgramDuplicateFinderData struct {
	MinCloneSlider int    `json:"min_clone_slider"`
	MaxEditSlider  int    `json:"max_edit_slider"`
	MaxFuzzySlider int    `json:"max_fuzzy_slider"`
	SourceLanguage string `json:"source_language"`
}

func NgramDuplicateFinderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data NgramDuplicateFinderData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	fmt.Println("Ngram Duplicate Finder Data Received:")
	fmt.Printf("Min Clone Slider: %d\n", data.MinCloneSlider)
	fmt.Printf("Max Edit Slider: %d\n", data.MaxEditSlider)
	fmt.Printf("Max Fuzzy Slider: %d\n", data.MaxFuzzySlider)
	fmt.Printf("Source Language: %s\n", data.SourceLanguage)

	response := map[string]string{"status": "success", "message": "Data received successfully"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Page 4
type HeuristicNgramFinderData struct {
	ExtensionPointCheckbox bool `json:"extention_point_checkbox"`
}

func HeuristicNgramFinderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data HeuristicNgramFinderData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	fmt.Println("Heuristic Ngram Finder Data Received:")
	fmt.Printf("Extension Point Checkbox: %t\n", data.ExtensionPointCheckbox)

	response := map[string]string{"status": "success", "message": "Data received successfully"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/automatic_mode", AutomaticModeHandler)
	http.HandleFunc("/interactive_mode", InteractiveModeHandler)
	http.HandleFunc("/ngram_finder", NgramDuplicateFinderHandler)
	http.HandleFunc("/heuristic_finder", HeuristicNgramFinderHandler)

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
