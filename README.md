# web-research

Web research CLI for AI coding agents (Claude Code, Codex). Combines [TinyFish](https://tinyfish.ai) web search with [Groq](https://console.groq.com) (Llama 3.1) summarization into a single fast binary.

**Pipeline:** search → fetch → summarize → answer

## Why

- Native `WebSearch` dumps raw snippets into Claude context (expensive)
- Native `WebFetch` sends full HTML to Claude (very expensive)
- `wr` offloads both to external APIs, returning only the relevant summary

## Install

**Prerequisites:** Go 1.22+

```bash
git clone https://github.com/mrvarmazyar/web-research
cd web-research
go build -o ~/.local/bin/wr ./cmd/wr
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
