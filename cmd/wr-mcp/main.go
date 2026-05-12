package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mrvarmazyar/web-research/internal/research"
)

var logger = log.New(os.Stderr, "[wr-mcp] ", log.LstdFlags)

func main() {
	svc := research.NewService()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "web-research",
		Version: "1.0.0",
	}, nil)

	addSearchTool(server, svc)
	addFetchTool(server, svc)
	addResearchTool(server, svc)

	logger.Println("starting web-research MCP server on stdio")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Fatalf("server error: %v", err)
	}
}

type searchArgs struct {
	Query string `json:"query" jsonschema:"Search query"`
	Limit int    `json:"limit,omitempty" jsonschema:"Max results (default 5 max 10)"`
}

func addSearchTool(server *mcp.Server, svc *research.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_search",
		Description: "Search the web and return clean title, URL, and snippet results.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args searchArgs) (*mcp.CallToolResult, any, error) {
		resp, err := svc.Search(ctx, research.SearchRequest{
			Query: args.Query,
			Limit: args.Limit,
		})
		if err != nil {
			return nil, nil, err
		}
		return textResult(formatSearch(resp)), nil, nil
	})
}

type fetchArgs struct {
	URL      string `json:"url" jsonschema:"URL to fetch (must start with http:// or https://)"`
	Prompt   string `json:"prompt,omitempty" jsonschema:"Focus prompt for summarization"`
	Provider string `json:"provider,omitempty" jsonschema:"Summarizer provider: groq or copilot"`
	Model    string `json:"model,omitempty" jsonschema:"Summarizer model override"`
}

func addFetchTool(server *mcp.Server, svc *research.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_fetch",
		Description: "Fetch a URL, clean the page, and return a focused summary.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args fetchArgs) (*mcp.CallToolResult, any, error) {
		resp, err := svc.Fetch(ctx, research.FetchRequest{
			URL:      args.URL,
			Prompt:   args.Prompt,
			Provider: args.Provider,
			Model:    args.Model,
		})
		if err != nil {
			return nil, nil, err
		}
		return textResult(formatFetch(resp)), nil, nil
	})
}

type researchArgs struct {
	Query      string `json:"query" jsonschema:"Research query"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"Pages to fetch (default 3 max 5)"`
	Focus      string `json:"focus,omitempty" jsonschema:"Focus prompt to guide summarization"`
	Provider   string `json:"provider,omitempty" jsonschema:"Summarizer provider: groq or copilot"`
	Model      string `json:"model,omitempty" jsonschema:"Summarizer model override"`
}

func addResearchTool(server *mcp.Server, svc *research.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_research",
		Description: "Search the web, fetch top pages, summarize them, and return a focused research answer with sources.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args researchArgs) (*mcp.CallToolResult, any, error) {
		resp, err := svc.Research(ctx, research.ResearchRequest{
			Query:      args.Query,
			MaxResults: args.MaxResults,
			Focus:      args.Focus,
			Provider:   args.Provider,
			Model:      args.Model,
		})
		if err != nil {
			return nil, nil, err
		}
		return textResult(formatResearch(resp)), nil, nil
	})
}

func textResult(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: s}},
	}
}
