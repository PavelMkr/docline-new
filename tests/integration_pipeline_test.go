package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	alg "github.com/PavelMkr/docline-new/internal/algorithms"
	"github.com/PavelMkr/docline-new/internal/framework"
	rep "github.com/PavelMkr/docline-new/internal/report"
	"github.com/PavelMkr/docline-new/pkg/docline"
)

func registerFullFramework(t *testing.T, tmpDir string) *framework.Framework {
	t.Helper()
	cfg := &framework.Config{
		ResultsDirectory:    tmpDir,
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "automatic",
	}
	fw := framework.NewFramework(cfg)
	reg := fw.GetRegistry()
	if err := framework.RegisterBuiltInPlugins(reg); err != nil {
		t.Fatalf("RegisterBuiltInPlugins: %v", err)
	}
	if err := alg.RegisterCloneFinders(reg); err != nil {
		t.Fatalf("RegisterCloneFinders: %v", err)
	}
	if err := rep.RegisterDocumentPlugins(reg); err != nil {
		t.Fatalf("RegisterDocumentPlugins: %v", err)
	}
	if err := rep.RegisterReportGenerators(reg); err != nil {
		t.Fatalf("RegisterReportGenerators: %v", err)
	}
	return fw
}

func writeDuplicateDocBook(t *testing.T, dir string) string {
	t.Helper()
	p := filepath.Join(dir, "dup.xml")
	const doc = `<?xml version="1.0" encoding="UTF-8"?>
<book>
  <para>alpha beta gamma delta epsilon</para>
  <para>alpha beta gamma delta epsilon</para>
  <para>unique content only once here</para>
</book>`
	if err := os.WriteFile(p, []byte(doc), 0o644); err != nil {
		t.Fatalf("write doc: %v", err)
	}
	return p
}

func TestFramework_Integration_FindsClonesWithStatistics(t *testing.T) {
	tmpDir := t.TempDir()
	fw := registerFullFramework(t, tmpDir)
	docPath := writeDuplicateDocBook(t, tmpDir)

	result, err := fw.AnalyzeDocument(docPath, "automatic", framework.CloneFinderConfig{
		MinCloneLength: 3,
		MinGroupPower:  2,
		CustomParams: map[string]interface{}{
			"convert_to_drl": false,
			"strict_filter": false,
		},
	})
	if err != nil {
		t.Fatalf("AnalyzeDocument: %v", err)
	}
	if len(result.Groups) == 0 {
		t.Fatal("expected at least one clone group")
	}
	if result.Statistics.TotalGroups != len(result.Groups) {
		t.Fatalf("stats.TotalGroups %d != len(Groups) %d", result.Statistics.TotalGroups, len(result.Groups))
	}
	if result.Statistics.TotalFragments < 2 {
		t.Fatalf("expected TotalFragments >= 2, got %d", result.Statistics.TotalFragments)
	}
	if src, ok := result.Metadata["source_file"].(string); !ok || src != docPath {
		t.Fatalf("metadata source_file = %v", result.Metadata["source_file"])
	}
	if finder, ok := result.Metadata["finder"].(string); !ok || finder != "automatic" {
		t.Fatalf("metadata finder = %v", result.Metadata["finder"])
	}
}

func TestFramework_Integration_LineNumbersInFragmentMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	fw := registerFullFramework(t, tmpDir)

	// Use DocBook XML so analysis does not depend on pandoc (plain .txt triggers conversion).
	const doc = `<?xml version="1.0" encoding="UTF-8"?>
<book>
  <para>line one tokens here</para>
  <para>alpha beta gamma delta epsilon zeta</para>
  <para>alpha beta gamma delta epsilon zeta</para>
</book>`
	docPath := filepath.Join(tmpDir, "lines.xml")
	if err := os.WriteFile(docPath, []byte(doc), 0o644); err != nil {
		t.Fatalf("write doc: %v", err)
	}

	result, err := fw.AnalyzeDocument(docPath, "automatic", framework.CloneFinderConfig{
		MinCloneLength: 5,
		MinGroupPower:  2,
		CustomParams: map[string]interface{}{
			"convert_to_drl": false,
			"strict_filter": false,
		},
	})
	if err != nil {
		t.Fatalf("AnalyzeDocument: %v", err)
	}
	if len(result.Groups) == 0 {
		t.Fatal("expected clone groups for repeated line content")
	}

	hasLineMeta := false
	for _, g := range result.Groups {
		for _, fr := range g.Fragments {
			if fr.Metadata == nil {
				continue
			}
			if _, ok := fr.Metadata["source_line_start"]; ok {
				hasLineMeta = true
			}
		}
	}
	if !hasLineMeta {
		t.Fatal("expected source_line_start in fragment metadata")
	}
}

func TestFramework_Integration_AnalyzeAndGenerateReport(t *testing.T) {
	tmpDir := t.TempDir()
	fw := registerFullFramework(t, tmpDir)
	docPath := writeDuplicateDocBook(t, tmpDir)

	result, err := fw.AnalyzeDocument(docPath, "automatic", framework.CloneFinderConfig{
		MinCloneLength: 3,
		MinGroupPower:  2,
		CustomParams: map[string]interface{}{
			"convert_to_drl": false,
			"strict_filter": false,
		},
	})
	if err != nil {
		t.Fatalf("AnalyzeDocument: %v", err)
	}

	for _, format := range []string{"html", "json", "csv"} {
		out := filepath.Join(tmpDir, "report."+format)
		if err := fw.GenerateReport(result, format, out); err != nil {
			t.Fatalf("GenerateReport %s: %v", format, err)
		}
		b, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("read %s: %v", format, err)
		}
		if len(b) == 0 {
			t.Fatalf("%s report is empty", format)
		}
		if format == "json" {
			var payload map[string]interface{}
			if err := json.Unmarshal(b, &payload); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if _, ok := payload["stats"]; !ok {
				t.Fatal("json report should include stats from settings")
			}
		}
	}
}

func TestDocline_Integration_FullPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	docPath := writeDuplicateDocBook(t, tmpDir)

	d := docline.New(&docline.Config{
		ResultsDirectory:    tmpDir,
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "automatic",
	})

	result, err := d.AnalyzeDocument(docPath, "automatic", docline.AutomaticConfig{
		MinCloneLength: 3,
		MinGroupPower:  2,
		ConvertToDRL:   boolPtr(false),
		StrictFilter:   boolPtr(false),
	})
	if err != nil {
		t.Fatalf("AnalyzeDocument: %v", err)
	}
	if len(result.Groups) == 0 {
		t.Fatal("expected clone groups via public API")
	}

	out := filepath.Join(tmpDir, "public_report.html")
	if err := d.GenerateReport(result, "html", out); err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	html, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), "Clone Analysis Report") {
		t.Fatal("expected report title in HTML output")
	}
}

func TestDocline_Integration_HeuristicAndNgramModes(t *testing.T) {
	tmpDir := t.TempDir()
	docPath := writeDuplicateDocBook(t, tmpDir)

	d := docline.New(&docline.Config{
		ResultsDirectory:    tmpDir,
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "automatic",
	})

	cases := []struct {
		name   string
		finder string
		cfg    docline.FinderModeConfig
	}{
		{
			name:   "heuristic",
			finder: "heuristic",
			cfg: docline.HeuristicConfig{
				MinCloneLength:      3,
				MinGroupPower:       2,
				SimilarityThreshold: 0.5,
			},
		},
		{
			name:   "ngram",
			finder: "ngram",
			cfg: docline.NgramConfig{
				MinCloneLength: 2,
				MinGroupPower:  2,
				MaxEdit:        1,
				MaxFuzzy:       50,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := d.AnalyzeDocument(docPath, tc.finder, tc.cfg)
			if err != nil {
				t.Fatalf("AnalyzeDocument: %v", err)
			}
			if res == nil {
				t.Fatal("expected non-nil result")
			}
			if res.Metadata["finder"] != tc.finder {
				t.Fatalf("finder metadata = %v", res.Metadata["finder"])
			}
		})
	}
}

func TestFramework_AnalyzeDocument_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	fw := registerFullFramework(t, tmpDir)

	_, err := fw.AnalyzeDocument(filepath.Join(tmpDir, "missing.xml"), "automatic", framework.CloneFinderConfig{
		MinCloneLength: 2,
		MinGroupPower:  2,
	})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "failed to read document") {
		t.Fatalf("error = %v", err)
	}
}

func TestFramework_AnalyzeDocument_UnknownFinder(t *testing.T) {
	tmpDir := t.TempDir()
	fw := registerFullFramework(t, tmpDir)
	docPath := writeDuplicateDocBook(t, tmpDir)

	_, err := fw.AnalyzeDocument(docPath, "nonexistent", framework.CloneFinderConfig{
		MinCloneLength: 2,
		MinGroupPower:  2,
	})
	if err == nil {
		t.Fatal("expected error for unknown finder")
	}
	if !strings.Contains(err.Error(), "clone finder") {
		t.Fatalf("error = %v", err)
	}
}

func boolPtr(b bool) *bool { return &b }
