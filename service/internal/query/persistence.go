package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrPersistenceUnavailable = errors.New("index persistence unavailable")

type retrieverStore interface {
	Load(context.Context) (map[string][]memoryChunk, error)
	Save(context.Context, map[string][]memoryChunk) error
}

type fileRetrieverStore struct {
	path string
}

func newFileRetrieverStore(path string) retrieverStore {
	return fileRetrieverStore{path: path}
}

func (s fileRetrieverStore) Load(context.Context) (map[string][]memoryChunk, error) {
	if s.path == "" {
		return map[string][]memoryChunk{}, nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string][]memoryChunk{}, nil
		}
		return nil, err
	}

	state := persistedRetrieverState{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return state.toDocuments(), nil
}

func (s fileRetrieverStore) Save(ctx context.Context, documents map[string][]memoryChunk) error {
	if s.path == "" {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	state := persistedRetrieverState{Version: 1, Documents: make(map[string][]persistedChunk, len(documents))}
	for documentID, chunks := range documents {
		state.Documents[documentID] = make([]persistedChunk, 0, len(chunks))
		for _, chunk := range chunks {
			state.Documents[documentID] = append(state.Documents[documentID], persistedChunk{
				DocumentID: chunk.documentID,
				ChunkID:    chunk.chunkID,
				Text:       chunk.text,
				Embedding:  cloneVector(chunk.embedding),
				Dimension:  chunk.dimension,
			})
		}
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("%w: %v", ErrPersistenceUnavailable, err)
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(s.path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPersistenceUnavailable, err)
	}
	defer os.Remove(tmp.Name())

	enc := json.NewEncoder(tmp)
	if err := enc.Encode(state); err != nil {
		tmp.Close()
		return fmt.Errorf("%w: %v", ErrPersistenceUnavailable, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrPersistenceUnavailable, err)
	}

	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("%w: %v", ErrPersistenceUnavailable, err)
	}

	return nil
}

type persistedRetrieverState struct {
	Version   int                       `json:"version"`
	Documents map[string][]persistedChunk `json:"documents"`
}

type persistedChunk struct {
	DocumentID string    `json:"document_id"`
	ChunkID    string    `json:"chunk_id"`
	Text       string    `json:"text"`
	Embedding  []float64 `json:"embedding"`
	Dimension  int       `json:"dimension"`
}

func (s persistedRetrieverState) toDocuments() map[string][]memoryChunk {
	documents := make(map[string][]memoryChunk, len(s.Documents))
	for documentID, chunks := range s.Documents {
		documents[documentID] = make([]memoryChunk, 0, len(chunks))
		for _, chunk := range chunks {
			documents[documentID] = append(documents[documentID], memoryChunk{
				documentID: chunk.DocumentID,
				chunkID:    chunk.ChunkID,
				text:       chunk.Text,
				embedding:  cloneVector(chunk.Embedding),
				dimension:  chunk.Dimension,
			})
		}
	}

	return documents
}
