package research

import (
	"fmt"
	"strings"
)

func validateSearch(req SearchRequest) error {
	if strings.TrimSpace(req.Query) == "" {
		return fmt.Errorf("query must not be empty")
	}
	if req.Limit < 0 {
		return fmt.Errorf("limit must be >= 0")
	}
	if req.Limit > maxSearchLimit {
		return fmt.Errorf("limit must be <= %d", maxSearchLimit)
	}
	return nil
}

func validateFetch(req FetchRequest) error {
	if strings.TrimSpace(req.URL) == "" {
		return fmt.Errorf("url must not be empty")
	}
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		return fmt.Errorf("url must start with http:// or https://")
	}
	return nil
}

func validateResearch(req ResearchRequest) error {
	if strings.TrimSpace(req.Query) == "" {
		return fmt.Errorf("query must not be empty")
	}
	if req.MaxResults < 0 {
		return fmt.Errorf("max_results must be >= 0")
	}
	if req.MaxResults > maxMaxResults {
		return fmt.Errorf("max_results must be <= %d", maxMaxResults)
	}
	return nil
}
