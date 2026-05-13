package internal

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	alg "github.com/PavelMkr/docline-new/internal/algorithms"
	"github.com/PavelMkr/docline-new/internal/framework"
	rep "github.com/PavelMkr/docline-new/internal/report"
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

// TestFramework_AnalyzeDocument_YAML_AllBuiltInFinders ensures .yaml files are
// read through the same document pipeline and can be analyzed by all local
// built-in finders.
func TestFramework_AnalyzeDocument_YAML_AllBuiltInFinders(t *testing.T) {
	tmpDir := t.TempDir()
	docPath := filepath.Join(tmpDir, "doc.yaml")

	const doc = `rules:
  - duplicate fragment here
  - duplicate fragment here
metadata:
  owner: team`
	if err := os.WriteFile(docPath, []byte(doc), 0o644); err != nil {
		t.Fatalf("write temp yaml doc: %v", err)
	}

	cfg := &framework.Config{
		ResultsDirectory:    tmpDir,
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
	}
	fw := framework.NewFramework(cfg)

	if err := framework.RegisterBuiltInPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterBuiltInPlugins: %v", err)
	}
	if err := rep.RegisterDocumentPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterDocumentPlugins: %v", err)
	}
	if err := alg.RegisterCloneFinders(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterCloneFinders: %v", err)
	}

	finders := []string{"automatic", "heuristic", "ngram"}
	for _, finder := range finders {
		result, err := fw.AnalyzeDocument(docPath, finder, framework.CloneFinderConfig{
			MinCloneLength: 2,
			MinGroupPower:  2,
		})
		if err != nil {
			t.Fatalf("AnalyzeDocument for %s failed: %v", finder, err)
		}
		if result == nil {
			t.Fatalf("expected non-nil result for finder %s", finder)
		}
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

// TestFramework_GenerateReport_AllFormats checks html, csv, and json outputs are non-empty and json is valid.
func TestFramework_GenerateReport_AllFormats(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &framework.Config{ResultsDirectory: tmpDir, DefaultReportFormat: "html"}
	fw := framework.NewFramework(cfg)
	if err := framework.RegisterBuiltInPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterBuiltInPlugins: %v", err)
	}
	if err := alg.RegisterCloneFinders(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterCloneFinders: %v", err)
	}
	if err := rep.RegisterDocumentPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterDocumentPlugins: %v", err)
	}
	if err := rep.RegisterReportGenerators(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterReportGenerators: %v", err)
	}

	longDup := strings.TrimSpace(strings.Repeat("tokword ", 20)) // >10 tokens; must appear in CSV like HTML/JSON
	result := &framework.AnalysisResult{
		Groups: []framework.CloneGroup{
			{
				Fragments: []framework.TextFragment{
					{Content: "a b c", StartPos: 0, EndPos: 3},
					{Content: "a b c", StartPos: 10, EndPos: 13},
				},
				Power:     2,
				Archetype: "a b c",
			},
			{
				Fragments: []framework.TextFragment{
					{Content: longDup, StartPos: 20, EndPos: 40},
					{Content: longDup, StartPos: 100, EndPos: 120},
				},
				Power:     2,
				Archetype: longDup,
			},
		},
		Metadata: map[string]interface{}{"source_file": "smoke.txt"},
	}

	wantRows := 0
	for _, g := range result.Groups {
		if len(g.Fragments) > 0 {
			wantRows++
		}
	}

	for _, format := range []string{"html", "csv", "json"} {
		out := filepath.Join(tmpDir, "out."+format)
		if err := fw.GenerateReport(result, format, out); err != nil {
			t.Fatalf("format %s: %v", format, err)
		}
		st, err := os.Stat(out)
		if err != nil {
			t.Fatalf("format %s stat: %v", format, err)
		}
		if st.Size() == 0 {
			t.Fatalf("format %s: empty file", format)
		}
		if format == "json" {
			b, err := os.ReadFile(out)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]interface{}
			if err := json.Unmarshal(b, &payload); err != nil {
				t.Fatalf("format json invalid: %v", err)
			}
			groups, ok := payload["groups"].([]interface{})
			if !ok {
				t.Fatalf("json groups type %T", payload["groups"])
			}
			if len(groups) != wantRows {
				t.Fatalf("json groups len %d want %d", len(groups), wantRows)
			}
		}
	}

	csvPath := filepath.Join(tmpDir, "out.csv")
	raw, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	cr := csv.NewReader(bytes.NewReader(raw))
	cr.Comma = ';'
	recs, err := cr.ReadAll()
	if err != nil {
		t.Fatalf("csv read: %v", err)
	}
	if len(recs) != wantRows+1 {
		t.Fatalf("csv rows %d want header+%d data rows", len(recs), wantRows)
	}
}

// Regression: GenerateReport must not panic when Metadata is nil or source_file is absent (see examples/custom_report).
func TestFramework_GenerateReport_NoSourceFileInMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	fw := framework.NewFramework(&framework.Config{ResultsDirectory: tmpDir})
	if err := framework.RegisterBuiltInPlugins(fw.GetRegistry()); err != nil {
		t.Fatal(err)
	}
	if err := rep.RegisterReportGenerators(fw.GetRegistry()); err != nil {
		t.Fatal(err)
	}
	result := &framework.AnalysisResult{
		Groups: []framework.CloneGroup{
			{Fragments: []framework.TextFragment{{Content: "x y", StartPos: 0, EndPos: 2}, {Content: "x y", StartPos: 5, EndPos: 7}}, Power: 2, Archetype: "x y"},
		},
		Metadata: nil,
	}
	out := filepath.Join(tmpDir, "r.html")
	if err := fw.GenerateReport(result, "html", out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal(err)
	}
}
