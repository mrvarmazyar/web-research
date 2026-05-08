package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrvarmazyar/web-research/internal/cache"
	"github.com/mrvarmazyar/web-research/internal/fetch"
	"github.com/mrvarmazyar/web-research/internal/search"
	"github.com/mrvarmazyar/web-research/internal/summarize"
)

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
	default:
		usage()
		os.Exit(1)
	}
}

func cmdSearch(query string) {
	results, err := search.Search(query)
	if err != nil {
		fatalf("search: %v", err)
	}
	for _, r := range results {
		fmt.Printf("%d. [%s] %s\n   %s\n   %s\n\n", r.Position, r.SiteName, r.Title, r.Snippet, r.URL)
	}
}

func cmdFetch(url, prompt string) {
	if cached, ok := cache.Get(url); ok {
		printSummary(cached, prompt)
		return
	}
	content, err := fetch.Fetch(url)
	if err != nil {
		fatalf("fetch: %v", err)
	}
	_ = cache.Set(url, content)
	printSummary(content, prompt)
}

func cmdResearch(query string) {
	fmt.Fprintf(os.Stderr, "→ searching: %s\n\n", query)
	results, err := search.Search(query)
	if err != nil {
		fatalf("search: %v", err)
	}

	limit := min(3, len(results))
	fmt.Fprintf(os.Stderr, "→ fetching top %d results\n\n", limit)

	for i := 0; i < limit; i++ {
		r := results[i]
		fmt.Printf("## %d. %s\n%s\n\n", i+1, r.Title, r.URL)

		var content string
		if cached, ok := cache.Get(r.URL); ok {
			content = cached
		} else {
			content, err = fetch.Fetch(r.URL)
			if err != nil {
				fmt.Printf("(fetch failed: %v)\n\n", err)
				continue
			}
			_ = cache.Set(r.URL, content)
		}

		summary, err := summarize.Summarize(content, query)
		if err != nil {
			fmt.Println(truncate(content, 500))
		} else {
			fmt.Println(summary)
		}
		fmt.Println()
	}
}

func cmdSetup() {
	type check struct {
		key  string
		hint string
	}
	checks := []check{
		{"TINYFISH_API_KEY", "https://agent.tinyfish.ai → Settings → API Keys"},
		{"GROQ_API_KEY", "https://console.groq.com → API Keys (free tier available)"},
	}

	allOk := true
	for _, c := range checks {
		if os.Getenv(c.key) == "" {
			fmt.Printf("MISSING  %s\n         %s\n\n", c.key, c.hint)
			allOk = false
		} else {
			fmt.Printf("OK       %s\n", c.key)
		}
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
  wr search <query>          Search web via tinyfish
  wr fetch <url> <prompt>    Fetch URL and summarize with Groq
  wr research <query>        Search + fetch top 3 + summarize
  wr setup                   Check env var configuration

ENV VARS
  TINYFISH_API_KEY    Web search  (tinyfish.ai, free)
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
