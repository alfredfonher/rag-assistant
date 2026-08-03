package query

import (
	"context"
	"errors"
	"sync"

	"rag-assistant/service/internal/domain"
)

var (
	ErrEmptyDocument = errors.New("document content is empty")
	ErrMutationStale = errors.New("mutation token is stale")
	ErrMutationUsed  = errors.New("mutation token is used")
)

type MutationToken struct {
	retriever  *MemoryRetriever
	documentID string
	prior      []memoryChunk
	existed    bool
	revision   uint64
	mu         sync.Mutex
	used       bool
}

func (r *MemoryRetriever) ReplaceDocument(ctx context.Context, documentID, content string) ([]domain.Citation, *MutationToken, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	chunks := chunkDocument(documentID, content)
	if len(chunks) == 0 {
		return nil, nil, ErrEmptyDocument
	}
	indexed, err := r.embedChunks(ctx, chunks)
	if err != nil {
		return nil, nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	prior, existed := r.documents[documentID]
	updated := cloneDocuments(r.documents)
	updated[documentID] = cloneChunks(indexed)
	if err := r.save(ctx, updated); err != nil {
		return nil, nil, err
	}
	r.documents = updated
	r.revision++
	return citationsFromChunks(indexed), &MutationToken{retriever: r, documentID: documentID, prior: cloneChunks(prior), existed: existed, revision: r.revision}, nil
}

func (r *MemoryRetriever) DeleteDocument(ctx context.Context, documentID string) (*MutationToken, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	prior, existed := r.documents[documentID]
	if !existed {
		return nil, nil
	}
	updated := cloneDocuments(r.documents)
	delete(updated, documentID)
	if err := r.save(ctx, updated); err != nil {
		return nil, err
	}
	r.documents = updated
	r.revision++
	return &MutationToken{retriever: r, documentID: documentID, prior: cloneChunks(prior), existed: true, revision: r.revision}, nil
}

func (t *MutationToken) Rollback(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.used {
		return ErrMutationUsed
	}
	r := t.retriever
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	t.used = true
	if r.revision != t.revision {
		return ErrMutationStale
	}
	updated := cloneDocuments(r.documents)
	if t.existed {
		updated[t.documentID] = cloneChunks(t.prior)
	} else {
		delete(updated, t.documentID)
	}
	if err := r.save(ctx, updated); err != nil {
		return err
	}
	r.documents = updated
	r.revision++
	return nil
}

func (r *MemoryRetriever) save(ctx context.Context, documents map[string][]memoryChunk) error {
	if r.store == nil {
		return nil
	}
	if err := r.store.Save(ctx, documents); err != nil {
		return errors.Join(ErrPersistenceUnavailable, err)
	}
	return nil
}
