package persistence

import (
	"encoding/json"
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
	_ = os.MkdirAll(dir, 0o755)
	if err := r.load(); err != nil && !os.IsNotExist(err) {
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
	return nil, os.ErrNotExist
}

func (r *FileAgentRepo) List() ([]domain.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.read()
}

func (r *FileAgentRepo) Create(a *domain.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, _ := r.read()
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	data = append(data, *a)
	return r.write(data)
}

func (r *FileAgentRepo) Update(a *domain.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, _ := r.read()
	for i := range data {
		if data[i].ID == a.ID {
			a.CreatedAt = data[i].CreatedAt
			a.UpdatedAt = time.Now().UTC()
			data[i] = *a
			return r.write(data)
		}
	}
	return os.ErrNotExist
}

func (r *FileAgentRepo) Delete(id string) error {
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

func (r *FileAgentRepo) read() ([]domain.Agent, error) {
	raw, err := os.ReadFile(r.path)
	if err != nil {
		return nil, err
	}
	var out []domain.Agent
	return out, json.Unmarshal(raw, &out)
}

func (r *FileAgentRepo) write(data []domain.Agent) error {
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

func (r *FileAgentRepo) load() error {
	_, err := r.read()
	return err
}
