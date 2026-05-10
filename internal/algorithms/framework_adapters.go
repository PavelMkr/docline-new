package internal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PavelMkr/docline-new/internal/framework"
)

// AutomaticModeAdapter adapts AutomaticModeSettings/ProcessAutomaticMode to the
// framework.CloneFinder interface.
type AutomaticModeAdapter struct{}

func (a *AutomaticModeAdapter) Name() string {
	return "automatic"
}

func (a *AutomaticModeAdapter) Description() string {
	return "Automatic mode clone finder using window-based exact matching"
}

func (a *AutomaticModeAdapter) FindClones(text string, cfg framework.CloneFinderConfig) ([]framework.CloneGroup, error) {
	settings := AutomaticModeSettings{
		MinCloneLength:  defaultInt(cfg.MinCloneLength, 20),
		ConvertToDRL:    getBool(cfg.CustomParams, "convert_to_drl", true),
		ArchetypeLength: getInt(cfg.CustomParams, "archetype_length", 5),
		StrictFilter:    getBool(cfg.CustomParams, "strict_filter", true),
	}

	groups, err := ProcessAutomaticMode(text, settings)
	if err != nil {
		return nil, err
	}

	// Apply additional MinGroupPower filter if requested via framework config.
	if cfg.MinGroupPower > 0 {
		filtered := make([]framework.CloneGroup, 0, len(groups))
		for _, g := range groups {
			if len(g.Fragments) >= cfg.MinGroupPower {
				filtered = append(filtered, g)
			}
		}
		groups = filtered
	}

	return groups, nil
}

// InteractiveModeAdapter adapts InteractiveModeSettings/ProcessInteractiveMode
// to the framework.CloneFinder interface.
// type InteractiveModeAdapter struct{}
//
// func (a *InteractiveModeAdapter) Name() string {
// 	return "interactive"
// }
//
// func (a *InteractiveModeAdapter) Description() string {
// 	return "Interactive mode clone finder with configurable length ranges"
// }
//
// func (a *InteractiveModeAdapter) FindClones(text string, cfg framework.CloneFinderConfig) ([]framework.CloneGroup, error) {
// 	settings := InteractiveModeSettings{
// 		MinCloneLength: defaultInt(cfg.MinCloneLength, 10),
// 		MaxCloneLength: getInt(cfg.CustomParams, "max_clone_length", 0),
// 		MinGroupPower:  defaultInt(cfg.MinGroupPower, 2),
// 		UseArchetype:   getBool(cfg.CustomParams, "use_archetype", false),
// 	}
//
// 	groups, err := ProcessInteractiveMode(text, settings)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return groups, nil
// }

// NGramAdapter adapts NgramDuplicateFinderData/FindDuplicatesByNGram to the
// framework.CloneFinder interface.
type NGramAdapter struct{}

func (a *NGramAdapter) Name() string {
	return "ngram"
}

func (a *NGramAdapter) Description() string {
	return "N-gram based duplicate finder using similarity metrics"
}

func (a *NGramAdapter) FindClones(text string, cfg framework.CloneFinderConfig) ([]framework.CloneGroup, error) {
	minClone := defaultInt(cfg.MinCloneLength, 2)
	if minClone < 1 {
		minClone = 1
	}

	data := NgramDuplicateFinderData{
		MinCloneSlider: minClone,
		MaxEditSlider:  getInt(cfg.CustomParams, "max_edit", 1),
		MaxFuzzySlider: getInt(cfg.CustomParams, "max_fuzzy", 1),
		// SourceLanguage: getString(cfg.CustomParams, "source_language", "english"),
		FilePath: getString(cfg.CustomParams, "file_path", ""),
	}

	parts := splitTextIntoParts(text)
	if len(parts) == 0 {
		return nil, nil
	}

	duplicates := FindDuplicatesByNGram(data, parts)
	groups := convertNGramResultsToGroups(duplicates)

	// Apply optional MinGroupPower from framework config.
	if cfg.MinGroupPower > 0 {
		filtered := make([]framework.CloneGroup, 0, len(groups))
		for _, g := range groups {
			if len(g.Fragments) >= cfg.MinGroupPower {
				filtered = append(filtered, g)
			}
		}
		groups = filtered
	}

	return groups, nil
}

// Heuristic
type HeuristicModeAdapter struct{}

func (a *HeuristicModeAdapter) Name() string {
	return "heuristic"
}

func (a *HeuristicModeAdapter) Description() string {
	return "Heuristic NLP duplicate finder based on token overlap"
}

func (a *HeuristicModeAdapter) FindClones(text string, cfg framework.CloneFinderConfig) ([]framework.CloneGroup, error) {
	segments := splitTextIntoSentenceSegments(text)
	if len(segments) == 0 {
		return nil, nil
	}

	minCloneLength := defaultInt(cfg.MinCloneLength, 4)
	threshold := getFloat(cfg.CustomParams, "similarity_threshold", 0.72)
	if threshold <= 0 {
		threshold = 0.72
	}
	if threshold > 1 {
		threshold = 1
	}

	parent := make([]int, len(segments))
	for i := range parent {
		parent[i] = i
	}

	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			sim := jaccardSimilarity(segments[i].tokenFreq, segments[j].tokenFreq)
			if sim >= threshold {
				union(parent, i, j)
			}
		}
	}

	components := map[int][]int{}
	for i := range segments {
		root := find(parent, i)
		components[root] = append(components[root], i)
	}

	groups := make([]framework.CloneGroup, 0, len(components))
	for _, idxs := range components {
		if len(idxs) < maxInt(defaultInt(cfg.MinGroupPower, 2), 2) {
			continue
		}

		sort.Slice(idxs, func(i, j int) bool {
			return segments[idxs[i]].startToken < segments[idxs[j]].startToken
		})

		fragments := make([]framework.TextFragment, 0, len(idxs))
		for _, idx := range idxs {
			seg := segments[idx]
			if seg.tokenCount < minCloneLength {
				continue
			}
			fragments = append(fragments, framework.TextFragment{
				Content:  seg.text,
				StartPos: seg.startToken,
				EndPos:   seg.endToken,
			})
		}

		if len(fragments) < maxInt(defaultInt(cfg.MinGroupPower, 2), 2) {
			continue
		}

		groups = append(groups, framework.CloneGroup{
			Fragments: fragments,
			Power:     len(fragments),
			Archetype: fragments[0].Content,
			Metadata: map[string]interface{}{
				"finder":               "heuristic-nlp",
				"similarity_threshold": threshold,
			},
		})
	}

	return groups, nil
}

// RegisterCloneFinders registers all built-in clone finders in the given registry.
func RegisterCloneFinders(reg *framework.PluginRegistry) error {
	if err := reg.RegisterCloneFinder(&AutomaticModeAdapter{}); err != nil {
		return fmt.Errorf("register automatic finder: %w", err)
	}
	// if err := reg.RegisterCloneFinder(&InteractiveModeAdapter{}); err != nil {
	// 	return fmt.Errorf("register interactive finder: %w", err)
	// }
	if err := reg.RegisterCloneFinder(&HeuristicModeAdapter{}); err != nil {
		return fmt.Errorf("register heuristic finder: %w", err)
	}
	if err := reg.RegisterCloneFinder(&NGramAdapter{}); err != nil {
		return fmt.Errorf("register ngram finder: %w", err)
	}
	if err := reg.RegisterCloneFinder(&OpenAI{}); err != nil {
		return fmt.Errorf("register openai fuzzy finder: %w", err)
	}
	return nil
}

// splitTextIntoParts is a local equivalent of the old SplitTextIntoParts helper,
// kept here to avoid depending on the CLI package.
func splitTextIntoParts(text string) []string {
	re := regexp.MustCompile(`[.!?]`)
	rawParts := re.Split(text, -1)

	var parts []string
	for _, p := range rawParts {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// convertNGramResultsToGroups converts the map-based n-gram results into
// framework.CloneGroup structures.
func convertNGramResultsToGroups(ngramResults map[string][]string) []framework.CloneGroup {
	var groups []framework.CloneGroup
	for _, fragments := range ngramResults {
		if len(fragments) == 0 {
			continue
		}
		group := framework.CloneGroup{
			Fragments: make([]framework.TextFragment, len(fragments)),
			Power:     len(fragments),
		}
		for i, frag := range fragments {
			toks := strings.Fields(frag)
			start := 0
			end := start + len(toks)
			group.Fragments[i] = framework.TextFragment{
				Content:  frag,
				StartPos: start,
				EndPos:   end,
			}
		}
		group.Archetype = group.Fragments[0].Content
		groups = append(groups, group)
	}
	return groups
}

// Helper accessors for CloneFinderConfig.CustomParams.

func getBool(m map[string]interface{}, key string, def bool) bool {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getInt(m map[string]interface{}, key string, def int) int {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		switch vv := v.(type) {
		case int:
			return vv
		case int32:
			return int(vv)
		case int64:
			return int(vv)
		case float32:
			return int(vv)
		case float64:
			return int(vv)
		}
	}
	return def
}

func getString(m map[string]interface{}, key, def string) string {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func defaultInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
