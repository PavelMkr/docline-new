package internal

import (
	"os"
	"path/filepath"
	"testing"

	alg "github.com/PavelMkr/docline-new/internal/algorithms"
	"github.com/PavelMkr/docline-new/internal/framework"
	rep "github.com/PavelMkr/docline-new/internal/report"
)

// fakePandocConverter is a test double that pretends to convert documents to DocBook XML.
// It is intentionally registered under the name "pandoc" to match framework.core.go lookup.
type fakePandocConverter struct {
	t      *testing.T
	called bool
}

func (f *fakePandocConverter) Name() string { return "pandoc" }

func (f *fakePandocConverter) Convert(inputPath string, outputFormat string) (string, error) {
	f.called = true
	out := filepath.Join(filepath.Dir(inputPath), filepath.Base(inputPath)+".converted.xml")
	doc := `<?xml version="1.0" encoding="UTF-8"?>
<book>
  <para>duplicate fragment here</para>
  <para>duplicate fragment here</para>
</book>`
	if err := os.WriteFile(out, []byte(doc), 0o644); err != nil {
		f.t.Fatalf("write fake converted docbook: %v", err)
	}
	return out, nil
}

func (f *fakePandocConverter) IsConversionNeeded(filePath string) bool {
	switch filepath.Ext(filePath) {
	case ".txt", ".html", ".htm", ".md", ".doc", ".docx", ".odt", ".rtf":
		return true
	default:
		return false
	}
}

func (f *fakePandocConverter) SupportedInputFormats() []string {
	return []string{".doc", ".docx", ".odt", ".rtf", ".md", ".txt", ".html", ".htm"}
}

func (f *fakePandocConverter) SupportedOutputFormats() []string {
	return []string{".xml", ".dbk", ".docbook"}
}

func newFrameworkForFormatTests(t *testing.T) (*framework.Framework, *fakePandocConverter) {
	t.Helper()
	cfg := &framework.Config{
		ResultsDirectory:    t.TempDir(),
		DefaultReportFormat: "html",
		DefaultTokenizer:    "space",
	}
	fw := framework.NewFramework(cfg)

	if err := framework.RegisterBuiltInPlugins(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterBuiltInPlugins: %v", err)
	}
	if err := alg.RegisterCloneFinders(fw.GetRegistry()); err != nil {
		t.Fatalf("RegisterCloneFinders: %v", err)
	}

	// Register parsers explicitly; do NOT register the real pandoc adapter here.
	if err := fw.GetRegistry().RegisterDocumentParser(&rep.DocBookParserAdapter{}); err != nil {
		t.Fatalf("register docbook parser: %v", err)
	}
	if err := fw.GetRegistry().RegisterDocumentParser(&rep.YAMLParserAdapter{}); err != nil {
		t.Fatalf("register yaml parser: %v", err)
	}

	fake := &fakePandocConverter{t: t}
	if err := fw.GetRegistry().RegisterDocumentConverter(fake); err != nil {
		t.Fatalf("register fake pandoc converter: %v", err)
	}

	return fw, fake
}

func TestFramework_AnalyzeDocument_TXT_UsesConverterPipeline(t *testing.T) {
	fw, fake := newFrameworkForFormatTests(t)

	p := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	res, err := fw.AnalyzeDocument(p, "automatic", framework.CloneFinderConfig{MinCloneLength: 2, MinGroupPower: 2})
	if err != nil {
		t.Fatalf("AnalyzeDocument: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if !fake.called {
		t.Fatalf("expected converter to be used for .txt")
	}
}

func TestFramework_AnalyzeDocument_HTML_UsesConverterPipeline(t *testing.T) {
	fw, fake := newFrameworkForFormatTests(t)

	p := filepath.Join(t.TempDir(), "doc.html")
	if err := os.WriteFile(p, []byte("<p>hello</p>"), 0o644); err != nil {
		t.Fatalf("write html: %v", err)
	}

	res, err := fw.AnalyzeDocument(p, "automatic", framework.CloneFinderConfig{MinCloneLength: 2, MinGroupPower: 2})
	if err != nil {
		t.Fatalf("AnalyzeDocument: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if !fake.called {
		t.Fatalf("expected converter to be used for .html")
	}
}

func TestFramework_AnalyzeDocument_YML_UsesYAMLParser(t *testing.T) {
	fw, fake := newFrameworkForFormatTests(t)

	p := filepath.Join(t.TempDir(), "doc.yml")
	if err := os.WriteFile(p, []byte("a: duplicate fragment here\nb: duplicate fragment here\n"), 0o644); err != nil {
		t.Fatalf("write yml: %v", err)
	}

	res, err := fw.AnalyzeDocument(p, "automatic", framework.CloneFinderConfig{MinCloneLength: 2, MinGroupPower: 2})
	if err != nil {
		t.Fatalf("AnalyzeDocument: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if fake.called {
		t.Fatalf("did not expect converter to be used for .yml (yaml parser should handle it)")
	}
}

