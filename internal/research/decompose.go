package research

import (
	"context"
	"strings"

	"github.com/mrvarmazyar/web-research/internal/summarize"
)

const decomposePrompt = "Generate 2-3 distinct search queries covering different angles of the above research question. Return only the queries, one per line, no numbering, no explanation."

func parseSubQueries(raw, fallback string) []string {
	lines := strings.Split(raw, "\n")
	queries := make([]string, 0, len(lines))
	for _, line := range lines {
		if q := strings.TrimSpace(line); q != "" {
			queries = append(queries, q)
		}
	}
	if len(queries) == 0 {
		return []string{fallback}
	}
	if len(queries) > 3 {
		queries = queries[:3]
	}
	return queries
}

func generateSubQueries(ctx context.Context, query string, opts summarize.Options) []string {
	raw, err := summarize.Summarize(ctx, query, decomposePrompt, opts)
	if err != nil || strings.TrimSpace(raw) == "" {
		return []string{query}
	}
	return parseSubQueries(raw, query)
}
