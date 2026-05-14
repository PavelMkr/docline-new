package internal

import (
	"os"
	"path/filepath"
	"testing"

	rep "github.com/PavelMkr/docline-new/internal/report"
)

func TestIsConversionNeeded(t *testing.T) {
	conv := rep.NewDocumentConverter()
	if !conv.IsConversionNeeded("file.docx") {
		t.Error("expected conversion needed for .docx")
	}
	if conv.IsConversionNeeded("file.xml") {
		t.Error("expected no conversion needed for .xml")
	}
}

func TestDocumentConverter_IsConversionNeeded_Table(t *testing.T) {
	conv := rep.NewDocumentConverter()
	tests := []struct {
		path string
		want bool
	}{
		{"a.docx", true},
		{"a.DOCX", true},
		{"b.xml", false},
		{"b.dbk", false},
		{"b.docbook", false},
		{"unknown.pdf", false},
		{"readme.txt", true},
		{"page.htm", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := conv.IsConversionNeeded(tt.path); got != tt.want {
				t.Fatalf("IsConversionNeeded(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestDocumentConverter_CleanupTempFile(t *testing.T) {
	conv := rep.NewDocumentConverter()
	p := filepath.Join(t.TempDir(), "docline_cleanup_test.xml")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	if err := conv.CleanupTempFile(p); err != nil {
		t.Fatalf("CleanupTempFile: %v", err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, stat err = %v", err)
	}
}
