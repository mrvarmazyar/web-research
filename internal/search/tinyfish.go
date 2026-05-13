package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const endpoint = "https://api.search.tinyfish.ai"

type Result struct {
	Position int    `json:"position"`
	SiteName string `json:"site_name"`
	Title    string `json:"title"`
	Snippet  string `json:"snippet"`
	URL      string `json:"url"`
}

type response struct {
	Results []Result `json:"results"`
}

// SearchTinyFish queries the TinyFish search API.
func SearchTinyFish(query string) ([]Result, error) {
	apiKey := os.Getenv("TINYFISH_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TINYFISH_API_KEY not set — get free key at https://agent.tinyfish.ai")
	}

	u := endpoint + "?" + url.Values{"query": {query}}.Encode()
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("tinyfish returned HTTP %d", resp.StatusCode)
	}

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	return result.Results, nil
}
