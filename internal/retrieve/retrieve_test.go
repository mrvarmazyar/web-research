package retrieve

import (
	"strings"
	"testing"
)

func TestTopKEmptyContent(t *testing.T) {
	got := TopK("", "query", 5)
	if len(got) != 0 {
		t.Fatalf("TopK on empty content = %d chunks, want 0", len(got))
	}
}

func TestTopKCount(t *testing.T) {
	content := "chunk one\n\nchunk two\n\nchunk three\n\nchunk four\n\nchunk five\n\nchunk six"
	got := TopK(content, "query", 3)
	if len(got) != 3 {
		t.Fatalf("TopK(k=3) = %d chunks, want 3", len(got))
	}
}

func TestTopKMoreThanAvailable(t *testing.T) {
	content := "only one chunk"
	got := TopK(content, "query", 10)
	if len(got) != 1 {
		t.Fatalf("TopK(k=10) on 1 chunk = %d, want 1", len(got))
	}
}

func TestTopKZeroK(t *testing.T) {
	content := "a\n\nb\n\nc\n\nd\n\ne\n\nf"
	got := TopK(content, "query", 0)
	if len(got) != 5 {
		t.Fatalf("TopK(k=0) should default to 5, got %d", len(got))
	}
}

func TestTopKNegativeK(t *testing.T) {
	content := "a\n\nb\n\nc\n\nd\n\ne\n\nf"
	got := TopK(content, "query", -1)
	if len(got) != 5 {
		t.Fatalf("TopK(k=-1) should default to 5, got %d", len(got))
	}
}

func TestTopKPreservesReadingOrder(t *testing.T) {
	content := "golang concurrency goroutines\n\npython baking recipes flour\n\ngolang channels select"
	got := TopK(content, "golang concurrency channels", 2)
	if len(got) != 2 {
		t.Fatalf("want 2 chunks, got %d", len(got))
	}
	if got[0].Index >= got[1].Index {
		t.Errorf("chunks not in reading order: indices %d, %d", got[0].Index, got[1].Index)
	}
}

func TestTopKJoinedSeparator(t *testing.T) {
	content := "first paragraph\n\nsecond paragraph\n\nthird paragraph"
	got := TopKJoined(content, "", 2)
	if !strings.Contains(got, "---") {
		t.Errorf("TopKJoined output missing separator, got: %q", got)
	}
}

func TestTopKEmptyQuery(t *testing.T) {
	content := "a\n\nb\n\nc\n\nd\n\ne\n\nf"
	got := TopK(content, "", 3)
	if len(got) != 3 {
		t.Fatalf("TopK(empty query) = %d, want 3", len(got))
	}
	if got[0].Index != 0 || got[1].Index != 1 || got[2].Index != 2 {
		t.Errorf("empty query: expected indices 0,1,2, got %d,%d,%d", got[0].Index, got[1].Index, got[2].Index)
	}
}
