package persistence_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/persistence"
)

func TestRepositoryEmptyAndNullLists(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"agents.json", "collections.json", "documents.json"} {
		must(t, os.WriteFile(filepath.Join(dir, name), []byte("null"), 0o644))
	}
	agents, err := persistence.NewFileAgentRepo(dir)
	must(t, err)
	collections, err := persistence.NewFileCollectionRepo(dir)
	must(t, err)
	documents, err := persistence.NewFileDocumentRepo(dir)
	must(t, err)

	values, err := agents.List()
	assertEmpty(t, values, err)
	collectionValues, err := collections.List()
	assertEmpty(t, collectionValues, err)
	collectionValues, err = collections.ListByAgent("missing")
	assertEmpty(t, collectionValues, err)
	documentValues, err := documents.List()
	assertEmpty(t, documentValues, err)
	documentValues, err = documents.ListByCollection("missing")
	assertEmpty(t, documentValues, err)

	must(t, os.Remove(filepath.Join(dir, "agents.json")))
	must(t, os.Remove(filepath.Join(dir, "collections.json")))
	values, err = agents.List()
	assertEmpty(t, values, err)
	collectionValues, err = collections.List()
	assertEmpty(t, collectionValues, err)
	collectionValues, err = collections.ListByAgent("missing")
	assertEmpty(t, collectionValues, err)
}

func TestRepositoryMutationSafety(t *testing.T) {
	dir := t.TempDir()
	agents, err := persistence.NewFileAgentRepo(dir)
	must(t, err)
	collections, err := persistence.NewFileCollectionRepo(dir)
	must(t, err)
	documents, err := persistence.NewFileDocumentRepo(dir)
	must(t, err)
	must(t, agents.Create(&domain.Agent{ID: "same"}))
	must(t, collections.Create(&domain.Collection{ID: "same"}))
	must(t, documents.Create(&domain.Document{ID: "same"}))

	for _, err := range []error{
		agents.Create(&domain.Agent{ID: "same"}),
		collections.Create(&domain.Collection{ID: "same"}),
		documents.Create(&domain.Document{ID: "same"}),
	} {
		if !errors.Is(err, domain.ErrDuplicate) {
			t.Fatalf("expected duplicate error, got %v", err)
		}
	}

	for name, mutate := range map[string]func() error{
		"agents.json":      func() error { return agents.Create(&domain.Agent{ID: "new"}) },
		"collections.json": func() error { return collections.Delete("same") },
		"documents.json":   func() error { return documents.UpdateStatus("same", domain.DocStatusReady, 1, "") },
	} {
		path := filepath.Join(dir, name)
		must(t, os.WriteFile(path, []byte("corrupt"), 0o644))
		if err := mutate(); !errors.Is(err, domain.ErrPersistence) {
			t.Fatalf("%s: expected persistence error, got %v", name, err)
		}
		raw, err := os.ReadFile(path)
		must(t, err)
		if string(raw) != "corrupt" {
			t.Fatalf("%s was overwritten", name)
		}
	}
}

func assertEmpty[T any](t *testing.T, values []T, err error) {
	t.Helper()
	if err != nil || values == nil || len(values) != 0 {
		t.Fatalf("expected non-nil empty list, got %#v, %v", values, err)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
