package research

import "testing"

func TestStripListPrefix(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"1. stripe webhook idempotency", "stripe webhook idempotency"},
		{"2. golang concurrency", "golang concurrency"},
		{"10. some query", "some query"},
		{"1) stripe webhook", "stripe webhook"},
		{"2) another query", "another query"},
		{"- stripe webhooks", "stripe webhooks"},
		{"* golang channels", "golang channels"},
		{"• bullet query", "bullet query"},
		{"no prefix query", "no prefix query"},
		{"1.not a prefix", "1.not a prefix"}, // no space after dot
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := stripListPrefix(tt.in); got != tt.want {
				t.Errorf("stripListPrefix(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseSubQueries(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		fallback string
		want     []string
	}{
		{
			name:     "empty raw returns fallback",
			raw:      "",
			fallback: "original query",
			want:     []string{"original query"},
		},
		{
			name:     "whitespace-only returns fallback",
			raw:      "   \n  \n",
			fallback: "original query",
			want:     []string{"original query"},
		},
		{
			name:     "normal three lines",
			raw:      "golang concurrency patterns\ngolang goroutine best practices\ngolang channel usage",
			fallback: "golang",
			want:     []string{"golang concurrency patterns", "golang goroutine best practices", "golang channel usage"},
		},
		{
			name:     "blank lines stripped",
			raw:      "query one\n\nquery two\n",
			fallback: "fallback",
			want:     []string{"query one", "query two"},
		},
		{
			name:     "more than 3 lines capped at 3",
			raw:      "q1\nq2\nq3\nq4\nq5",
			fallback: "fallback",
			want:     []string{"q1", "q2", "q3"},
		},
		{
			name:     "single line",
			raw:      "single query",
			fallback: "fallback",
			want:     []string{"single query"},
		},
		{
			name:     "numbered LLM output stripped",
			raw:      "1. stripe webhook idempotency\n2. stripe retry handling\n3. stripe idempotency key",
			fallback: "stripe",
			want:     []string{"stripe webhook idempotency", "stripe retry handling", "stripe idempotency key"},
		},
		{
			name:     "bulleted LLM output stripped",
			raw:      "- golang concurrency\n* goroutine patterns\n• channel usage",
			fallback: "golang",
			want:     []string{"golang concurrency", "goroutine patterns", "channel usage"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSubQueries(tt.raw, tt.fallback)
			if len(got) != len(tt.want) {
				t.Fatalf("parseSubQueries() = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
