package llama

import (
	"context"
	"fmt"
	"strings"

	"rag-assistant/service/internal/domain"
)

const systemPrompt = "You are a helpful assistant that answers questions based on the provided context. Answer ONLY using the provided context. If the context does not contain enough information, say so briefly. Be concise and direct."

// Composer implements query.AnswerComposer using the llama-server LLM.
type Composer struct {
	client *Client
}

// NewComposer creates an LLM-backed answer composer.
func NewComposer(client *Client) *Composer {
	return &Composer{client: client}
}

// Compose generates an answer from the query and retrieved citations.
func (c *Composer) Compose(ctx context.Context, request domain.QueryRequest, citations []domain.Citation) (string, error) {
	ctxText := buildContext(citations)
	if strings.TrimSpace(ctxText) == "" {
		return "", fmt.Errorf("retrieved context is empty")
	}

	prompt := fmt.Sprintf("Context:\n%s\n\nQuestion: %s", ctxText, request.Query)

	content, err := c.client.ChatCompletion(ctx, ChatRequest{
		Model: "gemma3",
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   512,
		Temperature: 0.3,
		TopP:        0.9,
	})
	if err != nil {
		return "", fmt.Errorf("llm generation failed: %w", err)
	}

	answer := strings.TrimSpace(content)
	if answer == "" {
		return "", fmt.Errorf("llm returned empty answer")
	}

	return answer, nil
}

func buildContext(citations []domain.Citation) string {
	var parts []string
	seen := make(map[string]struct{})
	for _, c := range citations {
		snippet := strings.TrimSpace(c.Snippet)
		if snippet == "" {
			continue
		}
		if _, ok := seen[snippet]; ok {
			continue
		}
		seen[snippet] = struct{}{}
		parts = append(parts, snippet)
	}
	return strings.Join(parts, "\n\n")
}
