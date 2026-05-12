package research

type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet,omitempty"`
}

type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
}

type FetchRequest struct {
	URL      string `json:"url"`
	Prompt   string `json:"prompt,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

type FetchResponse struct {
	URL                string      `json:"url"`
	Summary            string      `json:"summary"`
	CacheHit           bool        `json:"cache_hit"`
	SourceLengthChars  int         `json:"source_length_chars,omitempty"`
	SummaryLengthChars int         `json:"summary_length_chars,omitempty"`
	TokenStats         *TokenStats `json:"token_stats,omitempty"`
}

type ResearchRequest struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results,omitempty"`
	Focus      string `json:"focus,omitempty"`
	Provider   string `json:"provider,omitempty"`
	Model      string `json:"model,omitempty"`
}

type SourceSummary struct {
	Title    string `json:"title,omitempty"`
	URL      string `json:"url"`
	Summary  string `json:"summary"`
	CacheHit bool   `json:"cache_hit,omitempty"`
}

type ResearchStats struct {
	SearchedResults int `json:"searched_results"`
	FetchedPages    int `json:"fetched_pages"`
	CacheHits       int `json:"cache_hits"`
}

type ResearchResponse struct {
	Query   string          `json:"query"`
	Answer  string          `json:"answer"`
	Sources []SourceSummary `json:"sources"`
	Stats   ResearchStats   `json:"stats"`
}

type TokenStats struct {
	RawEstimatedTokens     int     `json:"raw_estimated_tokens"`
	SummaryEstimatedTokens int     `json:"summary_estimated_tokens"`
	ReductionPercent       float64 `json:"reduction_percent"`
}

func estimateTokens(s string) int {
	return len(s) / 4
}

func reductionPct(raw, summary int) float64 {
	if raw == 0 {
		return 0
	}
	return (1 - float64(summary)/float64(raw)) * 100
}
