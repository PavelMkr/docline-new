package internal

import (
	"testing"

	"github.com/PavelMkr/docline-new/internal/framework"
)

type dummyPlugin struct {
	initCalled bool
}

func (d *dummyPlugin) Name() string    { return "dummy-plugin" }
func (d *dummyPlugin) Version() string { return "0.0.1" }
func (d *dummyPlugin) Initialize(_ map[string]interface{}) error {
	d.initCalled = true
	return nil
}
func (d *dummyPlugin) Shutdown() error { return nil }

type dummyTokenizer struct{}

func (d *dummyTokenizer) Name() string                { return "dummy-tokenizer" }
func (d *dummyTokenizer) Tokenize(text string) []string { return []string{text} }

func TestPluginRegistry_DuplicateRegistrationErrors(t *testing.T) {
	reg := framework.NewPluginRegistry()

	if err := reg.RegisterTextTokenizer(&dummyTokenizer{}); err != nil {
		t.Fatalf("RegisterTextTokenizer: %v", err)
	}
	if err := reg.RegisterTextTokenizer(&dummyTokenizer{}); err == nil {
		t.Fatalf("expected duplicate tokenizer registration to fail")
	}

	p := &dummyPlugin{}
	if err := reg.RegisterPlugin(p, map[string]interface{}{"k": "v"}); err != nil {
		t.Fatalf("RegisterPlugin: %v", err)
	}
	if !p.initCalled {
		t.Fatalf("expected plugin Initialize to be called")
	}
	if err := reg.RegisterPlugin(&dummyPlugin{}, nil); err == nil {
		t.Fatalf("expected duplicate plugin registration to fail")
	}
	if err := reg.UnregisterPlugin("missing"); err == nil {
		t.Fatalf("expected unregister missing to fail")
	}
	if err := reg.UnregisterPlugin(p.Name()); err != nil {
		t.Fatalf("UnregisterPlugin: %v", err)
	}
}

