# MCP Server Setup

`wr-mcp` exposes web-research as MCP tools over stdio for Claude Desktop, Cursor, Codex, and any MCP-compatible client.

## Build

```bash
go build -o ~/.local/bin/wr-mcp ./cmd/wr-mcp
```

## Environment Variables

```bash
export TINYFISH_API_KEY="..."   # https://agent.tinyfish.ai → API Keys
export GROQ_API_KEY="..."       # https://console.groq.com → API Keys (free)
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
        "GROQ_API_KEY": "your_key"
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
    "GROQ_API_KEY": "your_key"
  }
}
```

## Tools

| Tool | Description |
|------|-------------|
| `web_search` | Search the web, returns title + URL + snippet |
| `web_fetch` | Fetch one URL, return focused summary |
| `web_research` | Search + fetch top pages + synthesized answer with sources |

## Logs

All server logs go to stderr. stdout is reserved for MCP protocol messages.
