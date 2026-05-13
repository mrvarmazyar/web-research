package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mrvarmazyar/web-research/internal/cache"
	"github.com/mrvarmazyar/web-research/internal/fetch"
	"github.com/mrvarmazyar/web-research/internal/research"
	"github.com/mrvarmazyar/web-research/internal/summarize"
)

var svc = research.NewService()

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "search":
		requireArgs(3, "wr search <query>")
		cmdSearch(strings.Join(os.Args[2:], " "))
	case "fetch":
		requireArgs(4, "wr fetch <url> <prompt>")
		cmdFetch(os.Args[2], strings.Join(os.Args[3:], " "))
	case "research":
		requireArgs(3, "wr research <query>")
		cmdResearch(strings.Join(os.Args[2:], " "))
	case "setup":
		cmdSetup()
	case "bench":
		cmdBench(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func cmdSearch(query string) {
	resp, err := svc.Search(context.Background(), research.SearchRequest{Query: query})
	if err != nil {
		fatalf("search: %v", err)
	}
	for i, r := range resp.Results {
		fmt.Printf("%d. [%s]\n   %s\n   %s\n\n", i+1, r.Title, r.Snippet, r.URL)
	}
}

func cmdFetch(url, prompt string) {
	resp, err := svc.Fetch(context.Background(), research.FetchRequest{URL: url, Prompt: prompt})
	if err != nil {
		fatalf("fetch: %v", err)
	}
	fmt.Println(resp.Summary)
}

func cmdResearch(query string) {
	fmt.Fprintf(os.Stderr, "→ searching: %s\n\n", query)
	resp, err := svc.Research(context.Background(), research.ResearchRequest{Query: query})
	if err != nil {
		fatalf("research: %v", err)
	}

	fmt.Println(resp.Answer)
	fmt.Printf("\n## Sources (%d fetched, %d cache hits)\n", resp.Stats.FetchedPages, resp.Stats.CacheHits)
	for i, src := range resp.Sources {
		fmt.Printf("%d. %s\n   %s\n\n", i+1, src.Title, src.URL)
	}
}

var defaultBenchTargets = []struct{ url, prompt string }{
	{"https://nextjs.org/docs/app/building-your-application/routing/middleware", "locale prefix matching"},
	{"https://docs.stripe.com/webhooks", "signature verification best practices"},
	{"https://next-intl.dev/docs/routing/configuration", "localePrefix always options"},
}

func cmdBench(args []string) {
	type row struct {
		url      string
		rawHTML  int
		rawMD    int
		wrOut    int
		fetchErr error
	}

	targets := defaultBenchTargets
	// Allow custom URL: wr bench <url> <prompt>
	if len(args) >= 2 {
		targets = []struct{ url, prompt string }{{args[0], strings.Join(args[1:], " ")}}
	}

	fmt.Fprintf(os.Stderr, "Running benchmark on %d URL(s)...\n\n", len(targets))

	rows := make([]row, 0, len(targets))
	httpClient := &http.Client{Timeout: 30 * time.Second}

	for _, t := range targets {
		r := row{url: t.url}

		// 1. Raw HTML size
		req, _ := http.NewRequest("GET", t.url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; wr-bench/1.0)")
		resp, err := httpClient.Do(req)
		if err != nil {
			r.fetchErr = err
			rows = append(rows, r)
			continue
		}
		rawBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
		resp.Body.Close()
		r.rawHTML = len(rawBytes)

		// 2. Raw markdown (fetch without Groq)
		mdContent, err := fetch.Fetch(t.url)
		if err != nil {
			r.fetchErr = err
			rows = append(rows, r)
			continue
		}
		_ = cache.Set(t.url, mdContent)
		r.rawMD = len(mdContent)

		// 3. wr output (fetch + Groq summary)
		summary, _ := summarize.Summarize(mdContent, t.prompt)
		r.wrOut = len(summary)

		rows = append(rows, r)
	}

	// Print table
	sep := strings.Repeat("─", 100)
	fmt.Printf("%-42s  %10s  %10s  %10s  %8s  %8s\n",
		"URL", "Raw HTML", "Raw MD", "wr output", "vs HTML", "vs MD")
	fmt.Println(sep)

	var totalHTML, totalMD, totalWR int
	for _, r := range rows {
		short := r.url
		if len(short) > 42 {
			short = short[:39] + "..."
		}
		if r.fetchErr != nil {
			fmt.Printf("%-42s  ERROR: %v\n", short, r.fetchErr)
			continue
		}
		savHTML := pct(r.wrOut, r.rawHTML)
		savMD := pct(r.wrOut, r.rawMD)
		fmt.Printf("%-42s  %10s  %10s  %10s  %7.1f%%  %7.1f%%\n",
			short,
			humanBytes(r.rawHTML),
			humanBytes(r.rawMD),
			humanBytes(r.wrOut),
			savHTML,
			savMD,
		)
		totalHTML += r.rawHTML
		totalMD += r.rawMD
		totalWR += r.wrOut
	}

	fmt.Println(sep)
	fmt.Printf("%-42s  %10s  %10s  %10s  %7.1f%%  %7.1f%%\n",
		"TOTAL",
		humanBytes(totalHTML),
		humanBytes(totalMD),
		humanBytes(totalWR),
		pct(totalWR, totalHTML),
		pct(totalWR, totalMD),
	)

	fmt.Printf("\n≈ tokens (chars ÷ 4):\n")
	fmt.Printf("  Raw HTML:  %d tokens\n", totalHTML/4)
	fmt.Printf("  Raw MD:    %d tokens\n", totalMD/4)
	fmt.Printf("  wr output: %d tokens\n", totalWR/4)
	fmt.Printf("  Saved vs raw MD:   ~%d tokens (%.1f%%)\n",
		(totalMD-totalWR)/4, pct(totalWR, totalMD))
	fmt.Printf("  Saved vs raw HTML: ~%d tokens (%.1f%%)\n",
		(totalHTML-totalWR)/4, pct(totalWR, totalHTML))
}

func pct(out, in int) float64 {
	if in == 0 {
		return 0
	}
	return (1 - float64(out)/float64(in)) * 100
}

func humanBytes(b int) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/1024/1024)
	case b >= 1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func cmdSetup() {
	allOk := true

	// Search provider: at least one of TINYFISH_API_KEY or TAVILY_API_KEY must be set.
	tinyfish := os.Getenv("TINYFISH_API_KEY")
	tavily := os.Getenv("TAVILY_API_KEY")
	switch {
	case tinyfish != "" && tavily != "":
		provider := os.Getenv("SEARCH_PROVIDER")
		if provider == "" {
			provider = "tavily"
		}
		fmt.Printf("OK       SEARCH_PROVIDER (%s) — both keys set\n", provider)
	case tavily != "":
		fmt.Printf("OK       SEARCH_PROVIDER (tavily)\n")
	case tinyfish != "":
		fmt.Printf("OK       SEARCH_PROVIDER (tinyfish)\n")
	default:
		fmt.Printf("MISSING  TINYFISH_API_KEY or TAVILY_API_KEY (need at least one)\n")
		fmt.Printf("         https://app.tavily.com → API Keys (1000 free credits/month)\n")
		fmt.Printf("         https://agent.tinyfish.ai → Settings → API Keys\n\n")
		allOk = false
	}

	// Other required keys.
	if os.Getenv("GROQ_API_KEY") == "" {
		fmt.Printf("MISSING  GROQ_API_KEY\n         https://console.groq.com → API Keys (free tier available)\n\n")
		allOk = false
	} else {
		fmt.Printf("OK       GROQ_API_KEY\n")
	}

	fmt.Println()
	if allOk {
		fmt.Println("All configured. Try: wr research \"your query here\"")
	} else {
		fmt.Println("Add missing keys to ~/.zshrc, then: source ~/.zshrc && wr setup")
	}
}

func printSummary(content, prompt string) {
	summary, err := summarize.Summarize(content, prompt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: summarize failed, showing raw content")
		fmt.Println(truncate(content, 3000))
		return
	}
	fmt.Println(summary)
}

func usage() {
	fmt.Print(`wr — web research CLI

COMMANDS
  wr search <query>          Search web (provider auto-detected)
  wr fetch <url> <prompt>    Fetch URL and summarize with Groq
  wr research <query>        Search + fetch top 3 + summarize
  wr setup                   Check env var configuration
  wr bench                   Measure token savings vs raw HTML/MD
  wr bench <url> <prompt>    Benchmark a specific URL

ENV VARS
  TINYFISH_API_KEY    Web search  (tinyfish.ai, free)
  TAVILY_API_KEY      Web search  (tavily.com, 1000 free credits/month)
  SEARCH_PROVIDER     Search backend: 'tavily' or 'tinyfish' (auto-detected if omitted)
  GROQ_API_KEY        Summarizer  (console.groq.com, free)
  WR_CACHE_DAYS       Cache TTL in days (default: 7)

EXAMPLES
  wr search "next.js middleware locale redirect"
  wr fetch https://nextjs.org/docs/app/building-your-application/routing/middleware "how to match locale paths"
  wr research "stripe webhook idempotency best practices"
`)
}

func requireArgs(n int, usage string) {
	if len(os.Args) < n {
		fatalf("usage: %s", usage)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "\n[truncated]"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
