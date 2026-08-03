package registry_test

import (
	"testing"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/persistence"
	"rag-assistant/service/internal/registry"
)

func repos(t *testing.T) (domain.AgentRepo, domain.CollectionRepo, domain.DocumentRepo) {
	t.Helper()
	dir := t.TempDir()
	agents, err := persistence.NewFileAgentRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	collections, err := persistence.NewFileCollectionRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	documents, err := persistence.NewFileDocumentRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	service := registry.New(agents, collections, documents)
	return service.Agents(), service.Collections(), service.Documents()
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
