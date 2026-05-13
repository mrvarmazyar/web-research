package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

	args := os.Args[2:]
	opts, remaining, err := parseSummaryOptions(args)
	if err != nil {
		fatalf("options: %v", err)
	}

	switch os.Args[1] {
	case "search":
		requireMinArgs(remaining, 1, "wr search <query>")
		cmdSearch(strings.Join(remaining, " "))
	case "fetch":
		requireMinArgs(remaining, 2, "wr fetch [--provider groq|copilot] [--model MODEL] <url> <prompt>")
		cmdFetch(remaining[0], strings.Join(remaining[1:], " "), opts)
	case "research":
		requireMinArgs(remaining, 1, "wr research [--provider groq|copilot] [--model MODEL] <query>")
		cmdResearch(strings.Join(remaining, " "), opts)
	case "setup":
		cmdSetup(opts)
	case "bench":
		cmdBench(remaining, opts)
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

func cmdFetch(url, prompt string, opts summaryOptions) {
	resp, err := svc.Fetch(context.Background(), research.FetchRequest{URL: url, Prompt: prompt, Provider: opts.Provider, Model: opts.Model})
	if err != nil {
		fatalf("fetch: %v", err)
	}
	fmt.Println(resp.Summary)
}

func cmdResearch(query string, opts summaryOptions) {
	fmt.Fprintf(os.Stderr, "→ searching: %s\n\n", query)
	resp, err := svc.Research(context.Background(), research.ResearchRequest{Query: query, Provider: opts.Provider, Model: opts.Model})
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

func cmdBench(args []string, opts summaryOptions) {
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

		// 3. wr output (fetch + configured summarizer)
		summary, _ := summarize.Summarize(context.Background(), mdContent, t.prompt, summarize.Options{Provider: opts.Provider, Model: opts.Model})
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

func cmdSetup(opts summaryOptions) {
	type check struct {
		key  string
		hint string
	}
	resolved := summarize.ResolveOptions(summarize.Options{Provider: opts.Provider, Model: opts.Model})
	if err := summarize.ValidateProvider(resolved.Provider); err != nil {
		fatalf("setup: %v", err)
	}
	checks := []check{
		{"TINYFISH_API_KEY", "https://agent.tinyfish.ai → Settings → API Keys"},
	}
	switch resolved.Provider {
	case summarize.ProviderCopilot:
		checks = append(checks,
			check{"COPILOT_CLI", "Install the standalone `copilot` CLI so `wr` can invoke it"},
		)
	case summarize.ProviderGroq:
		checks = append(checks,
			check{"GROQ_API_KEY", "https://console.groq.com → API Keys (free tier available)"},
		)
	}

	allOk := true
	for _, c := range checks {
		if !hasConfiguredValue(c.key) {
			fmt.Printf("MISSING  %s\n         %s\n\n", c.key, c.hint)
			allOk = false
		} else {
			fmt.Printf("OK       %s\n", c.key)
		}
	}
	if resolved.Provider == summarize.ProviderCopilot {
		switch {
		case hasConfiguredValue("COPILOT_ENV_AUTH"):
			fmt.Printf("OK       COPILOT_AUTH\n")
		case hasConfiguredValue("COPILOT_CLI"):
			fmt.Printf("INFO     COPILOT_AUTH\n         No token env vars detected. If you've already run `copilot login`, that may still work; the login session cannot be verified non-interactively.\n")
			allOk = false
		default:
			fmt.Printf("MISSING  COPILOT_AUTH\n         Authenticate with `copilot login` or export COPILOT_GITHUB_TOKEN, GH_TOKEN, or GITHUB_TOKEN\n\n")
			allOk = false
		}
	}
	if resolved.Model != "" {
		fmt.Printf("OK       summarizer model = %s\n", resolved.Model)
	}
	fmt.Printf("OK       summarizer provider = %s\n", resolved.Provider)

	fmt.Println()
	if allOk {
		fmt.Println("All configured. Try: wr research \"your query here\"")
	} else {
		fmt.Println("Add missing keys to ~/.zshrc, then: source ~/.zshrc && wr setup")
	}
}

func usage() {
	fmt.Print(`wr — web research CLI

COMMANDS
  wr search <query>          Search web via tinyfish
  wr fetch [--provider groq|copilot] [--model MODEL] <url> <prompt>
                            Fetch URL and summarize with chosen provider
  wr research [--provider groq|copilot] [--model MODEL] <query>
                            Search + fetch top 3 + summarize
  wr setup [--provider groq|copilot] [--model MODEL]
                            Check env var configuration
  wr bench [--provider groq|copilot] [--model MODEL]
                            Measure token savings vs raw HTML/MD
  wr bench [--provider groq|copilot] [--model MODEL] <url> <prompt>
                            Benchmark a specific URL

ENV VARS
  TINYFISH_API_KEY    Web search  (tinyfish.ai, free)
  WR_SUMMARIZER_PROVIDER  Summarizer provider default (groq|copilot)
  WR_SUMMARIZER_MODEL     Summarizer model override
  GROQ_API_KEY            Groq summarizer key
  COPILOT_MODEL           Copilot model default
  COPILOT_GITHUB_TOKEN    Copilot auth token (or use copilot login / GH_TOKEN / GITHUB_TOKEN)
  WR_CACHE_DAYS       Cache TTL in days (default: 7)

EXAMPLES
  wr search "next.js middleware locale redirect"
  wr fetch --provider groq --model llama-3.1-8b-instant https://nextjs.org/docs/app/building-your-application/routing/middleware "how to match locale paths"
  wr research --provider copilot --model gpt-5-mini "stripe webhook idempotency best practices"
`)
}

func requireMinArgs(args []string, n int, usage string) {
	if len(args) < n {
		fatalf("usage: %s", usage)
	}
}

func hasConfiguredValue(key string) bool {
	if key == "COPILOT_CLI" {
		_, err := exec.LookPath("copilot")
		return err == nil
	}
	if key == "COPILOT_ENV_AUTH" {
		for _, candidate := range []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
			if os.Getenv(candidate) != "" {
				return true
			}
		}
		return false
	}
	return os.Getenv(key) != ""
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
