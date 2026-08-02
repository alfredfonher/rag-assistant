package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/query"
)

func TestIngestAllowedNestedRelativeFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeDocument(t, root, filepath.Join("guides", "guide.md"), "# Title\r\n\r\nalpha beta\r\n\r\ngamma delta\r\n")
	retriever := query.NewMemoryRetriever()
	service := newService(t, retriever, root)

	response := service.Ingest(context.Background(), domain.IngestRequest{Path: filepath.Join("guides", "guide.md")})
	if response.State != domain.DocumentStateIndexed || response.DocumentID != "guide" {
		t.Fatalf("unexpected ingest response: %#v", response)
	}
	if len(response.Citations) == 0 || response.Citations[0].Snippet == "" {
		t.Fatalf("expected internal ingest citations to retain indexed text, got %#v", response.Citations)
	}

	queryResponse := query.New(retriever, nil).Query(context.Background(), domain.QueryRequest{Query: "delta"})
	if queryResponse.State != domain.QueryStateAnswered {
		t.Fatalf("expected answered query, got %q", queryResponse.State)
	}
	if len(queryResponse.Citations) == 0 || queryResponse.Citations[0].DocumentID != "guide" {
		t.Fatalf("expected citation for ingested document, got %#v", queryResponse.Citations)
	}
	if queryResponse.Citations[0].Snippet == "" {
		t.Fatalf("expected query citation snippet to remain available, got %#v", queryResponse.Citations)
	}

	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if strings.Contains(string(raw), "normalized_text") {
		t.Fatalf("public response leaked normalized content: %s", raw)
	}
}

func TestIngestRejectsUnsafeOrUnsupportedPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeDocument(t, root, "notes.pdf", "alpha beta")
	service := newService(t, query.NewMemoryRetriever(), root)

	tests := []struct {
		name     string
		path     string
		wantCode string
	}{
		{name: "empty", path: "  ", wantCode: "invalid_request"},
		{name: "dot", path: ".", wantCode: "invalid_path"},
		{name: "absolute", path: filepath.Join(root, "guide.md"), wantCode: "invalid_path"},
		{name: "parent traversal", path: filepath.Join("..", "guide.md"), wantCode: "invalid_path"},
		{name: "embedded traversal", path: "guides/../guide.md", wantCode: "invalid_path"},
		{name: "windows absolute", path: `C:\documents\guide.md`, wantCode: "invalid_path"},
		{name: "unsupported extension", path: "notes.pdf", wantCode: "unsupported_document"},
		{name: "missing file", path: "missing.md", wantCode: "document_not_found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := service.Ingest(context.Background(), domain.IngestRequest{Path: tt.path})
			if response.State != domain.DocumentStateUnsupported {
				t.Fatalf("expected unsupported state, got %q", response.State)
			}
			if response.Error == nil || response.Error.Code != tt.wantCode {
				t.Fatalf("expected %s error, got %#v", tt.wantCode, response.Error)
			}
			if strings.Contains(response.Error.Message, root) {
				t.Fatalf("error leaked ingest root: %q", response.Error.Message)
			}
		})
	}
}

func TestIngestRejectsSymlinkEscape(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := writeDocument(t, t.TempDir(), "outside.md", "secret")
	if err := os.Symlink(outside, filepath.Join(root, "escape.md")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	response := newService(t, query.NewMemoryRetriever(), root).Ingest(context.Background(), domain.IngestRequest{Path: "escape.md"})
	if response.Error == nil || response.Error.Code != "path_outside_ingest_root" {
		t.Fatalf("expected symlink escape rejection, got %#v", response)
	}
	if strings.Contains(response.Error.Message, outside) {
		t.Fatalf("error leaked outside path: %q", response.Error.Message)
	}
}

func TestIngestRejectsNonRegularAndOversizedFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "directory.md"), 0o700); err != nil {
		t.Fatalf("create directory fixture: %v", err)
	}
	writeDocument(t, root, "large.md", strings.Repeat("x", int(maxDocumentSize+1)))
	service := newService(t, query.NewMemoryRetriever(), root)

	tests := []struct {
		path     string
		wantCode string
	}{
		{path: "directory.md", wantCode: "document_not_regular"},
		{path: "large.md", wantCode: "document_too_large"},
	}
	for _, tt := range tests {
		t.Run(tt.wantCode, func(t *testing.T) {
			response := service.Ingest(context.Background(), domain.IngestRequest{Path: tt.path})
			if response.Error == nil || response.Error.Code != tt.wantCode {
				t.Fatalf("expected %s, got %#v", tt.wantCode, response)
			}
		})
	}
}

func TestNewAllowsMissingRootAndIngestReturnsStableError(t *testing.T) {
	t.Parallel()

	missingRoot := filepath.Join(t.TempDir(), "missing")
	service, err := New(query.NewMemoryRetriever(), missingRoot)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	response := service.Ingest(context.Background(), domain.IngestRequest{Path: "guide.md"})
	if response.State != domain.DocumentStateUnsupported {
		t.Fatalf("expected unsupported state, got %q", response.State)
	}
	if response.Error == nil || response.Error.Code != "ingest_root_unavailable" || response.Error.Message != "configured ingest root is unavailable" {
		t.Fatalf("unexpected ingest error: %#v", response.Error)
	}
	if strings.Contains(response.Error.Message, missingRoot) {
		t.Fatalf("error leaked ingest root: %q", response.Error.Message)
	}
}

func TestIngestEmptyContentAndEmbeddingFailure(t *testing.T) {
	t.Parallel()

	t.Run("empty content", func(t *testing.T) {
		root := t.TempDir()
		writeDocument(t, root, "empty.txt", "\n\n  \n")
		response := newService(t, query.NewMemoryRetriever(), root).Ingest(context.Background(), domain.IngestRequest{Path: "empty.txt"})
		if response.Error == nil || response.Error.Code != "empty_document" {
			t.Fatalf("expected empty_document error, got %#v", response.Error)
		}
	})

	t.Run("embedding failure", func(t *testing.T) {
		root := t.TempDir()
		writeDocument(t, root, "broken.md", "alpha beta")
		retriever, err := query.NewMemoryRetrieverWithProvider(failingEmbeddingProvider{failOn: "alpha beta", err: errors.New("embedding offline at /private/model")})
		if err != nil {
			t.Fatalf("create retriever: %v", err)
		}
		response := newService(t, retriever, root).Ingest(context.Background(), domain.IngestRequest{Path: "broken.md"})
		if response.State != domain.DocumentStateUnindexed || response.Error == nil || response.Error.Code != "embedding_unavailable" {
			t.Fatalf("unexpected response: %#v", response)
		}
		if strings.Contains(response.Error.Message, "/private/model") {
			t.Fatalf("index error leaked implementation detail: %q", response.Error.Message)
		}
	})
}

func TestPersistentIngestSurvivesRestart(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeDocument(t, root, "guide.md", "alpha beta\n\ngamma delta\n")
	statePath := filepath.Join(t.TempDir(), "index.json")
	provider := scriptedEmbeddingProvider{vectors: map[string][]float64{
		"alpha beta":  {1, 0},
		"gamma delta": {0, 1},
		"delta":       {0, 1},
	}}

	firstRetriever, err := query.NewPersistentMemoryRetrieverWithProvider(statePath, provider)
	if err != nil {
		t.Fatalf("create persistent retriever: %v", err)
	}
	if response := newService(t, firstRetriever, root).Ingest(context.Background(), domain.IngestRequest{Path: "guide.md"}); response.State != domain.DocumentStateIndexed {
		t.Fatalf("expected indexed state, got %q", response.State)
	}

	secondRetriever, err := query.NewPersistentMemoryRetrieverWithProvider(statePath, provider)
	if err != nil {
		t.Fatalf("reload persistent retriever: %v", err)
	}
	queryResponse := query.New(secondRetriever, nil).Query(context.Background(), domain.QueryRequest{Query: "delta"})
	if queryResponse.State != domain.QueryStateAnswered || len(queryResponse.Citations) == 0 || queryResponse.Citations[0].DocumentID != "guide" {
		t.Fatalf("expected persisted citation after restart, got %#v", queryResponse)
	}
}

func newService(t *testing.T, indexer Indexer, root string) *Service {
	t.Helper()
	service, err := New(indexer, root)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return service
}

func writeDocument(t *testing.T, root, name, content string) string {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create document directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write document: %v", err)
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
