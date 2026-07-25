package query

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rag-assistant/service/internal/domain"
)

func TestPersistentMemoryRetrieverReloadsIndexedDocuments(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "index.json")
	provider := scriptedEmbeddingProvider{vectors: map[string][]float64{
		"alpha beta":  {1, 0},
		"gamma delta": {0, 1},
		"delta":       {0, 1},
	}}

	firstRetriever, err := NewPersistentMemoryRetrieverWithProvider(statePath, provider)
	if err != nil {
		t.Fatalf("failed to create persistent retriever: %v", err)
	}
	if _, err := firstRetriever.IndexDocument(context.Background(), "guide", "alpha beta\n\ngamma delta"); err != nil {
		t.Fatalf("failed to index document: %v", err)
	}

	secondRetriever, err := NewPersistentMemoryRetrieverWithProvider(statePath, provider)
	if err != nil {
		t.Fatalf("failed to reload persistent retriever: %v", err)
	}

	response := New(secondRetriever, nil).Query(context.Background(), domain.QueryRequest{Query: "delta"})
	if response.State != domain.QueryStateAnswered {
		t.Fatalf("expected answered state after reload, got %q", response.State)
	}
	if len(response.Citations) == 0 || response.Citations[0].DocumentID != "guide" {
		t.Fatalf("expected persisted citation after reload, got %#v", response.Citations)
	}
}

func TestPersistentMemoryRetrieverRejectsLegacyEmbeddingDimension(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "index.json")
	legacyVector := make([]float64, 16)
	legacyVector[0] = 1
	data, err := json.Marshal(persistedRetrieverState{
		Version: 1,
		Documents: map[string][]persistedChunk{
			"guide": {{DocumentID: "guide", ChunkID: "guide-chunk-1", Text: "legacy", Embedding: legacyVector, Dimension: 16}},
		},
	})
	if err != nil {
		t.Fatalf("failed to encode legacy state: %v", err)
	}
	if err := os.WriteFile(statePath, data, 0o600); err != nil {
		t.Fatalf("failed to write legacy state: %v", err)
	}

	retriever, err := NewPersistentMemoryRetrieverWithProvider(statePath, fixedDimensionProvider{dimension: 768})
	if err != nil {
		t.Fatalf("failed to load legacy state: %v", err)
	}
	_, err = retriever.Retrieve(context.Background(), domain.QueryRequest{Query: "legacy"})
	if err == nil || !strings.Contains(err.Error(), "re-ingest") || !strings.Contains(err.Error(), "16 dimensions") {
		t.Fatalf("expected actionable dimension mismatch, got %v", err)
	}
}

type fixedDimensionProvider struct {
	dimension int
}

func (p fixedDimensionProvider) EmbedDocument(context.Context, string) ([]float64, error) {
	return make([]float64, p.dimension), nil
}

func (p fixedDimensionProvider) EmbedQuery(context.Context, string) ([]float64, error) {
	vector := make([]float64, p.dimension)
	vector[0] = 1
	return vector, nil
}
