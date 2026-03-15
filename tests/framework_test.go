package internal

import (
	"os"
	"path/filepath"
	"testing"

	alg "Docline/internal/algorithms"
	"Docline/internal/framework"
	rep "Docline/internal/report"
)

// dummyFinder is a minimal CloneFinder implementation used in registry tests.
type dummyFinder struct{}

func (d *dummyFinder) FindClones(text string, cfg framework.CloneFinderConfig) ([]framework.CloneGroup, error) {
	return nil, nil
}

func (d *dummyFinder) Name() string        { return "dummy" }
func (d *dummyFinder) Description() string { return "dummy finder" }

// TestPluginRegistry_Basic verifies that the registry can register and retrieve plugins.
func TestPluginRegistry_Basic(t *testing.T) {
	reg := framework.NewPluginRegistry()

	if err := reg.RegisterCloneFinder(&dummyFinder{}); err != nil {
		t.Fatalf("RegisterCloneFinder failed: %v", err)
	}

	if _, err := reg.GetCloneFinder("dummy"); err != nil {
		t.Fatalf("GetCloneFinder failed: %v", err)
	}

	if _, err := reg.GetCloneFinder("missing"); err == nil {
		t.Fatalf("expected error for missing clone finder")
	}
}

// TestFramework_AnalyzeDocument_Automatic ensures that AnalyzeDocument works end-to-end
// with the built-in automatic finder and DocBook parser.
func TestFramework_AnalyzeDocument_Automatic(t *testing.T) {
	tmpDir := t.TempDir()
	docPath := filepath.Join(tmpDir, "doc.xml")

	const doc = `<?xml version="1.0" encoding="UTF-8"?>
<book>
	<para>duplicate fragment here</para>
	<para>duplicate fragment here</para>
</book>`
	if err := os.WriteFile(docPath, []byte(doc), 0o644); err != nil {
		t.Fatalf("write temp doc: %v", err)
	}

	cfg := &framework.Config{
		ResultsDirectory:    tmpDir,
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "automatic",
	}
	fw := framework.NewFramework(cfg)

	// Register core utilities and built-in plugins.
	if err := framework.RegisterBuiltInPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterBuiltInPlugins: %v", err)
	}
	if err := rep.RegisterDocumentPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterDocumentPlugins: %v", err)
	}
	if err := rep.RegisterReportGenerators(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterReportGenerators: %v", err)
	}
	if err := alg.RegisterCloneFinders(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterCloneFinders: %v", err)
	}

	result, err := fw.AnalyzeDocument(docPath, "automatic", framework.CloneFinderConfig{
		MinCloneLength: 2,
		MinGroupPower:  2,
	})
	if err != nil {
		t.Fatalf("AnalyzeDocument failed: %v", err)
	}

	if result == nil {
		t.Fatalf("expected non-nil analysis result")
	}
}

// TestFramework_GenerateReport verifies that GenerateReport produces an HTML file.
func TestFramework_GenerateReport(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &framework.Config{
		ResultsDirectory:    tmpDir,
		DefaultReportFormat: "html",
	}
	fw := framework.NewFramework(cfg)

	if err := framework.RegisterBuiltInPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterBuiltInPlugins: %v", err)
	}
	if err := rep.RegisterReportGenerators(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterReportGenerators: %v", err)
	}

	groups := []framework.CloneGroup{
		{
			Fragments: []framework.TextFragment{
				{Content: "foo bar", StartPos: 0, EndPos: 2},
				{Content: "foo bar", StartPos: 10, EndPos: 12},
			},
			Power:     2,
			Archetype: "foo bar",
		},
	}

	result := &framework.AnalysisResult{
		Groups: groups,
		Metadata: map[string]interface{}{
			"source_file": "synthetic",
		},
	}

	outPath := filepath.Join(tmpDir, "report.html")
	if err := fw.GenerateReport(result, "html", outPath); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected report file to exist, got: %v", err)
	}
}
