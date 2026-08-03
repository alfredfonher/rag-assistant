package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"rag-assistant/service/internal/domain"
)

// FileAgentRepo stores agents as atomic JSON.
type FileAgentRepo struct {
	path string
	mu   sync.RWMutex
}

func NewFileAgentRepo(dir string) (*FileAgentRepo, error) {
	r := &FileAgentRepo{path: filepath.Join(dir, "agents.json")}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrPersistence, err)
	}
	if _, err := r.read(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *FileAgentRepo) Get(id string) (*domain.Agent, error) {
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

func (r *FileAgentRepo) List() ([]domain.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.read()
}

func (r *FileAgentRepo) Create(a *domain.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := r.read()
	if err != nil {
		return err
	}
	for i := range data {
		if data[i].ID == a.ID {
			return domain.ErrDuplicate
		}
	}
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	data = append(data, *a)
	return r.write(data)
}

func (r *FileAgentRepo) Update(a *domain.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := r.read()
	if err != nil {
		return err
	}
	for i := range data {
		if data[i].ID == a.ID {
			a.CreatedAt = data[i].CreatedAt
			a.UpdatedAt = time.Now().UTC()
			data[i] = *a
			return r.write(data)
		}
	}
	return domain.ErrNotFound
}

func (r *FileAgentRepo) Delete(id string) error {
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

func (r *FileAgentRepo) read() ([]domain.Agent, error) {
	return readJSONArray[domain.Agent](r.path)
}

func (r *FileAgentRepo) write(data []domain.Agent) error {
	return writeJSONArray(r.path, data)
}
