package registry_test

import (
	"errors"
	"testing"

	"rag-assistant/service/internal/domain"
)

func TestDocumentsRequireCollection(t *testing.T) {
	agents, collections, documents := repos(t)
	if err := documents.Create(&domain.Document{ID: "document", CollectionID: "missing"}); !errors.Is(err, domain.ErrCollectionNotFound) {
		t.Fatalf("create with missing collection: %v", err)
	}
	must(t, agents.Create(&domain.Agent{ID: "agent"}))
	must(t, collections.Create(&domain.Collection{ID: "collection", AgentID: "agent"}))
	must(t, collections.Create(&domain.Collection{ID: "target", AgentID: "agent"}))
	must(t, documents.Create(&domain.Document{ID: "document", CollectionID: "collection", Filename: "original"}))
	must(t, documents.Update(&domain.Document{ID: "document", CollectionID: "target", Filename: "updated"}))
	updated, err := documents.Get("document")
	must(t, err)
	if updated.CollectionID != "target" || updated.Filename != "updated" {
		t.Fatalf("valid update: %#v", updated)
	}
	if err := documents.Update(&domain.Document{ID: "document", CollectionID: "missing", Filename: "rejected"}); !errors.Is(err, domain.ErrCollectionNotFound) {
		t.Fatalf("update with missing collection: %v", err)
	}
	unchanged, err := documents.Get("document")
	must(t, err)
	if unchanged.CollectionID != "target" || unchanged.Filename != "updated" {
		t.Fatalf("rejected update changed document: %#v", unchanged)
	}
}
