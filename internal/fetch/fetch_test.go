package fetch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJinaURL(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"https://example.com/page", "https://r.jina.ai/https://example.com/page", false},
		{"http://example.com/page", "https://r.jina.ai/http://example.com/page", false},
		{"ftp://example.com", "", true},
		{"not-a-url", "", true},
		{"https://", "", true}, // missing host
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := jinaURL(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("jinaURL(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("jinaURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// setupJinaMock redirects Jina calls to srv and restores jinaBase on cleanup.
func setupJinaMock(t *testing.T, srv *httptest.Server) {
	t.Helper()
	orig := jinaBase
	jinaBase = srv.URL + "/"
	t.Cleanup(func() { jinaBase = orig })
}

func TestFetchDirectAboveThreshold_NoJina(t *testing.T) {
	richContent := strings.Repeat("word ", minContentChars/4)
	jinaCalled := false

	directSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, richContent)
	}))
	defer directSrv.Close()

	jinaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jinaCalled = true
		fmt.Fprintln(w, "jina content")
	}))
	defer jinaSrv.Close()
	setupJinaMock(t, jinaSrv)

	_, err := Fetch(directSrv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if jinaCalled {
		t.Error("Jina should not be called when direct content is above threshold")
	}
}

func TestFetchSparseContent_FallsBackToJina(t *testing.T) {
	jinaContent := strings.Repeat("b ", minContentChars)

	directSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "tiny") // below threshold
	}))
	defer directSrv.Close()

	jinaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, jinaContent)
	}))
	defer jinaSrv.Close()
	setupJinaMock(t, jinaSrv)

	content, err := Fetch(directSrv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !strings.Contains(content, "b") {
		t.Errorf("expected Jina content, got: %q", content[:min(50, len(content))])
	}
}

func TestFetchSparseContent_KeepsDirectIfJinaIsShorter(t *testing.T) {
	directContent := strings.Repeat("x", 50) // sparse but longer than Jina

	directSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, directContent)
	}))
	defer directSrv.Close()

	jinaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "y") // shorter than direct
	}))
	defer jinaSrv.Close()
	setupJinaMock(t, jinaSrv)

	content, err := Fetch(directSrv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !strings.Contains(content, "x") {
		t.Errorf("expected direct content to be kept, got: %q", content[:min(50, len(content))])
	}
}

func TestFetchDirectError_CombinesErrors(t *testing.T) {
	jinaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "jina also failed")
	}))
	defer jinaSrv.Close()
	setupJinaMock(t, jinaSrv)

	_, err := Fetch("http://127.0.0.1:1/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "direct fetch failed") {
		t.Errorf("error missing 'direct fetch failed': %q", err.Error())
	}
	if !strings.Contains(err.Error(), "jina fallback failed") {
		t.Errorf("error missing 'jina fallback failed': %q", err.Error())
	}
}

func TestFetchJinaNon200IncludesBody(t *testing.T) {
	jinaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintln(w, "rate limit exceeded")
	}))
	defer jinaSrv.Close()
	setupJinaMock(t, jinaSrv)

	_, err := fetchJina("https://example.com/page")
	if err == nil {
		t.Fatal("expected error for non-200")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error missing status code: %q", err.Error())
	}
	if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("error missing response body: %q", err.Error())
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
