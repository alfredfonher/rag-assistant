package llama

import (
	"context"
	"fmt"
)

// Embedder implements query.EmbeddingProvider using the llama-server HTTP API.
type Embedder struct {
	client *Client
}

// NewEmbedder creates an embedder backed by the llama-server.
func NewEmbedder(client *Client) *Embedder {
	return &Embedder{client: client}
}

// EmbedDocument returns a retrieval document embedding using nomic-embed-text.
func (e *Embedder) EmbedDocument(ctx context.Context, text string) ([]float64, error) {
	vectors, err := e.client.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) != 1 || len(vectors[0]) == 0 {
		return nil, fmt.Errorf("llama-server returned no document embedding")
	}
	return vectors[0], nil
}

// EmbedQuery returns a retrieval query embedding using nomic-embed-text.
func (e *Embedder) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	return e.client.EmbedQuery(ctx, text)
}
