package persistence

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"rag-assistant/service/internal/domain"
)

func TestJSONArrayFilesystemErrorsPreserveIdentityAndCleanup(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := readJSONArray[any](target)
	var pathErr *os.PathError
	if !errors.Is(err, domain.ErrPersistence) || !errors.As(err, &pathErr) {
		t.Fatalf("read error lost identity: %v", err)
	}
	err = writeJSONArray(target, []string{"value"})
	var linkErr *os.LinkError
	if !errors.Is(err, domain.ErrPersistence) || !errors.As(err, &linkErr) {
		t.Fatalf("rename error lost identity: %v", err)
	}
	entries, err := os.ReadDir(filepath.Dir(target))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "target" {
		t.Fatalf("temporary file residue: %v", entries)
	}
}
