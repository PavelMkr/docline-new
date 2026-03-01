package main

import (
	"fmt"
	"log"

	alg "Docline/internal/algorithms"
	rep "Docline/internal/report"
	"Docline/internal/framework"
)

func main() {
	// Create framework instance
	config := &framework.Config{
		ResultsDirectory:    "./results",
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "automatic",
	}

	fw := framework.NewFramework(config)

	// Register built-in components
	registerBuiltins(fw)

	// Analyze a document
	result, err := fw.AnalyzeDocument(
		"example.xml",
		"automatic",
		framework.CloneFinderConfig{
			MinCloneLength:      20,
			MinGroupPower:       2,
			SimilarityThreshold: 0.9,
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d clone groups\n", result.Statistics.TotalGroups)
	fmt.Printf("Total fragments: %d\n", result.Statistics.TotalFragments)

	// Generate report
	err = fw.GenerateReport(result, "html", "./results/report.html")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Analysis complete!")
}

func registerBuiltins(fw *framework.Framework) {
	registry := fw.GetRegistry()

	// Core framework utilities (tokenizer, similarity, filters)
	_ = framework.RegisterBuiltInPlugins(registry)

	// Built-in algorithms, document parser/converter and report generators.
	_ = alg.RegisterCloneFinders(registry)
	_ = rep.RegisterDocumentPlugins(registry)
	_ = rep.RegisterReportGenerators(registry)
}
