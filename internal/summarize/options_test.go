package summarize

import "testing"

func TestResolveOptionsDefaults(t *testing.T) {
	t.Setenv("WR_SUMMARIZER_PROVIDER", "")
	t.Setenv("WR_SUMMARIZER_MODEL", "")
	t.Setenv("COPILOT_MODEL", "")

	got := ResolveOptions(Options{})
	if got.Provider != ProviderGroq {
		t.Fatalf("provider = %q, want %q", got.Provider, ProviderGroq)
	}
	if got.Model != groqDefaultModel {
		t.Fatalf("model = %q, want %q", got.Model, groqDefaultModel)
	}
}

func TestResolveOptionsCopilotUsesEnvModel(t *testing.T) {
	t.Setenv("WR_SUMMARIZER_PROVIDER", "copilot")
	t.Setenv("WR_SUMMARIZER_MODEL", "")
	t.Setenv("COPILOT_MODEL", "gpt-5-mini")

	got := ResolveOptions(Options{})
	if got.Provider != ProviderCopilot {
		t.Fatalf("provider = %q, want %q", got.Provider, ProviderCopilot)
	}
	if got.Model != "gpt-5-mini" {
		t.Fatalf("model = %q, want %q", got.Model, "gpt-5-mini")
	}
}

func TestValidateProvider(t *testing.T) {
	for _, provider := range []string{"", "groq", "copilot", "  COPILOT  "} {
		if err := ValidateProvider(provider); err != nil {
			t.Fatalf("ValidateProvider(%q) unexpected error: %v", provider, err)
		}
	}
	if err := ValidateProvider("bad"); err == nil {
		t.Fatal("ValidateProvider(bad) expected error")
	}
}

func TestResolveOptionsInvalidProvider(t *testing.T) {
	t.Setenv("WR_SUMMARIZER_PROVIDER", "bad")
	got := ResolveOptions(Options{})
	if got.Provider != "bad" {
		t.Fatalf("provider = %q, want bad", got.Provider)
	}
	if err := validateResolvedOptions(got); err == nil {
		t.Fatal("ValidateProvider(bad) expected error")
	}
}
