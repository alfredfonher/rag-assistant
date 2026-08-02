package bootstrap

import (
	"path/filepath"
	"testing"
)

func TestNewStartsWithMissingIngestRoot(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "state")
	missingRoot := filepath.Join(t.TempDir(), "missing")
	t.Setenv("RAG_DATA_DIR", dataDir)
	t.Setenv("RAG_INGEST_ROOT", missingRoot)

	app, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if app == nil || app.Server == nil {
		t.Fatalf("expected initialized app, got %#v", app)
	}
}
