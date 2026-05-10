package research

import "testing"

func TestValidateSearch(t *testing.T) {
	tests := []struct {
		name    string
		req     SearchRequest
		wantErr bool
	}{
		{"empty query", SearchRequest{Query: ""}, true},
		{"whitespace query", SearchRequest{Query: "   "}, true},
		{"negative limit", SearchRequest{Query: "q", Limit: -1}, true},
		{"limit too high", SearchRequest{Query: "q", Limit: 11}, true},
		{"valid", SearchRequest{Query: "q", Limit: 5}, false},
		{"zero limit ok", SearchRequest{Query: "q", Limit: 0}, false},
		{"max limit ok", SearchRequest{Query: "q", Limit: 10}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSearch(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSearch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFetch(t *testing.T) {
	tests := []struct {
		name    string
		req     FetchRequest
		wantErr bool
	}{
		{"empty url", FetchRequest{URL: ""}, true},
		{"no scheme", FetchRequest{URL: "example.com"}, true},
		{"ftp scheme", FetchRequest{URL: "ftp://example.com"}, true},
		{"http ok", FetchRequest{URL: "http://example.com"}, false},
		{"https ok", FetchRequest{URL: "https://example.com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFetch(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFetch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateResearch(t *testing.T) {
	tests := []struct {
		name    string
		req     ResearchRequest
		wantErr bool
	}{
		{"empty query", ResearchRequest{Query: ""}, true},
		{"max_results too high", ResearchRequest{Query: "q", MaxResults: 6}, true},
		{"valid", ResearchRequest{Query: "q", MaxResults: 3}, false},
		{"zero max ok", ResearchRequest{Query: "q", MaxResults: 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResearch(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateResearch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTokenEstimation(t *testing.T) {
	text := "hello world" // 11 chars → 2 tokens
	got := estimateTokens(text)
	if got != 2 {
		t.Errorf("estimateTokens(%q) = %d, want 2", text, got)
	}
}

func TestReductionPct(t *testing.T) {
	if pct := reductionPct(1000, 500); pct != 50 {
		t.Errorf("reductionPct(1000, 500) = %v, want 50", pct)
	}
	if pct := reductionPct(0, 100); pct != 0 {
		t.Errorf("reductionPct(0, 100) = %v, want 0", pct)
	}
}
