# MCP Server Setup

`wr-mcp` exposes web-research as MCP tools over stdio for Claude Desktop, Cursor, Codex, and any MCP-compatible client.

## Build

```bash
go build -o ~/.local/bin/wr-mcp ./cmd/wr-mcp
```

## Environment Variables

### Groq example

```bash
export TINYFISH_API_KEY="..."   # https://agent.tinyfish.ai → API Keys
export WR_SUMMARIZER_PROVIDER="groq"      # optional: groq (default) or copilot
export WR_SUMMARIZER_MODEL="llama-3.1-8b-instant"  # optional
export GROQ_API_KEY="..."       # required when provider=groq
export WR_CACHE_DAYS=7          # optional, default 7
```

### Copilot example

Install the standalone Copilot CLI first.

macOS / Linux (recommended):

```bash
brew install copilot-cli
```

macOS / Linux (shell installer fallback):

```bash
curl -fsSL https://gh.io/copilot-install | bash
```

Cross-platform fallback:

```bash
npm install -g @github/copilot
```

```bash
export TINYFISH_API_KEY="..."
export WR_SUMMARIZER_PROVIDER="copilot"
export WR_SUMMARIZER_MODEL="gpt-5-mini"
export COPILOT_MODEL="gpt-5-mini"         # optional when provider=copilot
# Authenticate Copilot via the standalone `copilot login` or export COPILOT_GITHUB_TOKEN / GH_TOKEN / GITHUB_TOKEN
export WR_CACHE_DAYS=7          # optional, default 7
```

## Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "web-research": {
      "command": "/Users/YOUR_USER/.local/bin/wr-mcp",
      "env": {
        "TINYFISH_API_KEY": "your_key",
        "WR_SUMMARIZER_PROVIDER": "groq",
        "WR_SUMMARIZER_MODEL": "llama-3.1-8b-instant",
        "GROQ_API_KEY": "your_key",
        "WR_CACHE_DAYS": "7"
      }
    }
  }
}
```

## Cursor

Edit `.cursor/mcp.json` in your project or `~/.cursor/mcp.json` globally:

```json
{
  "mcpServers": {
    "web-research": {
      "command": "/Users/YOUR_USER/.local/bin/wr-mcp",
      "env": {
        "TINYFISH_API_KEY": "your_key",
        "WR_SUMMARIZER_PROVIDER": "copilot",
        "WR_SUMMARIZER_MODEL": "gpt-5-mini"
      }
    }
  }
}
```

## Generic MCP client

Any client supporting the stdio MCP transport:

```json
{
  "command": "/Users/YOUR_USER/.local/bin/wr-mcp",
  "transport": "stdio",
  "env": {
    "TINYFISH_API_KEY": "your_key",
    "WR_SUMMARIZER_PROVIDER": "groq",
    "WR_SUMMARIZER_MODEL": "llama-3.1-8b-instant",
    "GROQ_API_KEY": "your_key"
  }
}
```

## Tools

| Tool | Description |
|------|-------------|
| `web_search` | Search the web, returns title + URL + snippet |
| `web_fetch` | Fetch one URL, return focused summary; accepts optional `provider` and `model` |
| `web_research` | Search + fetch top pages + synthesized answer with sources; accepts optional `provider` and `model` |

## Logs

All server logs go to stderr. stdout is reserved for MCP protocol messages.
