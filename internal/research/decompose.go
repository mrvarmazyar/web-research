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
		if q := stripListPrefix(strings.TrimSpace(line)); q != "" {
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

// stripListPrefix removes common LLM list prefixes: "1.", "2)", "-", "*", "•".
func stripListPrefix(s string) string {
	// Numbered: "1. " or "1) " (1-2 digits)
	for i := 1; i <= 2 && i < len(s); i++ {
		if s[i-1] >= '0' && s[i-1] <= '9' {
			rest := s[i:]
			if strings.HasPrefix(rest, ". ") || strings.HasPrefix(rest, ") ") {
				return strings.TrimSpace(rest[2:])
			}
		} else {
			break
		}
	}
	// Bullets: "- ", "* ", "• "
	for _, prefix := range []string{"• ", "* ", "- "} {
		if strings.HasPrefix(s, prefix) {
			return strings.TrimSpace(s[len(prefix):])
		}
	}
	return s
}

func generateSubQueries(ctx context.Context, query string, opts summarize.Options) []string {
	raw, err := summarize.Summarize(ctx, query, decomposePrompt, opts)
	if err != nil || strings.TrimSpace(raw) == "" {
		return []string{query}
	}
	return parseSubQueries(raw, query)
}
