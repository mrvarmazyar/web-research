package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

const (
	maxBytes        = 2 * 1024 * 1024 // 2MB
	minContentChars = 500              // below this, assume JS-rendered page and try Jina
)

var jinaBase = "https://r.jina.ai/"

var client = &http.Client{Timeout: 30 * time.Second}

// Fetch retrieves rawURL and returns clean markdown. If the page appears to be
// JS-rendered (content below minContentChars), it retries via the Jina reader
// API which executes JS server-side and returns clean markdown directly.
func Fetch(rawURL string) (string, error) {
	content, err := fetchDirect(rawURL)
	if err != nil {
		jina, jinaErr := fetchJina(rawURL)
		if jinaErr == nil {
			return jina, nil
		}
		return "", fmt.Errorf("direct fetch failed: %w; jina fallback failed: %v", err, jinaErr)
	}

	if len(strings.TrimSpace(content)) < minContentChars {
		if jina, jinaErr := fetchJina(rawURL); jinaErr == nil && len(jina) > len(content) {
			return jina, nil
		}
	}

	return content, nil
}

func fetchDirect(rawURL string) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; web-research/1.0; +https://github.com/mrvarmazyar/web-research)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d from %s", resp.StatusCode, rawURL)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}

	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(string(body))
	if err != nil {
		return stripTags(string(body)), nil
	}

	return markdown, nil
}

func jinaURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported URL scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("URL missing host")
	}
	return jinaBase + rawURL, nil
}

func fetchJina(rawURL string) (string, error) {
	readerURL, err := jinaURL(rawURL)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, readerURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/markdown")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; web-research/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("jina fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("jina HTTP %d for %s: %s",
			resp.StatusCode, rawURL, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return "", fmt.Errorf("jina read failed: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}

func stripTags(html string) string {
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}
