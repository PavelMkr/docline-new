package internal

import (
	"strings"
	"testing"

	alg "github.com/PavelMkr/docline-new/internal/algorithms"
	"github.com/PavelMkr/docline-new/internal/framework"
)

func TestGenerateNGrams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		n    int
		want []string
	}{
		{
			name: "trigrams",
			text: "one two three four",
			n:    3,
			want: []string{"one two three", "two three four"},
		},
		{
			name: "n larger than word count",
			text: "a b",
			n:    5,
			want: nil,
		},
		{
			name: "n zero",
			text: "a b c",
			n:    0,
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := alg.GenerateNGrams(tt.text, tt.n)
			if len(got) != len(tt.want) {
				t.Fatalf("len %d want %d: %v", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("[%d] got %q want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestBuildNGramMapAndSimilarity(t *testing.T) {
	t.Parallel()

	m1 := alg.BuildNGramMap("alpha beta gamma", 2)
	m2 := alg.BuildNGramMap("alpha beta delta", 2)

	// shared bigram: "alpha beta"
	if m1["alpha beta"] != 1 || m2["alpha beta"] != 1 {
		t.Fatalf("expected shared bigram in both maps: m1=%v m2=%v", m1, m2)
	}

	sim := alg.CalculateNGramSimilarity(m1, m2)
	// |intersection|=1, |union|=3 => 1/3
	if sim < 0.33 || sim > 0.34 {
		t.Fatalf("expected ~0.33 similarity, got %v", sim)
	}

	if alg.CalculateNGramSimilarity(nil, nil) != 0 {
		t.Fatal("empty maps should yield 0 similarity")
	}
}

func TestFindDuplicatesByNGram(t *testing.T) {
	t.Parallel()

	texts := []string{
		"foo bar baz qux",
		"foo bar baz qux",
		"completely different text here",
	}
	data := alg.NgramDuplicateFinderData{
		MinCloneSlider: 2,
		MaxFuzzySlider: 50, // threshold 0.5
	}
	dups := alg.FindDuplicatesByNGram(data, texts)
	if len(dups) == 0 {
		t.Fatal("expected at least one duplicate pair")
	}
	foundPair := false
	for src, targets := range dups {
		if strings.Contains(src, "foo bar") {
			for _, tgt := range targets {
				if strings.Contains(tgt, "foo bar") {
					foundPair = true
				}
			}
		}
	}
	if !foundPair {
		t.Fatalf("expected similar segments to be linked, got %v", dups)
	}
}

func TestProcessAutomaticMode_FindsExactDuplicates(t *testing.T) {
	t.Parallel()

	const text = "alpha beta gamma delta alpha beta gamma delta"
	groups, err := alg.ProcessAutomaticMode(text, alg.AutomaticModeSettings{
		MinCloneLength:  3,
		ConvertToDRL:    false,
		ArchetypeLength: 3,
		StrictFilter:    false,
	})
	if err != nil {
		t.Fatalf("ProcessAutomaticMode: %v", err)
	}
	if len(groups) == 0 {
		t.Fatal("expected at least one clone group for repeated 3-token window")
	}
	totalFrags := 0
	for _, g := range groups {
		totalFrags += len(g.Fragments)
	}
	if totalFrags < 2 {
		t.Fatalf("expected multiple fragments, got %d in %d groups", totalFrags, len(groups))
	}
}

func TestAutomaticModeAdapter_MinGroupPower(t *testing.T) {
	t.Parallel()

	adapter := &alg.AutomaticModeAdapter{}
	const text = "one two three one two three"
	groups, err := adapter.FindClones(text, framework.CloneFinderConfig{
		MinCloneLength: 3,
		MinGroupPower:  10, // filter out everything
		CustomParams: map[string]interface{}{
			"convert_to_drl": false,
			"strict_filter": false,
		},
	})
	if err != nil {
		t.Fatalf("FindClones: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected no groups with MinGroupPower=10, got %d", len(groups))
	}
}

func TestHeuristicModeAdapter_FindsSimilarSegments(t *testing.T) {
	t.Parallel()

	adapter := &alg.HeuristicModeAdapter{}
	text := "The quick brown fox jumps over the lazy dog.\n" +
		"The quick brown fox leaps over the lazy dog.\n" +
		"Something entirely unrelated to animals."

	groups, err := adapter.FindClones(text, framework.CloneFinderConfig{
		MinCloneLength: 4,
		MinGroupPower:  2,
		CustomParams: map[string]interface{}{
			"similarity_threshold": 0.5,
		},
	})
	if err != nil {
		t.Fatalf("FindClones: %v", err)
	}
	if len(groups) == 0 {
		t.Fatal("expected heuristic finder to group similar sentences")
	}
}

func TestRegisterCloneFinders_ListsExpectedNames(t *testing.T) {
	reg := framework.NewPluginRegistry()
	if err := alg.RegisterCloneFinders(reg); err != nil {
		t.Fatalf("RegisterCloneFinders: %v", err)
	}
	names := reg.ListCloneFinders()
	want := map[string]bool{"automatic": true, "heuristic": true, "ngram": true, "openai": true}
	for _, n := range names {
		delete(want, n)
	}
	if len(want) > 0 {
		t.Fatalf("missing finders %v; registered: %v", want, names)
	}
}
