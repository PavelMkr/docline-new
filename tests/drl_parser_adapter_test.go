package internal

import (
	"strings"
	"testing"

	rep "github.com/PavelMkr/docline-new/internal/report"
)

func TestDRLParserAdapter_ExtractsAllTextNodes(t *testing.T) {
	xmlDoc := `<?xml version="1.0" encoding="UTF-8"?>
<d:DocumentationCore xmlns:d="http://math.spbu.ru/drl" xmlns="http://docbook.org/ns/docbook">
  <d:InfElement id="root" name="test">
    First text block
    <section>Second text block</section>
  </d:InfElement>
  <d:OtherElement>Third text block</d:OtherElement>
</d:DocumentationCore>`

	parser := &rep.DRLParserAdapter{}
	segments, err := parser.Parse(strings.NewReader(xmlDoc))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(segments) == 0 {
		t.Fatal("expected non-empty segments")
	}

	joined := strings.Join(segments, "\n")

	// Ensure text content is present
	for _, want := range []string{"First text block", "Second text block", "Third text block"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected parsed text to contain %q, got: %q", want, joined)
		}
	}

	// Ensure tags are not emitted as text
	if strings.Contains(joined, "<d:InfElement") || strings.Contains(joined, "</d:DocumentationCore>") {
		t.Fatalf("expected plain text only, got XML tags in output: %q", joined)
	}
}