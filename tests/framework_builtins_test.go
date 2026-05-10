package internal

import (
	"testing"

	"github.com/PavelMkr/docline-new/internal/framework"
)

func TestRegisterBuiltInPlugins_RegistersAll(t *testing.T) {
	reg := framework.NewPluginRegistry()
	if err := framework.RegisterBuiltInPlugins(reg); err != nil {
		t.Fatalf("RegisterBuiltInPlugins: %v", err)
	}

	if _, err := reg.GetTextTokenizer("space"); err != nil {
		t.Fatalf("expected space tokenizer: %v", err)
	}
	if _, err := reg.GetSimilarityCalculator("jaccard"); err != nil {
		t.Fatalf("expected jaccard similarity: %v", err)
	}
	if _, err := reg.GetFilter("strict"); err != nil {
		t.Fatalf("expected strict filter: %v", err)
	}
}

func TestJaccardSimilarityCalculator(t *testing.T) {
	j := &framework.JaccardSimilarityCalculator{}

	if got := j.CalculateSimilarity("", ""); got != 0 {
		t.Fatalf("expected 0 for empty union, got %v", got)
	}

	got := j.CalculateSimilarity("a b c", "b c d")
	// intersection {b,c} = 2; union {a,b,c,d} = 4 => 0.5
	if got != 0.5 {
		t.Fatalf("expected 0.5, got %v", got)
	}
}

func TestStrictFilter_RemoveOverlaps(t *testing.T) {
	f := &framework.StrictFilter{}
	groups := []framework.CloneGroup{
		{
			Archetype: "x",
			Fragments: []framework.TextFragment{
				{Content: "a", StartPos: 0, EndPos: 2},
				{Content: "b", StartPos: 1, EndPos: 3}, // overlaps with first
				{Content: "c", StartPos: 4, EndPos: 5}, // non-overlapping
			},
			Power: 3,
		},
	}

	out := f.Filter(groups, framework.FilterConfig{RemoveOverlaps: true})
	if len(out) != 1 {
		t.Fatalf("expected 1 group, got %d", len(out))
	}
	if len(out[0].Fragments) != 2 {
		t.Fatalf("expected 2 fragments after overlap removal, got %d", len(out[0].Fragments))
	}
	if out[0].Power != 2 {
		t.Fatalf("expected power=2 after overlap removal, got %d", out[0].Power)
	}
}

