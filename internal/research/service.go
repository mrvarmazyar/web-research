package research

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mrvarmazyar/web-research/internal/cache"
	"github.com/mrvarmazyar/web-research/internal/fetch"
	"github.com/mrvarmazyar/web-research/internal/retrieve"
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

	var summary string
	switch resolvedMode(req.Mode) {
	case "lossless":
		summary = content
	case "chunks":
		summary = retrieve.TopKJoined(content, req.Prompt, resolvedTopK(req.TopK))
	default:
		var err error
		summary, err = summarize.Summarize(ctx, content, req.Prompt, summarize.Options{
			Provider: req.Provider,
			Model:    req.Model,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: summarizer failed (%v); showing truncated content\n", err)
			summary = truncate(content, summarize.MaxFallbackChars)
		}
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

	type pageResult struct {
		source  SourceSummary
		summary string
		ok      bool
	}

	pageResults := make([]pageResult, limit)
	var cacheHits atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < limit; i++ {
		wg.Add(1)
		go func(i int, r search.Result) {
			defer wg.Done()

			var content string
			var hit bool

			if cached, ok := cache.Get(r.URL); ok {
				content = cached
				hit = true
				cacheHits.Add(1)
			} else {
				var fetchErr error
				content, fetchErr = fetch.Fetch(r.URL)
				if fetchErr != nil {
					return
				}
				_ = cache.Set(r.URL, content)
			}

			var summary string
			switch resolvedMode(req.Mode) {
			case "lossless":
				summary = content
			case "chunks":
				summary = retrieve.TopKJoined(content, focus, resolvedTopK(req.TopK))
			default:
				var sumErr error
				summary, sumErr = summarize.Summarize(ctx, content, focus, summarize.Options{
					Provider: req.Provider,
					Model:    req.Model,
				})
				if sumErr != nil {
					fmt.Fprintf(os.Stderr, "warning: summarizer failed (%v); showing truncated content\n", sumErr)
					summary = truncate(content, summarize.MaxFallbackChars)
				}
			}

			pageResults[i] = pageResult{
				source: SourceSummary{
					Title:    r.Title,
					URL:      r.URL,
					Summary:  summary,
					CacheHit: hit,
				},
				summary: fmt.Sprintf("Source: %s\n%s", r.URL, summary),
				ok:      true,
			}
		}(i, results[i])
	}

	wg.Wait()

	sources := make([]SourceSummary, 0, limit)
	var summaries []string
	for _, pr := range pageResults {
		if pr.ok {
			sources = append(sources, pr.source)
			summaries = append(summaries, pr.summary)
		}
	}

	combined := strings.Join(summaries, "\n\n---\n\n")
	var answer string
	if resolvedMode(req.Mode) == "summarize" {
		var err error
		answer, err = summarize.Summarize(ctx, combined, fmt.Sprintf("Based on these sources, provide a comprehensive answer to: %s", req.Query), summarize.Options{
			Provider: req.Provider,
			Model:    req.Model,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: summarizer failed (%v); showing combined source summaries\n", err)
			answer = combined
		}
	} else {
		answer = combined
	}

	return &ResearchResponse{
		Query:   req.Query,
		Answer:  answer,
		Sources: sources,
		Stats: ResearchStats{
			SearchedResults: len(results),
			FetchedPages:    len(sources),
			CacheHits:       int(cacheHits.Load()),
		},
	}, nil
}

func resolvedMode(m string) string {
	if m == "" {
		return "summarize"
	}
	return m
}

func resolvedTopK(k int) int {
	if k <= 0 {
		return 5
	}
	return k
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "\n[truncated]"
}
