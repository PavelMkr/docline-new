package internal

import (
	"strings"
	"testing"

	rep "github.com/PavelMkr/docline-new/internal/report"
)

func TestDocBookParserAdapter_Parse(t *testing.T) {
	var a rep.DocBookParserAdapter
	const xml = `<?xml version="1.0" encoding="UTF-8"?>
<book><para>alpha</para><para>beta</para></book>`
	segs, err := a.Parse(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(segs) != 2 || segs[0] != "alpha" || segs[1] != "beta" {
		t.Fatalf("segments = %#v", segs)
	}
	if a.Name() != "docbook" {
		t.Fatalf("Name = %q", a.Name())
	}
}

func TestPandocConverterAdapter_UnsupportedOutput(t *testing.T) {
	a := rep.NewPandocConverterAdapter()
	_, err := a.Convert("/nonexistent/path.docx", ".pdf")
	if err == nil {
		t.Fatal("expected error for unsupported output format")
	}
	if !strings.Contains(err.Error(), "unsupported output") {
		t.Fatalf("error = %v", err)
	}
}

func TestPandocConverterAdapter_IsConversionNeeded(t *testing.T) {
	a := rep.NewPandocConverterAdapter()
	if !a.IsConversionNeeded("x.docx") {
		t.Fatal("expected conversion for .docx")
	}
	if a.IsConversionNeeded("x.xml") {
		t.Fatal("expected no conversion for existing DocBook")
	}
}

func TestPandocConverterAdapter_SupportedFormatsNonEmpty(t *testing.T) {
	a := rep.NewPandocConverterAdapter()
	in := a.SupportedInputFormats()
	out := a.SupportedOutputFormats()
	if len(in) == 0 || len(out) == 0 {
		t.Fatalf("empty formats in=%d out=%d", len(in), len(out))
	}
	seenIn := map[string]bool{}
	for _, ext := range in {
		seenIn[ext] = true
	}
	if !seenIn[".docx"] {
		t.Fatal("expected .docx in SupportedInputFormats")
	}
	foundXML := false
	for _, ext := range out {
		if ext == ".xml" {
			foundXML = true
			break
		}
	}
	if !foundXML {
		t.Fatal("expected .xml in SupportedOutputFormats")
	}
}
