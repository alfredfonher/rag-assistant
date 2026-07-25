package persistence

import (
	"encoding/json"
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
	_ = os.MkdirAll(dir, 0o755)
	if err := r.load(); err != nil && !os.IsNotExist(err) {
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
	return nil, os.ErrNotExist
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
	var out []domain.Collection
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
	data, _ := r.read()
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	data = append(data, *c)
	return r.write(data)
}

func (r *FileCollectionRepo) Update(c *domain.Collection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, _ := r.read()
	for i := range data {
		if data[i].ID == c.ID {
			c.CreatedAt = data[i].CreatedAt
			c.UpdatedAt = time.Now().UTC()
			data[i] = *c
			return r.write(data)
		}
	}
	return os.ErrNotExist
}

func (r *FileCollectionRepo) Delete(id string) error {
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

func (r *FileCollectionRepo) read() ([]domain.Collection, error) {
	raw, err := os.ReadFile(r.path)
	if err != nil {
		return nil, err
	}
	var out []domain.Collection
	return out, json.Unmarshal(raw, &out)
}

func (r *FileCollectionRepo) write(data []domain.Collection) error {
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

func (r *FileCollectionRepo) load() error {
	_, err := r.read()
	return err
}
