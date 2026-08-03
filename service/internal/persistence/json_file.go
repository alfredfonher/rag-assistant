package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"rag-assistant/service/internal/domain"
)

func readJSONArray[T any](path string) ([]T, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []T{}, nil
	}
	if err != nil {
		return nil, persistenceError(err)
	}

	var values []T
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, persistenceError(err)
	}
	if values == nil {
		values = []T{}
	}
	return values, nil
}

func writeJSONArray[T any](path string, values []T) error {
	if values == nil {
		values = []T{}
	}
	raw, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return persistenceError(err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return persistenceError(err)
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()
	if err := tmp.Chmod(0o644); err != nil {
		return persistenceError(err)
	}
	if _, err := tmp.Write(raw); err != nil {
		return persistenceError(err)
	}
	if err := tmp.Close(); err != nil {
		return persistenceError(err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return persistenceError(err)
	}
	return nil
}

func persistenceError(err error) error {
	return fmt.Errorf("%w: %w", domain.ErrPersistence, err)
}
