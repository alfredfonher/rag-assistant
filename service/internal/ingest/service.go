package ingest

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	indexer    Indexer
	ingestRoot string
}

const maxDocumentSize int64 = 10 << 20

func New(indexer Indexer, ingestRoot string) (*Service, error) {
	root := strings.TrimSpace(ingestRoot)
	if root == "" {
		return nil, fmt.Errorf("ingest root is required")
	}

	return &Service{indexer: indexer, ingestRoot: root}, nil
}

func (s *Service) Ingest(ctx context.Context, request domain.IngestRequest) domain.IngestResponse {
	path := strings.TrimSpace(request.Path)
	if path == "" {
		return domain.IngestResponse{
			State: domain.DocumentStateUnsupported,
			Error: &domain.APIError{Code: "invalid_request", Message: "path is required"},
		}
	}
	if isAbsolutePath(path) || path == "." || hasTraversal(path) {
		return unsupported("invalid_path", "path must be relative to the configured ingest root")
	}

	if !isSupportedDocument(path) {
		return unsupported("unsupported_document", "only .txt, .md, and .markdown files are supported")
	}

	ingestRoot, err := s.resolveIngestRoot()
	if err != nil {
		return unsupported("ingest_root_unavailable", "configured ingest root is unavailable")
	}

	candidate := filepath.Join(ingestRoot, filepath.Clean(path))
	relative, err := filepath.Rel(ingestRoot, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return unsupported("invalid_path", "path must remain within the configured ingest root")
	}

	canonicalCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		if os.IsNotExist(err) {
			return unsupported("document_not_found", "document was not found in the configured ingest root")
		}
		return unsupported("document_unreadable", "document could not be resolved")
	}
	resolvedRelative, err := filepath.Rel(ingestRoot, canonicalCandidate)
	if err != nil || resolvedRelative == ".." || strings.HasPrefix(resolvedRelative, ".."+string(filepath.Separator)) {
		return unsupported("path_outside_ingest_root", "path must remain within the configured ingest root")
	}
	if !isSupportedDocument(canonicalCandidate) {
		return unsupported("unsupported_document", "only .txt, .md, and .markdown files are supported")
	}

	file, err := os.Open(canonicalCandidate)
	if err != nil {
		if os.IsNotExist(err) {
			return unsupported("document_not_found", "document was not found in the configured ingest root")
		}
		return unsupported("document_unreadable", "document could not be opened")
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return unsupported("document_unreadable", "document could not be inspected")
	}
	if !info.Mode().IsRegular() {
		return unsupported("document_not_regular", "document must be a regular file")
	}
	if info.Size() > maxDocumentSize {
		return unsupported("document_too_large", "document exceeds the 10 MiB size limit")
	}

	content, err := io.ReadAll(io.LimitReader(file, maxDocumentSize+1))
	if err != nil {
		return unsupported("document_unreadable", "document could not be read")
	}
	if int64(len(content)) > maxDocumentSize {
		return unsupported("document_too_large", "document exceeds the 10 MiB size limit")
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
			State:      domain.DocumentStateUnindexed,
			DocumentID: documentID,
			Error:      &domain.APIError{Code: code, Message: "document could not be added to the retrieval index"},
		}
	}

	return domain.IngestResponse{
		State:      domain.DocumentStateIndexed,
		DocumentID: documentID,
		Citations:  citations,
	}
}

func (s *Service) resolveIngestRoot() (string, error) {
	absoluteRoot, err := filepath.Abs(s.ingestRoot)
	if err != nil {
		return "", err
	}
	canonicalRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(canonicalRoot)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("ingest root is not a directory")
	}
	return canonicalRoot, nil
}

func unsupported(code, message string) domain.IngestResponse {
	return domain.IngestResponse{
		State: domain.DocumentStateUnsupported,
		Error: &domain.APIError{Code: code, Message: message},
	}
}

func hasTraversal(path string) bool {
	for _, part := range strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." {
			return true
		}
	}
	return false
}

func isAbsolutePath(path string) bool {
	if filepath.IsAbs(path) || filepath.VolumeName(path) != "" || strings.HasPrefix(path, "/") || strings.HasPrefix(path, `\`) {
		return true
	}
	return len(path) >= 3 && ((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) && path[1] == ':' && (path[2] == '/' || path[2] == '\\')
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
