package main

import (
	"fmt"
	"strings"

	"github.com/mrvarmazyar/web-research/internal/research"
)

func formatSearch(r *research.SearchResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Search results for: %s\n\n", r.Query)
	for _, res := range r.Results {
		fmt.Fprintf(&b, "[%s]\n%s\n%s\n\n", res.Title, res.URL, res.Snippet)
	}
	return strings.TrimSpace(b.String())
}

func formatFetch(r *research.FetchResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "URL: %s\n", r.URL)
	if r.CacheHit {
		fmt.Fprintf(&b, "Cache: hit\n")
	}
	if r.TokenStats != nil {
		fmt.Fprintf(&b, "Tokens: %d raw → %d summary (%.0f%% reduction)\n",
			r.TokenStats.RawEstimatedTokens,
			r.TokenStats.SummaryEstimatedTokens,
			r.TokenStats.ReductionPercent,
		)
	}
	fmt.Fprintf(&b, "\n%s", r.Summary)
	return b.String()
}

func formatResearch(r *research.ResearchResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Research: %s\n\n", r.Query)
	fmt.Fprintf(&b, "%s\n\n", r.Answer)
	fmt.Fprintf(&b, "## Sources (%d fetched, %d cache hits)\n\n",
		r.Stats.FetchedPages, r.Stats.CacheHits)
	for i, src := range r.Sources {
		fmt.Fprintf(&b, "%d. [%s](%s)\n", i+1, src.Title, src.URL)
	}
	return strings.TrimSpace(b.String())
}
