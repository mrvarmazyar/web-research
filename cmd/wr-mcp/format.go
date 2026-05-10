package main

import (
	"encoding/json"

	"github.com/mrvarmazyar/web-research/internal/research"
)

func formatSearch(r *research.SearchResponse) string {
	return toJSON(r)
}

func formatFetch(r *research.FetchResponse) string {
	return toJSON(r)
}

func formatResearch(r *research.ResearchResponse) string {
	return toJSON(r)
}

func toJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return `{"error":"failed to marshal response"}`
	}
	return string(b)
}
