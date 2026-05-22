# web-research

Token-efficient web research for AI agents.

Combines [TinyFish](https://tinyfish.ai) web search with configurable summarization providers (Groq and GitHub Copilot CLI). Use it as a CLI or as an MCP server.

Instead of dumping raw HTML, scripts, navbars, ads, and cookie banners into model context, it searches, fetches, and returns clean focused content with sources.

```
AI Agent / MCP Client
        ↓
wr-mcp (or wr CLI)
        ↓
decompose → parallel search → parallel fetch (+ Jina fallback) → summarize → cache
        ↓
clean context with sources
```

## Why

- Native `WebSearch` dumps raw snippets into Claude context (expensive)
- Native `WebFetch` sends full HTML to Claude (very expensive)
- `wr` offloads both to external APIs, returning only the relevant content

## Install

### From GitHub Releases (recommended)

Download a pre-built binary for your platform from the [releases page](https://github.com/mrvarmazyar/web-research/releases):

```bash
# macOS Apple Silicon
curl -L https://github.com/mrvarmazyar/web-research/releases/latest/download/wr-darwin-arm64 -o ~/.local/bin/wr && chmod +x ~/.local/bin/wr
curl -L https://github.com/mrvarmazyar/web-research/releases/latest/download/wr-mcp-darwin-arm64 -o ~/.local/bin/wr-mcp && chmod +x ~/.local/bin/wr-mcp

# macOS Intel
curl -L https://github.com/mrvarmazyar/web-research/releases/latest/download/wr-darwin-amd64 -o ~/.local/bin/wr && chmod +x ~/.local/bin/wr
curl -L https://github.com/mrvarmazyar/web-research/releases/latest/download/wr-mcp-darwin-amd64 -o ~/.local/bin/wr-mcp && chmod +x ~/.local/bin/wr-mcp

# Linux amd64
curl -L https://github.com/mrvarmazyar/web-research/releases/latest/download/wr-linux-amd64 -o ~/.local/bin/wr && chmod +x ~/.local/bin/wr
curl -L https://github.com/mrvarmazyar/web-research/releases/latest/download/wr-mcp-linux-amd64 -o ~/.local/bin/wr-mcp && chmod +x ~/.local/bin/wr-mcp
```

### Build from source

**Prerequisites:** Go 1.25+

```bash
git clone https://github.com/mrvarmazyar/web-research
cd web-research
go build -o ~/.local/bin/wr ./cmd/wr
go build -o ~/.local/bin/wr-mcp ./cmd/wr-mcp
```

Make sure `~/.local/bin` is in your PATH:
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

## Configuration

### Groq (default)

```bash
export TINYFISH_API_KEY="..."        # https://agent.tinyfish.ai → Settings → API Keys
export GROQ_API_KEY="..."            # https://console.groq.com → API Keys (free tier available)
export WR_SUMMARIZER_PROVIDER="groq" # optional, groq is the default
export WR_SUMMARIZER_MODEL="llama-3.1-8b-instant"  # optional
export WR_CACHE_DAYS=7               # optional, cache TTL in days (default: 7)
```

### Copilot CLI (experimental)

Install a Copilot CLI that exposes a `copilot` binary, then:

```bash
copilot login

export TINYFISH_API_KEY="..."
export WR_SUMMARIZER_PROVIDER="copilot"
export WR_SUMMARIZER_MODEL="gpt-5-mini"  # optional
```

Verify setup:
```bash
wr setup
```

## Usage

### `wr research` — full pipeline (recommended)

Auto-expands your query into 2-3 sub-queries, searches each in parallel, fetches top unique results concurrently, and synthesizes a final answer. Live progress is printed to stderr.

```bash
wr research "stripe webhook idempotency next.js app router"
wr research "tailwind v4 migration breaking changes"
wr research --provider copilot --model gpt-5-mini "next-intl RTL Arabic configuration"
```

Output during a research call:
```
→ decomposing query...
→ searching [3 queries]: stripe webhook idempotency | stripe retry handling | stripe idempotency key
→ fetched: https://stripe.com/docs/webhooks
→ cache hit: https://stripe.com/docs/error-handling
→ fetched: https://docs.stripe.com/webhooks/best-practices
```

### `wr fetch` — fetch and summarize a URL

```bash
wr fetch "https://nextjs.org/docs/app/routing/middleware" "how to match locale prefixes"
wr fetch --provider copilot "https://docs.stripe.com/webhooks" "signature verification"
```

### `wr search` — search only

Returns titles, snippets, and URLs without fetching.

```bash
wr search "golang http middleware pattern"
wr search "postgres JSONB index performance"
```

## Retrieval Modes

All fetch and research commands support `--mode` to control how content is processed:

| Mode | LLM call | Use case |
|------|----------|----------|
| `summarize` | Yes (default) | Focused summary, maximum compression |
| `lossless` | No | Full clean markdown, no information loss |
| `chunks` | No | TF-IDF top-k sections scored against query |

```bash
# Return full page markdown, no LLM
wr fetch --mode lossless https://docs.stripe.com/webhooks "signature verification"

# Return top 5 most relevant sections
wr fetch --mode chunks --top-k 5 https://docs.stripe.com/webhooks "signature verification"

# Research without LLM summarization
wr research --mode lossless "stripe webhook best practices"
wr research --mode chunks --top-k 3 "stripe webhook idempotency"
```

## Token Reduction Benchmark

```
wr bench
```

```
URL                                           Raw HTML      Raw MD   wr output   vs HTML     vs MD
────────────────────────────────────────────────────────────────────────────────────────────────────
https://nextjs.org/docs/...                     1.2 MB     29.7 KB       397 B    100.0%     98.7%
https://docs.stripe.com/webhooks                1.4 MB     39.3 KB      3.0 KB     99.8%     92.5%
https://next-intl.dev/docs/...                202.5 KB     24.2 KB      2.9 KB     98.5%     87.8%
────────────────────────────────────────────────────────────────────────────────────────────────────
TOTAL                                           2.9 MB     93.2 KB      6.3 KB     99.8%     93.3%
```

Token estimates (chars / 4):

| Method | Tokens | Reduction vs raw MD |
|--------|--------|---------------------|
| Raw HTML | 752,671 | — |
| Raw Markdown | 23,851 | — |
| `wr` output | 1,608 | **93.3%** |

## How It Works

```
wr research "query"
    │
    ├── LLM: expand into 2-3 sub-queries
    │
    ├── Parallel TinyFish search for each sub-query
    │   └── deduplicate URLs (first-seen wins)
    │
    ├── For each of top N unique results (parallel):
    │   ├── Check local cache (~/.cache/web-research/)
    │   ├── If miss: fetch URL → convert HTML → markdown
    │   │   └── If sparse (<500 chars): Jina reader fallback
    │   │       (handles JS-rendered SPAs automatically)
    │   ├── Store in cache
    │   └── Process by mode: summarize | lossless | chunks
    │
    └── Synthesize final answer (summarize mode only)
```

**Cache** is keyed by URL SHA-256, stored at `~/.cache/web-research/`, expires after `WR_CACHE_DAYS` days.

**Jina fallback:** Pages that return sparse HTML (JS-rendered SPAs like Next.js docs, Vercel dashboard) are automatically re-fetched via `r.jina.ai`, which executes JavaScript server-side and returns clean markdown. No API key required.

**Fallbacks:**
- Summarizer failure → returns truncated raw markdown
- Query decomposition failure → falls back to original single query
- Direct fetch + Jina both fail → combined error reported

## MCP Server

`wr-mcp` exposes web-research as MCP tools for Claude Desktop, Cursor, Codex, and any MCP-compatible client.

| Tool | Description |
|------|-------------|
| `web_search` | Search the web, return clean results |
| `web_fetch` | Fetch a URL and return content (`mode`, `top_k`, `provider`, `model` optional) |
| `web_research` | Search + fetch + synthesized answer (`mode`, `top_k`, `provider`, `model` optional) |

### Claude Code

```bash
claude mcp add web-research ~/.local/bin/wr-mcp
```

### Claude Desktop / Cursor

Add to your MCP config:

```json
{
  "mcpServers": {
    "web-research": {
      "command": "/Users/yourname/.local/bin/wr-mcp"
    }
  }
}
```

See [docs/mcp.md](docs/mcp.md) for more client configuration examples.

## AI Agent Integration

### Claude Code

Add to `~/.claude/CLAUDE.md`:

```markdown
## Web Research

For all web searches and URL fetching, use the `wr` CLI:
- Search: `wr search "<query>"`
- Fetch: `wr fetch "<url>" "<what to extract>"`
- Full research: `wr research "<query>"`

Skip for: GitHub repos/issues (use gh CLI), Jira/Confluence (use MCP tools), internal URLs.
```

### OpenAI Codex

Create `~/.codex/web-research.md` and reference it from `~/.codex/AGENTS.md`:

```markdown
@/Users/yourname/.codex/web-research.md
```

## Dependencies

| Package | Purpose |
|---------|---------|
| [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) | HTML → Markdown conversion |
| [TinyFish API](https://tinyfish.ai) | Live browser-rendered web search |
| [Groq API](https://console.groq.com) | Groq-hosted summarization |
| [GitHub Copilot CLI](https://github.com/github/copilot-cli) | Optional local summarizer backend |
| [Jina Reader](https://r.jina.ai) | JS-rendered page fallback (no key required) |

## License

MIT
