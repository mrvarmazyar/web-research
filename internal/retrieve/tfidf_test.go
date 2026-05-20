package retrieve

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	got := tokenize("Hello, World! This is a test.")
	want := []string{"hello", "world", "this", "is", "a", "test"}
	if len(got) != len(want) {
		t.Fatalf("tokenize() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tokenize()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScoreRelevance(t *testing.T) {
	chunks := []string{
		"golang concurrency goroutines channels",
		"python machine learning tensorflow keras",
		"golang channels select statement concurrency",
	}
	query := "golang concurrency"

	scores := scoreChunks(chunks, query)

	if len(scores) != len(chunks) {
		t.Fatalf("scoreChunks() returned %d scores, want %d", len(scores), len(chunks))
	}
	// chunk 0 and 2 should outscore chunk 1 (python/ML content)
	if scores[1] >= scores[0] {
		t.Errorf("python chunk scored %f >= golang chunk %f; expected lower", scores[1], scores[0])
	}
	if scores[1] >= scores[2] {
		t.Errorf("python chunk scored %f >= golang chunk %f; expected lower", scores[1], scores[2])
	}
}

func TestScoreEmptyQuery(t *testing.T) {
	chunks := []string{"first chunk", "second chunk", "third chunk"}
	scores := scoreChunks(chunks, "")
	for i, s := range scores {
		if s != 0 {
			t.Errorf("scores[%d] = %f, want 0 for empty query", i, s)
		}
	}
}

func TestScoreEmptyChunks(t *testing.T) {
	scores := scoreChunks([]string{}, "query")
	if len(scores) != 0 {
		t.Fatalf("scoreChunks() on empty input returned %d scores, want 0", len(scores))
	}
}
