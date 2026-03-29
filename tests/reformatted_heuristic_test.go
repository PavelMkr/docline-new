package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	alg "github.com/PavelMkr/docline-new/internal/algorithms"
	"github.com/PavelMkr/docline-new/internal/framework"
	rep "github.com/PavelMkr/docline-new/internal/report"
)

func TestHeuristic_CreatesReformattedFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "sample.drl")
	input := `<?xml version="1.0" encoding="UTF-8"?>
	<d:DocumentationCore xmlns:d="https://docbook.org/ns/docbook/">
	<d:InfElement>alpha	beta
	alpha	beta</d:InfElement>
	</d:DocumentationCore>`

	if err := os.WriteFile(srcPath, []byte(input), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	fw := framework.NewFramework(&framework.Config{
		ResultsDirectory:    tmpDir,
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "heuristic",
	})

	if err := framework.RegisterBuiltInPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterBuiltInPlugins: %v", err)
	}
	if err := rep.RegisterDocumentPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterDocumentPlugins: %v", err)
	}
	if err := alg.RegisterCloneFinders(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterCloneFinders: %v", err)
	}

	result, err := fw.AnalyzeDocumentWithConfig(srcPath, "heuristic", framework.CloneFinderConfig{
		MinCloneLength: 2,
		MinGroupPower:  1,
		CustomParams: map[string]interface{}{
			"extension_point_checkbox": true, // heuristic analysis enabled
		},
	})
	if err != nil {
		t.Fatalf("AnalyzeDocument failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil analysis result")
	}

	reformattedPath := srcPath + ".reformatted"
	data, err := os.ReadFile(reformattedPath)
	if err != nil {
		t.Fatalf("expected .reformatted file: %v", err)
	}
	text := string(data)

	if strings.Contains(text, "\r") {
		t.Fatalf("expected CR removed, got: %q", text)
	}
	if strings.Contains(text, "\t") {
		t.Fatalf("expected TAB replaced, got: %q", text)
	}
	if strings.Count(text, "alpha    beta") < 2 {
		t.Fatalf("expected at least 2 occurrences of normalized phrase, got: %q", text)
	}
	if !strings.Contains(text, "\n") {
		t.Fatalf("expected newlines preserved, got: %q", text)
	}

	// Optional metadata checks (current implementation writes this)
	if got, ok := result.Metadata["reformatted_file"].(string); !ok || got != reformattedPath {
		t.Fatalf("expected metadata reformatted_file=%q, got=%v", reformattedPath, result.Metadata["reformatted_file"])
	}
}