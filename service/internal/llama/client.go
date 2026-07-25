// Package llama provides HTTP adapters for the rag-llama-server.
package llama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is an HTTP client for the rag-llama-server.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new llama-server client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// EmbedRequest is the request body for POST /v1/embeddings/query.
type EmbedRequest struct {
	Text string `json:"text"`
}

// EmbedResponse is the response from POST /v1/embeddings/query.
type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
}

// ChatMessage is a single message in a chat completion request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the request body for POST /v1/chat/completions.
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
	TopP        float64       `json:"top_p"`
	Stop        []string      `json:"stop,omitempty"`
}

// ChatChoice is a single choice in a chat completion response.
type ChatChoice struct {
	Message ChatMessage `json:"message"`
}

// ChatResponse is the response from POST /v1/chat/completions.
type ChatResponse struct {
	ID      string       `json:"id"`
	Choices []ChatChoice `json:"choices"`
}

// HealthResponse is the response from GET /healthz.
type HealthResponse struct {
	Status          string `json:"status"`
	EmbeddingLoaded bool   `json:"embedding_loaded"`
	LLMLoaded       bool   `json:"llm_loaded"`
}

// EmbedQuery calls POST /v1/embeddings/query to get a query embedding.
func (c *Client) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	body, err := json.Marshal(EmbedRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/v1/embeddings/query", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var embedResp EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	if len(embedResp.Embedding) == 0 {
		return nil, fmt.Errorf("llama-server returned an empty query embedding")
	}

	return embedResp.Embedding, nil
}

// EmbedDocuments calls POST /v1/embeddings to get document embeddings.
func (c *Client) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
	payload := map[string]any{
		"input": texts,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/v1/embeddings", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	if len(result.Data) != len(texts) {
		return nil, fmt.Errorf("llama-server returned %d document embeddings for %d inputs", len(result.Data), len(texts))
	}

	vectors := make([][]float64, len(result.Data))
	for i, d := range result.Data {
		vectors[i] = d.Embedding
	}
	return vectors, nil
}

// ChatCompletion calls POST /v1/chat/completions.
func (c *Client) ChatCompletion(ctx context.Context, req ChatRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/v1/chat/completions", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode chat response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in chat response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// Health calls GET /healthz.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	resp, err := c.do(ctx, http.MethodGet, "/healthz", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("decode health response: %w", err)
	}

	return &health, nil
}

func (c *Client) do(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llama-server request failed: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("llama-server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return resp, nil
}
