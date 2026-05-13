package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const tavilyEndpoint = "https://api.tavily.com/search"

type tavilyRequest struct {
	APIKey     string `json:"api_key"`
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

type tavilyResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

type tavilyResponse struct {
	Results []tavilyResult `json:"results"`
}

// SearchTavily queries the Tavily search API and maps results to the common Result type.
func SearchTavily(query string) ([]Result, error) {
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TAVILY_API_KEY not set — get a key at https://app.tavily.com")
	}

	body, err := json.Marshal(tavilyRequest{
		APIKey:     apiKey,
		Query:      query,
		MaxResults: 10,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Post(tavilyEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("tavily returned HTTP %d", resp.StatusCode)
	}

	var tr tavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	results := make([]Result, len(tr.Results))
	for i, r := range tr.Results {
		results[i] = Result{
			Position: i + 1,
			Title:    r.Title,
			Snippet:  r.Content,
			URL:      r.URL,
		}
	}
	return results, nil
}
