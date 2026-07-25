package ingest

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/query"
)

type Indexer interface {
	IndexDocument(ctx context.Context, documentID, content string) ([]domain.Citation, error)
}

type Service struct {
	indexer Indexer
}

func New(indexer Indexer) *Service {
	return &Service{indexer: indexer}
}

func (s *Service) Ingest(ctx context.Context, request domain.IngestRequest) domain.IngestResponse {
	path := strings.TrimSpace(request.Path)
	if path == "" {
		return domain.IngestResponse{
			State: domain.DocumentStateUnsupported,
			Error: &domain.APIError{Code: "invalid_request", Message: "path is required"},
		}
	}

	if !isSupportedDocument(path) {
		return domain.IngestResponse{
			State: domain.DocumentStateUnsupported,
			Error: &domain.APIError{Code: "unsupported_document", Message: "only .txt, .md, and .markdown files are supported"},
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return domain.IngestResponse{
			State: domain.DocumentStateUnsupported,
			Error: &domain.APIError{Code: "document_unreadable", Message: err.Error()},
		}
	}

	normalized := normalizeContent(string(content))
	if normalized == "" {
		return domain.IngestResponse{
			State: domain.DocumentStateUnsupported,
			Error: &domain.APIError{Code: "empty_document", Message: "document contains no text"},
		}
	}

	documentID := documentIDFromPath(path)
	citations, err := s.indexer.IndexDocument(ctx, documentID, normalized)
	if err != nil {
		code := "embedding_unavailable"
		if errors.Is(err, query.ErrPersistenceUnavailable) {
			code = "index_persistence_unavailable"
		}
		return domain.IngestResponse{
			State:          domain.DocumentStateUnindexed,
			DocumentID:     documentID,
			NormalizedText: normalized,
			Error:          &domain.APIError{Code: code, Message: err.Error()},
		}
	}

	return domain.IngestResponse{
		State:          domain.DocumentStateIndexed,
		DocumentID:     documentID,
		NormalizedText: normalized,
		Citations:      citations,
	}
}

func isSupportedDocument(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".txt", ".md", ".markdown":
		return true
	default:
		return false
	}
}

func normalizeContent(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	return strings.TrimSpace(content)
}

func documentIDFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
