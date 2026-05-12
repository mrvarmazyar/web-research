package research

import (
	"context"
	"fmt"
	"strings"

	"github.com/mrvarmazyar/web-research/internal/cache"
	"github.com/mrvarmazyar/web-research/internal/fetch"
	"github.com/mrvarmazyar/web-research/internal/search"
	"github.com/mrvarmazyar/web-research/internal/summarize"
)

const (
	defaultSearchLimit = 5
	maxSearchLimit     = 10
	defaultMaxResults  = 3
	maxMaxResults      = 5
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Search(_ context.Context, req SearchRequest) (*SearchResponse, error) {
	if err := validateSearch(req); err != nil {
		return nil, err
	}
	if req.Limit == 0 {
		req.Limit = defaultSearchLimit
	}

	results, err := search.Search(req.Query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if req.Limit < len(results) {
		results = results[:req.Limit]
	}

	out := make([]SearchResult, len(results))
	for i, r := range results {
		out[i] = SearchResult{Title: r.Title, URL: r.URL, Snippet: r.Snippet}
	}
	return &SearchResponse{Query: req.Query, Results: out}, nil
}

func (s *Service) Fetch(ctx context.Context, req FetchRequest) (*FetchResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validateFetch(req); err != nil {
		return nil, err
	}

	var content string
	cacheHit := false

	if cached, ok := cache.Get(req.URL); ok {
		content = cached
		cacheHit = true
	} else {
		var err error
		content, err = fetch.Fetch(req.URL)
		if err != nil {
			return nil, fmt.Errorf("fetch failed: %w", err)
		}
		_ = cache.Set(req.URL, content)
	}

	summary, err := summarize.Summarize(ctx, content, req.Prompt, summarize.Options{
		Provider: req.Provider,
		Model:    req.Model,
	})
	if err != nil {
		summary = truncate(content, 3000)
	}

	rawLen := len(content)
	sumLen := len(summary)

	return &FetchResponse{
		URL:                req.URL,
		Summary:            summary,
		CacheHit:           cacheHit,
		SourceLengthChars:  rawLen,
		SummaryLengthChars: sumLen,
		TokenStats: &TokenStats{
			RawEstimatedTokens:     estimateTokens(content),
			SummaryEstimatedTokens: estimateTokens(summary),
			ReductionPercent:       reductionPct(rawLen, sumLen),
		},
	}, nil
}

func (s *Service) Research(ctx context.Context, req ResearchRequest) (*ResearchResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validateResearch(req); err != nil {
		return nil, err
	}
	if req.MaxResults == 0 {
		req.MaxResults = defaultMaxResults
	}

	results, err := search.Search(req.Query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	limit := req.MaxResults
	if limit > len(results) {
		limit = len(results)
	}

	focus := req.Focus
	if focus == "" {
		focus = req.Query
	}

	sources := make([]SourceSummary, 0, limit)
	cacheHits := 0
	var summaries []string

	for i := 0; i < limit; i++ {
		r := results[i]
		var content string
		hit := false

		if cached, ok := cache.Get(r.URL); ok {
			content = cached
			hit = true
			cacheHits++
		} else {
			content, err = fetch.Fetch(r.URL)
			if err != nil {
				continue
			}
			_ = cache.Set(r.URL, content)
		}

		summary, err := summarize.Summarize(ctx, content, focus, summarize.Options{
			Provider: req.Provider,
			Model:    req.Model,
		})
		if err != nil {
			summary = truncate(content, 500)
		}

		sources = append(sources, SourceSummary{
			Title:    r.Title,
			URL:      r.URL,
			Summary:  summary,
			CacheHit: hit,
		})
		summaries = append(summaries, fmt.Sprintf("Source: %s\n%s", r.URL, summary))
	}

	// Synthesize final answer from all source summaries
	combined := strings.Join(summaries, "\n\n---\n\n")
	answer, err := summarize.Summarize(ctx, combined, fmt.Sprintf("Based on these sources, provide a comprehensive answer to: %s", req.Query), summarize.Options{
		Provider: req.Provider,
		Model:    req.Model,
	})
	if err != nil {
		answer = combined
	}

	return &ResearchResponse{
		Query:   req.Query,
		Answer:  answer,
		Sources: sources,
		Stats: ResearchStats{
			SearchedResults: len(results),
			FetchedPages:    len(sources),
			CacheHits:       cacheHits,
		},
	}, nil
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "\n[truncated]"
}
