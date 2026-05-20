package main

import "testing"

func TestParseSummaryOptions(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		provider  string
		model     string
		mode      string
		topK      int
		remaining []string
		wantErr   bool
	}{
		{name: "no options", args: []string{"https://example.com", "prompt"}, remaining: []string{"https://example.com", "prompt"}},
		{name: "split flags", args: []string{"--provider", "copilot", "--model", "gpt-5-mini", "q"}, provider: "copilot", model: "gpt-5-mini", remaining: []string{"q"}},
		{name: "equals flags", args: []string{"--provider=groq", "--model=llama-3.1-8b-instant", "q"}, provider: "groq", model: "llama-3.1-8b-instant", remaining: []string{"q"}},
		{name: "stop parsing after first positional", args: []string{"query", "--provider", "copilot"}, remaining: []string{"query", "--provider", "copilot"}},
		{name: "double dash keeps remaining", args: []string{"--provider", "groq", "--", "--provider", "copilot"}, provider: "groq", remaining: []string{"--provider", "copilot"}},
		{name: "invalid provider", args: []string{"--provider", "bad", "q"}, wantErr: true},
		{name: "unknown option", args: []string{"--bogus", "q"}, wantErr: true},
		{name: "missing provider value", args: []string{"--provider"}, wantErr: true},
		{name: "missing model value", args: []string{"--model"}, wantErr: true},
		{name: "mode lossless", args: []string{"--mode", "lossless", "q"}, mode: "lossless", remaining: []string{"q"}},
		{name: "mode chunks", args: []string{"--mode", "chunks", "q"}, mode: "chunks", remaining: []string{"q"}},
		{name: "mode equals", args: []string{"--mode=chunks", "q"}, mode: "chunks", remaining: []string{"q"}},
		{name: "top-k flag", args: []string{"--top-k", "3", "q"}, topK: 3, remaining: []string{"q"}},
		{name: "top-k equals", args: []string{"--top-k=7", "q"}, topK: 7, remaining: []string{"q"}},
		{name: "invalid mode", args: []string{"--mode", "bad", "q"}, wantErr: true},
		{name: "missing mode value", args: []string{"--mode"}, wantErr: true},
		{name: "missing top-k value", args: []string{"--top-k"}, wantErr: true},
		{name: "top-k not int", args: []string{"--top-k", "abc", "q"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, remaining, err := parseSummaryOptions(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseSummaryOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Provider != tt.provider {
				t.Fatalf("provider = %q, want %q", got.Provider, tt.provider)
			}
			if got.Model != tt.model {
				t.Fatalf("model = %q, want %q", got.Model, tt.model)
			}
			if got.Mode != tt.mode {
				t.Fatalf("mode = %q, want %q", got.Mode, tt.mode)
			}
			if got.TopK != tt.topK {
				t.Fatalf("topK = %d, want %d", got.TopK, tt.topK)
			}
			if len(remaining) != len(tt.remaining) {
				t.Fatalf("remaining len = %d, want %d", len(remaining), len(tt.remaining))
			}
			for i := range remaining {
				if remaining[i] != tt.remaining[i] {
					t.Fatalf("remaining[%d] = %q, want %q", i, remaining[i], tt.remaining[i])
				}
			}
		})
	}
}
