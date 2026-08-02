package persistence

import (
	"encoding/json"
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
	_ = os.MkdirAll(dir, 0o755)
	if err := r.load(); err != nil && !os.IsNotExist(err) {
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
	return nil, os.ErrNotExist
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
	data, _ := r.read()
	now := time.Now().UTC()
	d.CreatedAt = now
	d.UpdatedAt = now
	data = append(data, *d)
	return r.write(data)
}

func (r *FileDocumentRepo) Update(d *domain.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, _ := r.read()
	for i := range data {
		if data[i].ID == d.ID {
			d.CreatedAt = data[i].CreatedAt
			d.UpdatedAt = time.Now().UTC()
			data[i] = *d
			return r.write(data)
		}
	}
	return os.ErrNotExist
}

func (r *FileDocumentRepo) UpdateStatus(id string, status string, chunksCount int, errMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, _ := r.read()
	for i := range data {
		if data[i].ID == id {
			data[i].Status = status
			data[i].ChunksCount = chunksCount
			data[i].ErrorMessage = errMsg
			data[i].UpdatedAt = time.Now().UTC()
			return r.write(data)
		}
	}
	return os.ErrNotExist
}

func (r *FileDocumentRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, _ := r.read()
	for i := range data {
		if data[i].ID == id {
			data = append(data[:i], data[i+1:]...)
			return r.write(data)
		}
	}
	return os.ErrNotExist
}

func (r *FileDocumentRepo) read() ([]domain.Document, error) {
	raw, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.Document{}, nil
		}
		return nil, err
	}
	var out []domain.Document
	return out, json.Unmarshal(raw, &out)
}

func (r *FileDocumentRepo) write(data []domain.Document) error {
	tmp := r.path + ".tmp"
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, r.path); err != nil {
		return err
	}
	return nil
}

func (r *FileDocumentRepo) load() error {
	_, err := r.read()
	return err
}
