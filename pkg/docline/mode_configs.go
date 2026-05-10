package docline

import internalFramework "github.com/PavelMkr/docline-new/internal/framework"

// Type-safe configuration for the "automatic" finder.
type AutomaticConfig struct {
	MinCloneLength  int
	MinGroupPower   int
	ConvertToDRL    *bool
	ArchetypeLength *int
	StrictFilter    *bool
}

func (c AutomaticConfig) FinderType() string { return "automatic" }

func (c AutomaticConfig) toInternal(_ string) internalFramework.CloneFinderConfig {
	cp := map[string]interface{}{}
	if c.ConvertToDRL != nil {
		cp["convert_to_drl"] = *c.ConvertToDRL
	}
	if c.ArchetypeLength != nil {
		cp["archetype_length"] = *c.ArchetypeLength
	}
	if c.StrictFilter != nil {
		cp["strict_filter"] = *c.StrictFilter
	}
	if len(cp) == 0 {
		cp = nil
	}
	return internalFramework.CloneFinderConfig{
		MinCloneLength: c.MinCloneLength,
		MinGroupPower:  c.MinGroupPower,
		CustomParams:   cp,
	}
}

// Type-safe configuration for the "interactive" finder.
// type InteractiveConfig struct {
// 	MinCloneLength int
// 	MaxCloneLength int
// 	MinGroupPower  int
// 	UseArchetype   *bool
// }
//
// func (c InteractiveConfig) FinderType() string { return "interactive" }
//
// func (c InteractiveConfig) toInternal(_ string) internalFramework.CloneFinderConfig {
// 	cp := map[string]interface{}{
// 		"max_clone_length": c.MaxCloneLength,
// 	}
// 	if c.UseArchetype != nil {
// 		cp["use_archetype"] = *c.UseArchetype
// 	}
// 	if len(cp) == 0 {
// 		cp = nil
// 	}
// 	return internalFramework.CloneFinderConfig{
// 		MinCloneLength: c.MinCloneLength,
// 		MinGroupPower:  c.MinGroupPower,
// 		CustomParams:   cp,
// 	}
// }

// Type-safe configuration for the "ngram" finder.
type NgramConfig struct {
	MinCloneLength int
	MinGroupPower  int
	MaxEdit        int
	MaxFuzzy       int
	SourceLanguage string
}

func (c NgramConfig) FinderType() string { return "ngram" }

func (c NgramConfig) toInternal(filePath string) internalFramework.CloneFinderConfig {
	cp := map[string]interface{}{
		"max_edit":  c.MaxEdit,
		"max_fuzzy": c.MaxFuzzy,
		// "source_language": c.SourceLanguage,
		"file_path": filePath,
	}
	if len(cp) == 0 {
		cp = nil
	}
	return internalFramework.CloneFinderConfig{
		MinCloneLength: c.MinCloneLength,
		MinGroupPower:  c.MinGroupPower,
		CustomParams:   cp,
	}
}

// Type-safe configuration for the "heuristic" finder.
type HeuristicConfig struct {
	MinCloneLength      int
	MinGroupPower       int
	SimilarityThreshold float64
}

func (c HeuristicConfig) FinderType() string { return "heuristic" }

func (c HeuristicConfig) toInternal(_ string) internalFramework.CloneFinderConfig {
	cp := map[string]interface{}{
		"similarity_threshold": c.SimilarityThreshold,
	}
	if len(cp) == 0 {
		cp = nil
	}
	return internalFramework.CloneFinderConfig{
		MinCloneLength: c.MinCloneLength,
		MinGroupPower:  c.MinGroupPower,
		CustomParams:   cp,
	}
}

type OpenAIConfig struct {
	MinGroupPower int
	Provider      string
	Model         string
	Prompt        string
	APIKey        string
	BaseURL       string
	Endpoint      string
	TimeoutSec    int
}

func (c OpenAIConfig) FinderType() string { return "openai" }

func (c OpenAIConfig) toInternal(_ string) internalFramework.CloneFinderConfig {
	cp := map[string]interface{}{
		"provider":    c.Provider,
		"model":       c.Model,
		"prompt":      c.Prompt,
		"base_url":    c.BaseURL,
		"endpoint":    c.Endpoint,
		"timeout_sec": c.TimeoutSec,
	}
	if c.APIKey != "" {
		cp["openai_api_key"] = c.APIKey
	}
	return internalFramework.CloneFinderConfig{
		MinGroupPower: c.MinGroupPower,
		CustomParams:  cp,
	}
}
