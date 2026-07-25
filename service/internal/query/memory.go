package query

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"rag-assistant/service/internal/domain"
)

type MemoryRetriever struct {
	mu        sync.RWMutex
	provider  EmbeddingProvider
	documents map[string][]memoryChunk
	store     retrieverStore
}

type memoryChunk struct {
	documentID string
	chunkID    string
	text       string
	embedding  []float64
	dimension  int
}

func NewMemoryRetriever() *MemoryRetriever {
	retriever, _ := NewMemoryRetrieverWithProvider(NewDeterministicEmbeddingProvider(16))
	return retriever
}

func NewMemoryRetrieverWithProvider(provider EmbeddingProvider) (*MemoryRetriever, error) {
	return newMemoryRetriever(provider, nil)
}

func NewPersistentMemoryRetriever(path string) (*MemoryRetriever, error) {
	return NewPersistentMemoryRetrieverWithProvider(path, NewDeterministicEmbeddingProvider(16))
}

func NewPersistentMemoryRetrieverWithProvider(path string, provider EmbeddingProvider) (*MemoryRetriever, error) {
	return newMemoryRetriever(provider, newFileRetrieverStore(path))
}

func newMemoryRetriever(provider EmbeddingProvider, store retrieverStore) (*MemoryRetriever, error) {
	if provider == nil {
		provider = NewDeterministicEmbeddingProvider(16)
	}

	retriever := &MemoryRetriever{
		provider:  provider,
		documents: make(map[string][]memoryChunk),
		store:     store,
	}

	if store != nil {
		documents, err := store.Load(context.Background())
		if err != nil {
			return nil, err
		}
		retriever.documents = documents
	}

	return retriever, nil
}

func (r *MemoryRetriever) IndexDocument(ctx context.Context, documentID, content string) ([]domain.Citation, error) {
	chunks := chunkDocument(documentID, content)
	if len(chunks) == 0 {
		return nil, nil
	}

	indexed := make([]memoryChunk, 0, len(chunks))
	for _, chunk := range chunks {
		embedding, err := r.provider.EmbedDocument(ctx, chunk.text)
		if err != nil {
			return nil, err
		}

		indexed = append(indexed, memoryChunk{
			documentID: chunk.documentID,
			chunkID:    chunk.chunkID,
			text:       chunk.text,
			embedding:  cloneVector(embedding),
			dimension:  len(embedding),
		})
	}

	r.mu.Lock()
	updated := cloneDocuments(r.documents)
	updated[documentID] = cloneChunks(indexed)
	if r.store != nil {
		if err := r.store.Save(ctx, updated); err != nil {
			r.mu.Unlock()
			return nil, fmt.Errorf("%w: %v", ErrPersistenceUnavailable, err)
		}
	}
	r.documents = updated
	r.mu.Unlock()

	citations := make([]domain.Citation, 0, len(indexed))
	for _, chunk := range indexed {
		citations = append(citations, citationFromChunk(chunk))
	}

	return citations, nil
}

func cloneDocuments(documents map[string][]memoryChunk) map[string][]memoryChunk {
	if len(documents) == 0 {
		return make(map[string][]memoryChunk)
	}

	cloned := make(map[string][]memoryChunk, len(documents))
	for documentID, chunks := range documents {
		cloned[documentID] = cloneChunks(chunks)
	}

	return cloned
}

func cloneChunks(chunks []memoryChunk) []memoryChunk {
	if len(chunks) == 0 {
		return nil
	}

	cloned := make([]memoryChunk, len(chunks))
	for i, chunk := range chunks {
		cloned[i] = chunk
		cloned[i].embedding = cloneVector(chunk.embedding)
	}

	return cloned
}

func (r *MemoryRetriever) Retrieve(ctx context.Context, request domain.QueryRequest) ([]domain.Citation, error) {
	queryVector, err := r.provider.EmbedQuery(ctx, request.Query)
	if err != nil {
		return nil, err
	}
	if isZeroVector(queryVector) {
		return nil, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	type scoredCitation struct {
		citation domain.Citation
		score    float64
	}

	var scored []scoredCitation
	for _, chunks := range r.documents {
		for _, chunk := range chunks {
			if len(chunk.embedding) != len(queryVector) {
				return nil, fmt.Errorf("embedding dimension mismatch: persisted index uses %d dimensions but the current provider returned %d; re-ingest documents to rebuild the index", len(chunk.embedding), len(queryVector))
			}
			score := cosineSimilarity(queryVector, chunk.embedding)
			if score == 0 {
				continue
			}

			scored = append(scored, scoredCitation{
				citation: citationFromChunk(chunk),
				score:    score,
			})
		}
	}

	if len(scored) == 0 {
		return nil, nil
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if scored[i].citation.DocumentID != scored[j].citation.DocumentID {
			return scored[i].citation.DocumentID < scored[j].citation.DocumentID
		}
		return scored[i].citation.ChunkID < scored[j].citation.ChunkID
	})

	limit := 3
	if len(scored) < limit {
		limit = len(scored)
	}

	result := make([]domain.Citation, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, scored[i].citation)
	}

	return result, nil
}

func chunkDocument(documentID, content string) []memoryChunk {
	parts := splitChunks(content)
	if len(parts) == 0 {
		return nil
	}

	chunks := make([]memoryChunk, 0, len(parts))
	for index, part := range parts {
		chunks = append(chunks, memoryChunk{
			documentID: documentID,
			chunkID:    fmt.Sprintf("%s-chunk-%d", documentID, index+1),
			text:       part,
		})
	}

	return chunks
}

func splitChunks(content string) []string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil
	}

	rawParts := strings.Split(trimmed, "\n\n")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		if value := strings.TrimSpace(part); value != "" {
			parts = append(parts, value)
		}
	}

	if len(parts) == 0 {
		return []string{trimmed}
	}

	return parts
}

func cloneVector(vector []float64) []float64 {
	if len(vector) == 0 {
		return nil
	}

	clone := make([]float64, len(vector))
	copy(clone, vector)
	return clone
}

func citationFromChunk(chunk memoryChunk) domain.Citation {
	return domain.Citation{
		DocumentID: chunk.documentID,
		ChunkID:    chunk.chunkID,
		Snippet:    chunk.text,
	}
}

func isZeroVector(vector []float64) bool {
	for _, value := range vector {
		if value != 0 {
			return false
		}
	}

	return true
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
