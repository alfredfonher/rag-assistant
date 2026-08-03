package registry_test

import (
	"errors"
	"testing"

	"rag-assistant/service/internal/domain"
)

func TestCollectionsRequireAgentOnCreateAndUpdate(t *testing.T) {
	agents, collections, _ := repos(t)
	if err := collections.Create(&domain.Collection{ID: "collection", AgentID: "missing"}); !errors.Is(err, domain.ErrAgentNotFound) {
		t.Fatalf("create with missing agent: %v", err)
	}
	must(t, agents.Create(&domain.Agent{ID: "agent"}))
	must(t, collections.Create(&domain.Collection{ID: "collection", AgentID: "agent"}))
	if err := collections.Update(&domain.Collection{ID: "collection", AgentID: "missing"}); !errors.Is(err, domain.ErrAgentNotFound) {
		t.Fatalf("update with missing agent: %v", err)
	}
}

func TestCollectionsGuardDelete(t *testing.T) {
	agents, collections, documents := repos(t)
	must(t, agents.Create(&domain.Agent{ID: "agent"}))
	must(t, collections.Create(&domain.Collection{ID: "collection", AgentID: "agent"}))
	must(t, documents.Create(&domain.Document{ID: "document", CollectionID: "collection"}))
	if err := collections.Delete("collection"); !errors.Is(err, domain.ErrCollectionInUse) {
		t.Fatalf("delete referenced collection: %v", err)
	}
}
