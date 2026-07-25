package persistence_test

import (
	"os"
	"path/filepath"
	"testing"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/persistence"
)

func TestAgentRepoCRUD(t *testing.T) {
	dir := t.TempDir()
	repo, err := persistence.NewFileAgentRepo(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Create
	a := &domain.Agent{ID: "a1", Name: "Test Agent", Model: "gpt-4", Temperature: 0.7, MaxTokens: 1024}
	if err := repo.Create(a); err != nil {
		t.Fatalf("create: %v", err)
	}
	if a.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}

	// Get
	got, err := repo.Get("a1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Test Agent" {
		t.Fatalf("got name %q, want %q", got.Name, "Test Agent")
	}

	// List
	list, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d agents, want 1", len(list))
	}

	// Update
	got.Name = "Updated Agent"
	if err := repo.Update(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	got2, _ := repo.Get("a1")
	if got2.Name != "Updated Agent" {
		t.Fatalf("got name %q after update, want %q", got2.Name, "Updated Agent")
	}

	// Delete
	if err := repo.Delete("a1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = repo.Get("a1")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestAgentRepoPersistenceAcrossReload(t *testing.T) {
	dir := t.TempDir()
	repo1, _ := persistence.NewFileAgentRepo(dir)
	repo1.Create(&domain.Agent{ID: "a1", Name: "Persistent Agent", Model: "gpt-4", Temperature: 0.5, MaxTokens: 512})

	// Reload
	repo2, err := persistence.NewFileAgentRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo2.Get("a1")
	if err != nil {
		t.Fatalf("reload get: %v", err)
	}
	if got.Name != "Persistent Agent" {
		t.Fatalf("got name %q, want %q", got.Name, "Persistent Agent")
	}
}

func TestCollectionRepoCRUD(t *testing.T) {
	dir := t.TempDir()
	repo, err := persistence.NewFileCollectionRepo(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Create
	c := &domain.Collection{ID: "c1", Name: "Test Collection", AgentID: "a1"}
	if err := repo.Create(c); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Get
	got, err := repo.Get("c1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Test Collection" {
		t.Fatalf("got name %q, want %q", got.Name, "Test Collection")
	}

	// ListByAgent
	repo.Create(&domain.Collection{ID: "c2", Name: "Other", AgentID: "a2"})
	list, err := repo.ListByAgent("a1")
	if err != nil {
		t.Fatalf("list by agent: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d collections for agent a1, want 1", len(list))
	}

	// Delete
	if err := repo.Delete("c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestDocumentRepoCRUD(t *testing.T) {
	dir := t.TempDir()
	repo, err := persistence.NewFileDocumentRepo(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Create
	d := &domain.Document{ID: "d1", CollectionID: "c1", Filename: "test.md", Path: "/tmp/test.md", Status: domain.DocStatusPending}
	if err := repo.Create(d); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Get
	got, err := repo.Get("d1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Filename != "test.md" {
		t.Fatalf("got filename %q, want %q", got.Filename, "test.md")
	}

	// UpdateStatus
	if err := repo.UpdateStatus("d1", domain.DocStatusReady, 5, ""); err != nil {
		t.Fatalf("update status: %v", err)
	}
	got2, _ := repo.Get("d1")
	if got2.Status != domain.DocStatusReady {
		t.Fatalf("got status %q, want %q", got2.Status, domain.DocStatusReady)
	}
	if got2.ChunksCount != 5 {
		t.Fatalf("got chunks %d, want 5", got2.ChunksCount)
	}

	// ListByCollection
	repo.Create(&domain.Document{ID: "d2", CollectionID: "c2", Filename: "other.md", Path: "/tmp/other.md", Status: domain.DocStatusPending})
	list, err := repo.ListByCollection("c1")
	if err != nil {
		t.Fatalf("list by collection: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d docs for c1, want 1", len(list))
	}

	// Delete
	if err := repo.Delete("d1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestDocumentRepoPersistenceAcrossReload(t *testing.T) {
	dir := t.TempDir()
	repo1, _ := persistence.NewFileDocumentRepo(dir)
	repo1.Create(&domain.Document{ID: "d1", CollectionID: "c1", Filename: "doc.md", Path: "/tmp/doc.md", Status: domain.DocStatusReady, ChunksCount: 3})

	repo2, err := persistence.NewFileDocumentRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo2.Get("d1")
	if err != nil {
		t.Fatalf("reload get: %v", err)
	}
	if got.ChunksCount != 3 {
		t.Fatalf("got chunks %d, want 3", got.ChunksCount)
	}
}

func TestRepoHandlesMissingFile(t *testing.T) {
	dir := t.TempDir()
	repo, err := persistence.NewFileAgentRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}
}

func TestRepoHandlesCorruptFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "agents.json"), []byte("not json"), 0o644)
	_, err := persistence.NewFileAgentRepo(dir)
	if err == nil {
		t.Fatal("expected error for corrupt file")
	}
}
