package retrieve

import (
	"sort"
	"strings"
)

const defaultK = 5

// Chunk is a scored segment of content with its original position.
type Chunk struct {
	Text  string
	Score float64
	Index int
}

// TopK splits content into chunks, scores each against query via TF-IDF cosine
// similarity, and returns the top-k chunks sorted by reading order.
// k <= 0 defaults to 5. Empty query returns first k chunks by position.
func TopK(content, query string, k int) []Chunk {
	if k <= 0 {
		k = defaultK
	}
	raw := splitChunks(content)
	if len(raw) == 0 {
		return nil
	}
	if k > len(raw) {
		k = len(raw)
	}

	chunks := make([]Chunk, len(raw))
	for i, text := range raw {
		chunks[i] = Chunk{Text: text, Index: i}
	}

	if strings.TrimSpace(query) == "" {
		return chunks[:k]
	}

	texts := make([]string, len(raw))
	for i, c := range chunks {
		texts[i] = c.Text
	}
	scores := scoreChunks(texts, query)
	for i := range chunks {
		chunks[i].Score = scores[i]
	}

	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Score > chunks[j].Score
	})
	topK := chunks[:k]

	sort.Slice(topK, func(i, j int) bool {
		return topK[i].Index < topK[j].Index
	})
	return topK
}

// TopKJoined calls TopK and joins the resulting chunks with "---" separators.
func TopKJoined(content, query string, k int) string {
	chunks := TopK(content, query, k)
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Text
	}
	return strings.Join(texts, "\n\n---\n\n")
}
