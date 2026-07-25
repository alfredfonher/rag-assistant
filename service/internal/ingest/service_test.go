package ingest

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/query"
)

func TestIngest(t *testing.T) {
	t.Parallel()

	t.Run("supported markdown becomes queryable", func(t *testing.T) {
		t.Parallel()

		retriever := query.NewMemoryRetriever()
		service := New(retriever)

		path := writeDocument(t, "guide.md", "# Title\n\nalpha beta\n\ngamma delta\n")
		response := service.Ingest(context.Background(), domain.IngestRequest{Path: path})

		if response.State != domain.DocumentStateIndexed {
			t.Fatalf("expected indexed state, got %q", response.State)
		}
		if response.DocumentID != "guide" {
			t.Fatalf("expected document id guide, got %q", response.DocumentID)
		}
		if response.NormalizedText == "" {
			t.Fatalf("expected normalized text")
		}

		queryService := query.New(retriever, nil)
		queryResponse := queryService.Query(context.Background(), domain.QueryRequest{Query: "delta"})
		if queryResponse.State != domain.QueryStateAnswered {
			t.Fatalf("expected answered query, got %q", queryResponse.State)
		}
		if len(queryResponse.Citations) == 0 || queryResponse.Citations[0].DocumentID != "guide" {
			t.Fatalf("expected citation for ingested document, got %#v", queryResponse.Citations)
		}
	})

	t.Run("unsupported extension fails", func(t *testing.T) {
		t.Parallel()

		service := New(query.NewMemoryRetriever())
		path := writeDocument(t, "notes.pdf", "alpha beta")
		response := service.Ingest(context.Background(), domain.IngestRequest{Path: path})

		if response.State != domain.DocumentStateUnsupported {
			t.Fatalf("expected unsupported state, got %q", response.State)
		}
		if response.Error == nil || response.Error.Code != "unsupported_document" {
			t.Fatalf("expected unsupported_document error, got %#v", response.Error)
		}
	})

	t.Run("empty content fails", func(t *testing.T) {
		t.Parallel()

		service := New(query.NewMemoryRetriever())
		path := writeDocument(t, "empty.txt", "\n\n  \n")
		response := service.Ingest(context.Background(), domain.IngestRequest{Path: path})

		if response.State != domain.DocumentStateUnsupported {
			t.Fatalf("expected unsupported state, got %q", response.State)
		}
		if response.Error == nil || response.Error.Code != "empty_document" {
			t.Fatalf("expected empty_document error, got %#v", response.Error)
		}
	})

	t.Run("embedding failure leaves document unindexed", func(t *testing.T) {
		t.Parallel()

		retriever, err := query.NewMemoryRetrieverWithProvider(failingEmbeddingProvider{failOn: "alpha beta", err: errors.New("embedding offline")})
		if err != nil {
			t.Fatalf("failed to create retriever: %v", err)
		}
		service := New(retriever)

		path := writeDocument(t, "broken.md", "alpha beta")
		response := service.Ingest(context.Background(), domain.IngestRequest{Path: path})

		if response.State != domain.DocumentStateUnindexed {
			t.Fatalf("expected unindexed state, got %q", response.State)
		}
		if response.Error == nil || response.Error.Code != "embedding_unavailable" {
			t.Fatalf("expected embedding_unavailable error, got %#v", response.Error)
		}

		queryResponse := query.New(retriever, nil).Query(context.Background(), domain.QueryRequest{Query: "alpha"})
		if queryResponse.State != domain.QueryStateInsufficientContext {
			t.Fatalf("expected unindexed document to stay unsearchable, got %q", queryResponse.State)
		}
	})
}

func TestPersistentIngestSurvivesRestart(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "index.json")
	provider := scriptedEmbeddingProvider{vectors: map[string][]float64{
		"alpha beta":  {1, 0},
		"gamma delta": {0, 1},
		"delta":       {0, 1},
	}}

	firstRetriever, err := query.NewPersistentMemoryRetrieverWithProvider(statePath, provider)
	if err != nil {
		t.Fatalf("failed to create persistent retriever: %v", err)
	}
	firstService := New(firstRetriever)
	path := writeDocument(t, "guide.md", "alpha beta\n\ngamma delta\n")
	if response := firstService.Ingest(context.Background(), domain.IngestRequest{Path: path}); response.State != domain.DocumentStateIndexed {
		t.Fatalf("expected indexed state, got %q", response.State)
	}

	secondRetriever, err := query.NewPersistentMemoryRetrieverWithProvider(statePath, provider)
	if err != nil {
		t.Fatalf("failed to reload persistent retriever: %v", err)
	}

	queryResponse := query.New(secondRetriever, nil).Query(context.Background(), domain.QueryRequest{Query: "delta"})
	if queryResponse.State != domain.QueryStateAnswered {
		t.Fatalf("expected answered query after restart, got %q", queryResponse.State)
	}
	if len(queryResponse.Citations) == 0 || queryResponse.Citations[0].DocumentID != "guide" {
		t.Fatalf("expected persisted citation after restart, got %#v", queryResponse.Citations)
	}
}

func writeDocument(t *testing.T, name, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test document: %v", err)
	}
	return path
}

type failingEmbeddingProvider struct {
	failOn string
	err    error
}

func (p failingEmbeddingProvider) EmbedDocument(_ context.Context, text string) ([]float64, error) {
	if text == p.failOn {
		return nil, p.err
	}

	return []float64{0, 0}, nil
}

func (p failingEmbeddingProvider) EmbedQuery(context.Context, string) ([]float64, error) {
	return []float64{0, 0}, nil
}

type scriptedEmbeddingProvider struct {
	vectors map[string][]float64
}

func (p scriptedEmbeddingProvider) EmbedDocument(ctx context.Context, text string) ([]float64, error) {
	return p.embed(ctx, text)
}

func (p scriptedEmbeddingProvider) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	return p.embed(ctx, text)
}

func (p scriptedEmbeddingProvider) embed(_ context.Context, text string) ([]float64, error) {
	if vector, ok := p.vectors[text]; ok {
		return append([]float64(nil), vector...), nil
	}

	return []float64{0, 0}, nil
}
