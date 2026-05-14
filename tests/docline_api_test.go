package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PavelMkr/docline-new/pkg/docline"
)

func TestDocline_AnalyzeDocument_NilFinderConfig(t *testing.T) {
	d := docline.New(&docline.Config{
		ResultsDirectory:    t.TempDir(),
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "automatic",
	})
	_, err := d.AnalyzeDocument(filepath.Join(t.TempDir(), "x.xml"), "automatic", nil)
	if err == nil {
		t.Fatal("expected error for nil FinderModeConfig")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Fatalf("error = %v", err)
	}
}

func TestDocline_AnalyzeDocument_FinderTypeMismatch(t *testing.T) {
	d := docline.New(&docline.Config{
		ResultsDirectory:    t.TempDir(),
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "automatic",
	})
	cfg := docline.AutomaticConfig{MinCloneLength: 2, MinGroupPower: 2}
	_, err := d.AnalyzeDocument(filepath.Join(t.TempDir(), "x.xml"), "heuristic", cfg)
	if err == nil {
		t.Fatal("expected error when finderType does not match config")
	}
	if !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("error = %v", err)
	}
}

func TestDocline_AnalyzeDocument_Automatic_EmptyFinderTypeUsesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	docPath := filepath.Join(tmpDir, "doc.xml")
	const xml = `<?xml version="1.0" encoding="UTF-8"?>
<book>
	<para>duplicate fragment here</para>
	<para>duplicate fragment here</para>
</book>`
	if err := os.WriteFile(docPath, []byte(xml), 0o644); err != nil {
		t.Fatalf("write doc: %v", err)
	}

	d := docline.New(&docline.Config{
		ResultsDirectory:    tmpDir,
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  "automatic",
	})
	cfg := docline.AutomaticConfig{MinCloneLength: 2, MinGroupPower: 2}
	res, err := d.AnalyzeDocument(docPath, "", cfg)
	if err != nil {
		t.Fatalf("AnalyzeDocument: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
}
