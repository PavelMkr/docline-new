package internal

import (
	"testing"

	alg "github.com/PavelMkr/docline-new/internal/algorithms"
	"github.com/PavelMkr/docline-new/internal/framework"
	rep "github.com/PavelMkr/docline-new/internal/report"
)

func TestSpaceTokenizer_Tokenize(t *testing.T) {
	t.Parallel()

	tok := &framework.SpaceTokenizer{}
	got := tok.Tokenize("  hello   world\nfoo\tbar  ")
	want := []string{"hello", "world", "foo", "bar"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] got %q want %q", i, got[i], want[i])
		}
	}
	if tok.Name() != "space" {
		t.Fatalf("Name() = %q", tok.Name())
	}
}

func TestStrictFilter_MinArchetypeLength(t *testing.T) {
	t.Parallel()

	f := &framework.StrictFilter{}
	groups := []framework.CloneGroup{
		{
			Archetype: "a b",
			Fragments: []framework.TextFragment{
				{Content: "a b", StartPos: 0, EndPos: 2},
				{Content: "a b", StartPos: 5, EndPos: 7},
			},
			Power: 2,
		},
		{
			Archetype: "one two three four five",
			Fragments: []framework.TextFragment{
				{Content: "one two three four five", StartPos: 0, EndPos: 5},
				{Content: "one two three four five", StartPos: 10, EndPos: 15},
			},
			Power: 2,
		},
	}

	out := f.Filter(groups, framework.FilterConfig{MinArchetypeLength: 5})
	if len(out) != 1 {
		t.Fatalf("expected 1 group after min archetype filter, got %d", len(out))
	}
	if out[0].Archetype != "one two three four five" {
		t.Fatalf("unexpected archetype %q", out[0].Archetype)
	}
}

func TestNewFramework_NilConfigDefaults(t *testing.T) {
	fw := framework.NewFramework(nil)
	if fw == nil {
		t.Fatal("expected non-nil framework")
	}
	if fw.GetRegistry() == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestPluginRegistry_DuplicateCloneFinderRegistration(t *testing.T) {
	reg := framework.NewPluginRegistry()
	dummy := &dummyFinder{}
	if err := reg.RegisterCloneFinder(dummy); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := reg.RegisterCloneFinder(dummy); err == nil {
		t.Fatal("expected duplicate clone finder registration to fail")
	}
}

func TestPluginRegistry_GetReportGeneratorUnknownFormat(t *testing.T) {
	reg := framework.NewPluginRegistry()
	if err := framework.RegisterBuiltInPlugins(reg); err != nil {
		t.Fatal(err)
	}
	if err := rep.RegisterReportGenerators(reg); err != nil {
		t.Fatal(err)
	}
	if _, err := reg.GetReportGenerator("pdf"); err == nil {
		t.Fatal("expected error for unknown report format")
	}
}

func TestPluginRegistry_GetDocumentParserUnknownExtension(t *testing.T) {
	reg := framework.NewPluginRegistry()
	if err := rep.RegisterDocumentPlugins(reg); err != nil {
		t.Fatal(err)
	}
	if _, err := reg.GetDocumentParser(".unknown"); err == nil {
		t.Fatal("expected error for unknown extension")
	}
}

func TestRegisterCloneFinders_DuplicateFails(t *testing.T) {
	reg := framework.NewPluginRegistry()
	if err := alg.RegisterCloneFinders(reg); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := alg.RegisterCloneFinders(reg); err == nil {
		t.Fatal("expected duplicate registration to fail")
	}
}
