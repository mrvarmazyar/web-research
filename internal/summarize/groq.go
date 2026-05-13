package summarize

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	groqEndpoint = "https://api.groq.com/openai/v1/chat/completions"
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

func summarizeWithGroq(ctx context.Context, content, prompt, model string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return fallbackContent(content), nil
	}

	payload, _ := json.Marshal(chatRequest{
		Model: model,
		Messages: []msg{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: buildUserPrompt(content, prompt)},
		},
		MaxTokens: maxOutputTokens,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", groqEndpoint, bytes.NewReader(payload))
	if err != nil {
		return content, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fallbackContent(content), fmt.Errorf("groq request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fallbackContent(content), fmt.Errorf("groq returned HTTP %d", resp.StatusCode)
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fallbackContent(content), err
	}
	if len(result.Choices) == 0 {
		return fallbackContent(content), fmt.Errorf("empty response from groq")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
