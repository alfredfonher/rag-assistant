package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"rag-assistant/service/internal/domain"
)

// FileDocumentRepo stores documents as atomic JSON.
type FileDocumentRepo struct {
	path string
	mu   sync.RWMutex
}

func NewFileDocumentRepo(dir string) (*FileDocumentRepo, error) {
	r := &FileDocumentRepo{path: filepath.Join(dir, "documents.json")}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrPersistence, err)
	}
	if _, err := r.read(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *FileDocumentRepo) Get(id string) (*domain.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, err := r.read()
	if err != nil {
		return nil, err
	}
	for i := range data {
		if data[i].ID == id {
			return &data[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *FileDocumentRepo) List() ([]domain.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.read()
}

func (r *FileDocumentRepo) ListByCollection(collectionID string) ([]domain.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, err := r.read()
	if err != nil {
		return nil, err
	}
	out := make([]domain.Document, 0)
	for _, d := range data {
		if d.CollectionID == collectionID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *FileDocumentRepo) Create(d *domain.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := r.read()
	if err != nil {
		return err
	}
	for i := range data {
		if data[i].ID == d.ID {
			return domain.ErrDuplicate
		}
	}
	now := time.Now().UTC()
	d.CreatedAt = now
	d.UpdatedAt = now
	data = append(data, *d)
	return r.write(data)
}

func (r *FileDocumentRepo) Update(d *domain.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := r.read()
	if err != nil {
		return err
	}
	for i := range data {
		if data[i].ID == d.ID {
			d.CreatedAt = data[i].CreatedAt
			d.UpdatedAt = time.Now().UTC()
			data[i] = *d
			return r.write(data)
		}
	}
	return domain.ErrNotFound
}

func (r *FileDocumentRepo) UpdateStatus(id string, status string, chunksCount int, errMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := r.read()
	if err != nil {
		return err
	}
	for i := range data {
		if data[i].ID == id {
			data[i].Status = status
			data[i].ChunksCount = chunksCount
			data[i].ErrorMessage = errMsg
			data[i].UpdatedAt = time.Now().UTC()
			return r.write(data)
		}
	}
	return domain.ErrNotFound
}

func (r *FileDocumentRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := r.read()
	if err != nil {
		return err
	}
	for i := range data {
		if data[i].ID == id {
			data = append(data[:i], data[i+1:]...)
			return r.write(data)
		}
	}
	return domain.ErrNotFound
}

func (r *FileDocumentRepo) read() ([]domain.Document, error) {
	return readJSONArray[domain.Document](r.path)
}

func (r *FileDocumentRepo) write(data []domain.Document) error {
	return writeJSONArray(r.path, data)
}
