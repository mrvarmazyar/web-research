package search

import "os"

// Search dispatches to the configured search backend.
//
// Selection logic:
//  1. If SEARCH_PROVIDER is "tavily", use Tavily.
//  2. If SEARCH_PROVIDER is "tinyfish" (or empty with no TAVILY_API_KEY), use TinyFish.
//  3. If SEARCH_PROVIDER is empty but TAVILY_API_KEY is set, use Tavily.
func Search(query string) ([]Result, error) {
	provider := os.Getenv("SEARCH_PROVIDER")

	switch provider {
	case "tavily":
		return SearchTavily(query)
	case "tinyfish":
		return SearchTinyFish(query)
	default:
		if os.Getenv("TAVILY_API_KEY") != "" {
			return SearchTavily(query)
		}
		return SearchTinyFish(query)
	}
}
