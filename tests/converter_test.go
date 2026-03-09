package internal

import (
	"testing"

	rep "Docline/internal/report"
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
