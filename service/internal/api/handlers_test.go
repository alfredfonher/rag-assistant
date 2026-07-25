package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rag-assistant/service/internal/api"
	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/persistence"
	"rag-assistant/service/internal/query"
)

func setupAgentHandler(t *testing.T) (*api.AgentHandler, *persistence.FileAgentRepo) {
	t.Helper()
	dir := t.TempDir()
	repo, err := persistence.NewFileAgentRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	return api.NewAgentHandler(repo), repo
}

func TestAgentCRUD(t *testing.T) {
	handler, _ := setupAgentHandler(t)

	// Create
	body, _ := json.Marshal(domain.Agent{Name: "Test Agent", Model: "gpt-4", Temperature: 0.7, MaxTokens: 1024})
	req := httptest.NewRequest(http.MethodPost, "/v1/agents", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: got %d, want %d", w.Code, http.StatusCreated)
	}
	var created domain.Agent
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.ID == "" {
		t.Fatal("expected ID to be set")
	}

	// List
	req = httptest.NewRequest(http.MethodGet, "/v1/agents", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list: got %d, want %d", w.Code, http.StatusOK)
	}
	var list []domain.Agent
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("list: got %d, want 1", len(list))
	}

	// Get
	req = httptest.NewRequest(http.MethodGet, "/v1/agents/"+created.ID, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get: got %d, want %d", w.Code, http.StatusOK)
	}

	// Update
	body, _ = json.Marshal(domain.Agent{Name: "Updated Agent"})
	req = httptest.NewRequest(http.MethodPut, "/v1/agents/"+created.ID, bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update: got %d, want %d", w.Code, http.StatusOK)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/v1/agents/"+created.ID, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d, want %d", w.Code, http.StatusNoContent)
	}

	// Get after delete
	req = httptest.NewRequest(http.MethodGet, "/v1/agents/"+created.ID, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAgentCreateValidation(t *testing.T) {
	handler, _ := setupAgentHandler(t)

	// Missing name
	body, _ := json.Marshal(domain.Agent{Model: "gpt-4"})
	req := httptest.NewRequest(http.MethodPost, "/v1/agents", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing name: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCollectionCRUD(t *testing.T) {
	dir := t.TempDir()
	repo, _ := persistence.NewFileCollectionRepo(dir)
	handler := api.NewCollectionHandler(repo)

	// Create
	body, _ := json.Marshal(domain.Collection{Name: "Test Collection", AgentID: "a1"})
	req := httptest.NewRequest(http.MethodPost, "/v1/collections", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: got %d, want %d", w.Code, http.StatusCreated)
	}
	var created domain.Collection
	json.Unmarshal(w.Body.Bytes(), &created)

	// List
	req = httptest.NewRequest(http.MethodGet, "/v1/collections", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list: got %d, want %d", w.Code, http.StatusOK)
	}

	// List by agent
	req = httptest.NewRequest(http.MethodGet, "/v1/collections?agent_id=a1", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	var list []domain.Collection
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("list by agent: got %d, want 1", len(list))
	}

	// Get
	req = httptest.NewRequest(http.MethodGet, "/v1/collections/"+created.ID, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get: got %d, want %d", w.Code, http.StatusOK)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/v1/collections/"+created.ID, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestDocumentCRUD(t *testing.T) {
	dir := t.TempDir()
	repo, _ := persistence.NewFileDocumentRepo(dir)
	handler := api.NewDocumentHandler(repo)

	// Create
	body, _ := json.Marshal(domain.Document{Filename: "test.md", CollectionID: "c1"})
	req := httptest.NewRequest(http.MethodPost, "/v1/documents", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: got %d, want %d", w.Code, http.StatusCreated)
	}
	var created domain.Document
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.Status != domain.DocStatusPending {
		t.Fatalf("status: got %q, want %q", created.Status, domain.DocStatusPending)
	}

	// List
	req = httptest.NewRequest(http.MethodGet, "/v1/documents", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list: got %d, want %d", w.Code, http.StatusOK)
	}

	// Get
	req = httptest.NewRequest(http.MethodGet, "/v1/documents/"+created.ID, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get: got %d, want %d", w.Code, http.StatusOK)
	}

	// Reindex
	req = httptest.NewRequest(http.MethodPost, "/v1/documents/"+created.ID+"/reindex", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("reindex: got %d, want %d", w.Code, http.StatusOK)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/v1/documents/"+created.ID, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestConversationHandler(t *testing.T) {
	dir := t.TempDir()
	store, _ := query.NewConversationStore(dir + "/conversations.json")
	handler := api.NewConversationHandler(store)

	// Create a conversation via store (normally done by query service)
	store.Append(context.Background(), "conv-1", query.ConversationTurn{Query: "hello", State: "answered", Answer: "hi"})

	// List
	req := httptest.NewRequest(http.MethodGet, "/v1/conversations", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list: got %d, want %d", w.Code, http.StatusOK)
	}

	// Get
	req = httptest.NewRequest(http.MethodGet, "/v1/conversations/conv-1", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get: got %d, want %d", w.Code, http.StatusOK)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/v1/conversations/conv-1", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d, want %d", w.Code, http.StatusNoContent)
	}
}
