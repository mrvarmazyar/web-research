# web-research

Token-efficient web research for AI agents.

Combines [TinyFish](https://tinyfish.ai) web search with [Groq](https://console.groq.com) (Llama 3.1) summarization. Use it as a CLI or as an MCP server.

Instead of dumping raw HTML, scripts, navbars, ads, and cookie banners into model context, it returns clean focused summaries with sources.

```
AI Agent / MCP Client
        ↓
wr-mcp (or wr CLI)
        ↓
search + fetch + clean + summarize + cache
        ↓
clean context with sources
```

## Why

- Native `WebSearch` dumps raw snippets into Claude context (expensive)
- Native `WebFetch` sends full HTML to Claude (very expensive)
- `wr` offloads both to external APIs, returning only the relevant summary

## Install

**Prerequisites:** Go 1.25+

```bash
git clone https://github.com/mrvarmazyar/web-research
cd web-research
go build -o ~/.local/bin/wr ./cmd/wr
go build -o ~/.local/bin/wr-mcp ./cmd/wr-mcp
```

Make sure `~/.local/bin` is in your PATH:
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bash_profile
source ~/.bash_profile
```

## Configuration

Add to `~/.bash_profile` or `~/.zshrc`:

```bash
export TINYFISH_API_KEY="..."   # https://agent.tinyfish.ai → Settings → API Keys
export GROQ_API_KEY="..."       # https://console.groq.com → API Keys (free tier)
export WR_CACHE_DAYS=7          # optional, cache TTL in days (default: 7)
```

Verify setup:
```bash
wr setup
```

Expected output:
```
OK       TINYFISH_API_KEY
OK       GROQ_API_KEY

All configured. Try: wr research "your query here"
```

## Usage

### `wr research`: full pipeline (recommended)

Searches, fetches top 3 results, summarizes each with Groq.

```bash
wr research "stripe webhook idempotency next.js app router"
wr research "tailwind v4 migration breaking changes"
wr research "next-intl v4 RTL Arabic configuration"
```

### `wr search`: search only

Returns titles, snippets, and URLs. Useful when you just need to find the right page.

```bash
wr search "golang http middleware pattern"
wr search "postgres JSONB index performance"
```

### `wr fetch`: fetch and summarize a URL

Fetches a URL, converts HTML to markdown, summarizes with Groq.

```bash
wr fetch "https://nextjs.org/docs/app/building-your-application/routing/middleware" "how to match locale prefixes"
wr fetch "https://docs.stripe.com/webhooks" "signature verification best practices"
```

## Token Reduction Benchmark

```
wr bench
```

```
URL                                           Raw HTML      Raw MD   wr output   vs HTML     vs MD
────────────────────────────────────────────────────────────────────────────────────────────────────
https://nextjs.org/docs/...                     1.2 MB     29.4 KB       760 B     99.9%     97.5%
https://docs.stripe.com/webhooks                1.4 MB     39.3 KB      11.8 KB     99.2%     70.0%
https://next-intl.dev/docs/...                202.5 KB     24.2 KB      11.8 KB     94.2%     51.3%
────────────────────────────────────────────────────────────────────────────────────────────────────
TOTAL                                           2.9 MB     92.8 KB      24.3 KB     99.2%     73.8%
```

Token estimates (chars / 4):

| Method | Tokens | Reduction vs HTML |
|--------|--------|-------------------|
| Raw HTML | 747,388 | — |
| Raw Markdown | 23,767 | 96.8% |
| `wr` output | 6,221 | **99.2%** |

## How It Works

```
wr research "query"
    │
    ├── TinyFish search API
    │   └── returns top 10 results (title, snippet, URL)
    │
    ├── For each of top 3 results:
    │   ├── Check local cache (~/.cache/web-research/)
    │   ├── If miss: fetch URL with net/http
    │   │   └── convert HTML → markdown (html-to-markdown)
    │   ├── Store in cache
    │   └── Groq API (llama-3.1-8b-instant)
    │       └── summarize markdown, extract relevant info
    │
    └── Print summaries
```

**Cache** is keyed by URL SHA-256, stored at `~/.cache/web-research/`, expires after `WR_CACHE_DAYS` days.

**Groq fallback:** if `GROQ_API_KEY` is unset, returns truncated raw markdown instead of a summary.

## MCP Server

`wr-mcp` exposes web-research as MCP tools for Claude Desktop, Cursor, Codex, and any MCP-compatible client.

| Tool | Purpose |
|------|---------|
| `web_search` | Search the web, return clean results |
| `web_fetch` | Fetch one URL and summarize it |
| `web_research` | Search + fetch top pages + synthesized answer with sources |

See [docs/mcp.md](docs/mcp.md) for client config examples (Claude Desktop, Cursor, generic).

## AI Agent Integration

### Claude Code

This repo includes a `SKILL.md` for Claude Code. Symlink it into your skills directory:

```bash
ln -s ~/web-research ~/.claude/skills/web-research
```

Then add to `~/.claude/CLAUDE.md`:

```markdown
## Web Research

For all web searches and URL fetching, use `wr`:
- Search: `wr search "<query>"`
- Fetch: `wr fetch "<url>" "<what to extract>"`
- Full research: `wr research "<query>"`

Skip for: GitHub repos/issues (use gh CLI), Jira/Confluence (use MCP tools), internal URLs.
```

Claude will automatically use `wr` instead of native `WebSearch`/`WebFetch`.

### OpenAI Codex

Create `~/.codex/web-research.md`:

```markdown
# Web Research

For all web searches and URL fetching, use the `wr` CLI instead of built-in search tools.

## Commands

\`\`\`bash
wr research "<query>"       # search + fetch top 3 + summarize (preferred)
wr search "<query>"         # search only → URLs + snippets
wr fetch "<url>" "<prompt>" # fetch single URL + summarize
\`\`\`

## When to Use
- Any research task, finding docs, "search for X", "look up X"
- Fetching any public documentation URL

## When to Skip
- GitHub repos/issues/PRs → use `gh` CLI
- Internal/authenticated URLs → use curl directly
```

Then reference it from `~/.codex/AGENTS.md`:

```markdown
@/Users/yourname/.codex/web-research.md
```

## Dependencies

| Package | Purpose |
|---------|---------|
| [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) | HTML → Markdown conversion |
| [TinyFish API](https://tinyfish.ai) | Live browser-rendered web search |
| [Groq API](https://console.groq.com) | Fast Llama 3.1 inference (free tier) |

## License

MIT
