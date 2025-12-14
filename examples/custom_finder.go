package main

import (
	"fmt"
	"strings"

	"Docline/framework"
)

// CustomCloneFinder demonstrates how to create a custom clone finder
type CustomCloneFinder struct {
	name string
}

func (c *CustomCloneFinder) Name() string {
	return c.name
}

func (c *CustomCloneFinder) Description() string {
	return "Custom clone finder that finds exact duplicate sentences"
}

func (c *CustomCloneFinder) FindClones(text string, config framework.CloneFinderConfig) ([]framework.CloneGroup, error) {
	// Split text into sentences
	sentences := strings.Split(text, ".")
	
	// Find duplicate sentences
	sentenceMap := make(map[string][]framework.TextFragment)
	for i, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) < config.MinCloneLength {
			continue
		}
		
		// Simple token count
		tokens := strings.Fields(sentence)
		if len(tokens) < config.MinCloneLength {
			continue
		}
		
		// Track position (approximate)
		startPos := 0
		if i > 0 {
			// Approximate position based on previous sentences
			for j := 0; j < i; j++ {
				startPos += len(strings.Fields(sentences[j]))
			}
		}
		
		frag := framework.TextFragment{
			Content:  sentence,
			StartPos: startPos,
			EndPos:   startPos + len(tokens),
		}
		
		sentenceMap[sentence] = append(sentenceMap[sentence], frag)
	}
	
	// Build groups
	var groups []framework.CloneGroup
	for sentence, fragments := range sentenceMap {
		if len(fragments) >= config.MinGroupPower {
			groups = append(groups, framework.CloneGroup{
				Fragments: fragments,
				Power:     len(fragments),
				Archetype: sentence,
			})
		}
	}
	
	return groups, nil
}

func main() {
	// Create framework
	fw := framework.NewFramework(nil)
	
	// Register custom finder
	customFinder := &CustomCloneFinder{name: "custom-sentence"}
	err := fw.GetRegistry().RegisterCloneFinder(customFinder)
	if err != nil {
		panic(err)
	}
	
	// Use custom finder
	result, err := fw.AnalyzeDocument(
		"example.txt",
		"custom-sentence",
		framework.CloneFinderConfig{
			MinCloneLength: 5,
			MinGroupPower:  2,
		},
	)
	
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Custom finder found %d groups\n", len(result.Groups))
}

