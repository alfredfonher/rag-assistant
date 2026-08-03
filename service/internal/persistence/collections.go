package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"rag-assistant/service/internal/domain"
)

// FileCollectionRepo stores collections as atomic JSON.
type FileCollectionRepo struct {
	path string
	mu   sync.RWMutex
}

func NewFileCollectionRepo(dir string) (*FileCollectionRepo, error) {
	r := &FileCollectionRepo{path: filepath.Join(dir, "collections.json")}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrPersistence, err)
	}
	if _, err := r.read(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *FileCollectionRepo) Get(id string) (*domain.Collection, error) {
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

func (r *FileCollectionRepo) List() ([]domain.Collection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.read()
}

func (r *FileCollectionRepo) ListByAgent(agentID string) ([]domain.Collection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, err := r.read()
	if err != nil {
		return nil, err
	}
	out := make([]domain.Collection, 0)
	for _, c := range data {
		if c.AgentID == agentID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *FileCollectionRepo) Create(c *domain.Collection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := r.read()
	if err != nil {
		return err
	}
	for i := range data {
		if data[i].ID == c.ID {
			return domain.ErrDuplicate
		}
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	data = append(data, *c)
	return r.write(data)
}

func (r *FileCollectionRepo) Update(c *domain.Collection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := r.read()
	if err != nil {
		return err
	}
	for i := range data {
		if data[i].ID == c.ID {
			c.CreatedAt = data[i].CreatedAt
			c.UpdatedAt = time.Now().UTC()
			data[i] = *c
			return r.write(data)
		}
	}
	return domain.ErrNotFound
}

func (r *FileCollectionRepo) Delete(id string) error {
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

func (r *FileCollectionRepo) read() ([]domain.Collection, error) {
	return readJSONArray[domain.Collection](r.path)
}

func (r *FileCollectionRepo) write(data []domain.Collection) error {
	return writeJSONArray(r.path, data)
}
