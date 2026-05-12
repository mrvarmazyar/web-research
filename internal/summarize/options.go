package summarize

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const (
	ProviderGroq    = "groq"
	ProviderCopilot = "copilot"

	defaultProvider  = ProviderGroq
	groqDefaultModel = "llama-3.1-8b-instant"
	maxInputRunes    = 12000 // ~3k tokens
	maxOutputTokens  = 1024
	systemPrompt     = "You are a concise technical research assistant. Extract and summarize relevant information from web content. Be direct and specific."
	defaultFetchTask = "summarize the key information from this page"
)

type Options struct {
	Provider string
	Model    string
}

func Summarize(ctx context.Context, content, prompt string, opts Options) (string, error) {
	content = truncateContent(content)
	resolved := ResolveOptions(opts)
	if err := validateResolvedOptions(resolved); err != nil {
		return content, err
	}

	switch resolved.Provider {
	case ProviderCopilot:
		return summarizeWithCopilot(ctx, content, prompt, resolved.Model)
	case ProviderGroq:
		return summarizeWithGroq(ctx, content, prompt, resolved.Model)
	default:
		return content, fmt.Errorf("unsupported provider %q", resolved.Provider)
	}
}

func ResolveOptions(opts Options) Options {
	provider := normalizeProvider(firstNonEmpty(opts.Provider, os.Getenv("WR_SUMMARIZER_PROVIDER")))
	if provider == "" {
		provider = defaultProvider
	}

	model := strings.TrimSpace(firstNonEmpty(opts.Model, os.Getenv("WR_SUMMARIZER_MODEL")))
	if model == "" && provider == ProviderCopilot {
		model = strings.TrimSpace(os.Getenv("COPILOT_MODEL"))
	}
	if model == "" && provider == ProviderGroq {
		model = groqDefaultModel
	}

	return Options{Provider: provider, Model: model}
}

func ValidateProvider(provider string) error {
	switch normalizeProvider(provider) {
	case "", ProviderGroq, ProviderCopilot:
		return nil
	default:
		return fmt.Errorf("provider must be one of: %s, %s", ProviderGroq, ProviderCopilot)
	}
}

func validateResolvedOptions(opts Options) error {
	if err := ValidateProvider(opts.Provider); err != nil {
		return err
	}
	if normalizeProvider(opts.Provider) == "" {
		return fmt.Errorf("provider must not be empty")
	}
	return nil
}

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

func truncateContent(content string) string {
	runes := []rune(content)
	if len(runes) > maxInputRunes {
		return string(runes[:maxInputRunes]) + "\n\n[content truncated]"
	}
	return content
}

func buildUserPrompt(content, prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		prompt = defaultFetchTask
	}
	return fmt.Sprintf("Content:\n%s\n\nTask: %s\n\nRespond concisely, only what's relevant.", content, prompt)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
