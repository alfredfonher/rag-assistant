package registry_test

import (
	"errors"
	"sync"
	"testing"

	"rag-assistant/service/internal/domain"
)

func TestAgentsPropagateDuplicateAndGuardDelete(t *testing.T) {
	agents, collections, _ := repos(t)
	must(t, agents.Create(&domain.Agent{ID: "agent"}))
	if err := agents.Create(&domain.Agent{ID: "agent"}); !errors.Is(err, domain.ErrDuplicate) {
		t.Fatalf("duplicate create: %v", err)
	}
	must(t, collections.Create(&domain.Collection{ID: "collection", AgentID: "agent"}))
	if err := agents.Delete("agent"); !errors.Is(err, domain.ErrAgentInUse) {
		t.Fatalf("delete referenced agent: %v", err)
	}
}

func TestAgentDeleteSerializesWithCollectionCreate(t *testing.T) {
	for i := 0; i < 20; i++ {
		agents, collections, _ := repos(t)
		must(t, agents.Create(&domain.Agent{ID: "agent"}))
		var createErr, deleteErr error
		var group sync.WaitGroup
		group.Add(2)
		go func() {
			defer group.Done()
			createErr = collections.Create(&domain.Collection{ID: "collection", AgentID: "agent"})
		}()
		go func() {
			defer group.Done()
			deleteErr = agents.Delete("agent")
		}()
		group.Wait()
		if createErr == nil && deleteErr == nil {
			t.Fatal("relationship check and mutation were not serialized")
		}
	}
}
