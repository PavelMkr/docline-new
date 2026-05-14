package internal

import (
	"strings"
	"testing"

	rep "github.com/PavelMkr/docline-new/internal/report"
)

func TestYAMLParserAdapter_NameAndFormats(t *testing.T) {
	var a rep.YAMLParserAdapter
	if a.Name() != "yaml" {
		t.Fatalf("Name() = %q", a.Name())
	}
	got := a.SupportedFormats()
	want := map[string]bool{".yaml": true, ".yml": true}
	if len(got) != len(want) {
		t.Fatalf("SupportedFormats len = %d, want %d", len(got), len(want))
	}
	for _, ext := range got {
		if !want[ext] {
			t.Fatalf("unexpected extension %q", ext)
		}
	}
}

func TestYAMLParserAdapter_Parse(t *testing.T) {
	var a rep.YAMLParserAdapter
	const yaml = "  key: value\n  other: 1\n"
	segs, err := a.Parse(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if segs[0] != strings.TrimSpace(yaml) {
		t.Fatalf("segment = %q, want trimmed raw", segs[0])
	}
}
