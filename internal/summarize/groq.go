package summarize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	groqEndpoint  = "https://api.groq.com/openai/v1/chat/completions"
	groqModel     = "llama-3.1-8b-instant"
	maxInputRunes = 12000 // ~3k tokens
	maxOutputTok  = 1024
)

type msg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model     string `json:"model"`
	Messages  []msg  `json:"messages"`
	MaxTokens int    `json:"max_tokens"`
}

type chatResponse struct {
	Choices []struct {
		Message msg `json:"message"`
	} `json:"choices"`
}

// Summarize extracts relevant content using Groq.
// Falls back to truncated raw content if GROQ_API_KEY is unset or request fails.
func Summarize(content, prompt string) (string, error) {
	runes := []rune(content)
	if len(runes) > maxInputRunes {
		content = string(runes[:maxInputRunes]) + "\n\n[content truncated]"
	}

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return content, nil
	}

	payload, _ := json.Marshal(chatRequest{
		Model: groqModel,
		Messages: []msg{
			{Role: "system", Content: "You are a concise technical research assistant. Extract and summarize relevant information from web content. Be direct and specific."},
			{Role: "user", Content: fmt.Sprintf("Content:\n%s\n\nTask: %s\n\nRespond concisely, only what's relevant.", content, prompt)},
		},
		MaxTokens: maxOutputTok,
	})

	req, err := http.NewRequest("POST", groqEndpoint, bytes.NewReader(payload))
	if err != nil {
		return content, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return content, fmt.Errorf("groq request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return content, fmt.Errorf("groq returned HTTP %d", resp.StatusCode)
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return content, err
	}
	if len(result.Choices) == 0 {
		return content, fmt.Errorf("empty response from groq")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
