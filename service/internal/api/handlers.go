package api

import (
	"net/http"
	"strings"
	"time"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/query"
)

// extractID pulls the last path segment as an ID: /v1/agents/abc → "abc"
func extractID(path, prefix string) string {
	prefix = prefix + "/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	if id == "" || strings.Contains(id, "/") {
		return ""
	}
	return id
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, domain.APIError{Code: code, Message: msg})
}

// --- Agents ---

type AgentHandler struct {
	repo domain.AgentRepo
}

func NewAgentHandler(repo domain.AgentRepo) *AgentHandler {
	return &AgentHandler{repo: repo}
}

func (h *AgentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/v1/agents")

	switch {
	case r.Method == http.MethodGet && id == "":
		h.list(w, r)
	case r.Method == http.MethodPost && id == "":
		h.create(w, r)
	case r.Method == http.MethodGet && id != "":
		h.get(w, r, id)
	case r.Method == http.MethodPut && id != "":
		h.update(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		h.delete(w, r, id)
	default:
		methodNotAllowed(w, "GET, POST, PUT, DELETE")
	}
}

func (h *AgentHandler) list(w http.ResponseWriter, _ *http.Request) {
	list, err := h.repo.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *AgentHandler) get(w http.ResponseWriter, _ *http.Request, id string) {
	a, err := h.repo.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "agent not found")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *AgentHandler) create(w http.ResponseWriter, r *http.Request) {
	var a domain.Agent
	if err := decodeJSON(r.Body, &a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "body must be valid JSON")
		return
	}
	if strings.TrimSpace(a.Name) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}
	if a.ID == "" {
		a.ID = generateID()
	}
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now

	if err := h.repo.Create(&a); err != nil {
		writeError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (h *AgentHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	var a domain.Agent
	if err := decodeJSON(r.Body, &a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "body must be valid JSON")
		return
	}
	a.ID = id
	if err := h.repo.Update(&a); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "agent not found")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *AgentHandler) delete(w http.ResponseWriter, _ *http.Request, id string) {
	if err := h.repo.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "agent not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Collections ---

type CollectionHandler struct {
	repo domain.CollectionRepo
}

func NewCollectionHandler(repo domain.CollectionRepo) *CollectionHandler {
	return &CollectionHandler{repo: repo}
}

func (h *CollectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/v1/collections")

	switch {
	case r.Method == http.MethodGet && id == "":
		h.list(w, r)
	case r.Method == http.MethodPost && id == "":
		h.create(w, r)
	case r.Method == http.MethodGet && id != "":
		h.get(w, r, id)
	case r.Method == http.MethodPut && id != "":
		h.update(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		h.delete(w, r, id)
	default:
		methodNotAllowed(w, "GET, POST, PUT, DELETE")
	}
}

func (h *CollectionHandler) list(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	var list []domain.Collection
	var err error
	if agentID != "" {
		list, err = h.repo.ListByAgent(agentID)
	} else {
		list, err = h.repo.List()
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *CollectionHandler) get(w http.ResponseWriter, _ *http.Request, id string) {
	c, err := h.repo.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "collection not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CollectionHandler) create(w http.ResponseWriter, r *http.Request) {
	var c domain.Collection
	if err := decodeJSON(r.Body, &c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "body must be valid JSON")
		return
	}
	if strings.TrimSpace(c.Name) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}
	if strings.TrimSpace(c.AgentID) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "agent_id is required")
		return
	}
	if c.ID == "" {
		c.ID = generateID()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	if err := h.repo.Create(&c); err != nil {
		writeError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *CollectionHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	var c domain.Collection
	if err := decodeJSON(r.Body, &c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "body must be valid JSON")
		return
	}
	c.ID = id
	if err := h.repo.Update(&c); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "collection not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CollectionHandler) delete(w http.ResponseWriter, _ *http.Request, id string) {
	if err := h.repo.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "collection not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Documents ---

type DocumentHandler struct {
	repo domain.DocumentRepo
}

func NewDocumentHandler(repo domain.DocumentRepo) *DocumentHandler {
	return &DocumentHandler{repo: repo}
}

func (h *DocumentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// Handle /v1/documents/{id}/reindex
	if strings.HasPrefix(path, "/v1/documents/") {
		rest := strings.TrimPrefix(path, "/v1/documents/")
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) == 2 && parts[1] == "reindex" && r.Method == http.MethodPost {
			h.reindex(w, r, parts[0])
			return
		}
	}

	id := extractID(r.URL.Path, "/v1/documents")

	switch {
	case r.Method == http.MethodGet && id == "":
		h.list(w, r)
	case r.Method == http.MethodPost && id == "":
		h.create(w, r)
	case r.Method == http.MethodGet && id != "":
		h.get(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		h.delete(w, r, id)
	default:
		methodNotAllowed(w, "GET, POST, DELETE")
	}
}

func (h *DocumentHandler) list(w http.ResponseWriter, r *http.Request) {
	collID := r.URL.Query().Get("collection_id")
	var list []domain.Document
	var err error
	if collID != "" {
		list, err = h.repo.ListByCollection(collID)
	} else {
		list, err = h.repo.List()
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *DocumentHandler) get(w http.ResponseWriter, _ *http.Request, id string) {
	d, err := h.repo.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "document not found")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *DocumentHandler) create(w http.ResponseWriter, r *http.Request) {
	var d domain.Document
	if err := decodeJSON(r.Body, &d); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "body must be valid JSON")
		return
	}
	if strings.TrimSpace(d.Filename) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "filename is required")
		return
	}
	if strings.TrimSpace(d.CollectionID) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "collection_id is required")
		return
	}
	if d.ID == "" {
		d.ID = generateID()
	}
	d.Status = domain.DocStatusPending
	now := time.Now().UTC()
	d.CreatedAt = now
	d.UpdatedAt = now

	if err := h.repo.Create(&d); err != nil {
		writeError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

func (h *DocumentHandler) delete(w http.ResponseWriter, _ *http.Request, id string) {
	if err := h.repo.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "document not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DocumentHandler) reindex(w http.ResponseWriter, _ *http.Request, id string) {
	if err := h.repo.UpdateStatus(id, domain.DocStatusPending, 0, ""); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "document not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"state": "pending", "document_id": id})
}

// --- Conversations ---

// ConversationStore is the minimal interface the handler needs.
// Defined as a concrete dependency on the query package's store.
type ConversationStore = query.FileConversationStore

func NewConversationHandler(store *query.FileConversationStore) *ConversationHandler {
	return &ConversationHandler{store: store}
}

type ConversationHandler struct {
	store *query.FileConversationStore
}

func (h *ConversationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/v1/conversations")

	switch {
	case r.Method == http.MethodGet && id == "":
		h.list(w, r)
	case r.Method == http.MethodGet && id != "":
		h.get(w, r, id)
	case r.Method == http.MethodPatch && id != "":
		h.rename(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		h.delete(w, r, id)
	default:
		methodNotAllowed(w, "GET, PATCH, DELETE")
	}
}

func (h *ConversationHandler) list(w http.ResponseWriter, _ *http.Request) {
	convs := h.store.List()
	type convSummary struct {
		ID    string `json:"id"`
		Turns int    `json:"turns_count"`
	}
	out := make([]convSummary, len(convs))
	for i := range convs {
		out[i] = convSummary{ID: convs[i].ID, Turns: len(convs[i].Turns)}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ConversationHandler) get(w http.ResponseWriter, _ *http.Request, id string) {
	c, ok := h.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "conversation not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *ConversationHandler) rename(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r.Body, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "body must be valid JSON")
		return
	}
	if err := h.store.Rename(id, body.Name); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "conversation not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "name": body.Name})
}

func (h *ConversationHandler) delete(w http.ResponseWriter, _ *http.Request, id string) {
	if err := h.store.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "conversation not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// generateID creates a simple unique ID (timestamp + random suffix).
func generateID() string {
	return time.Now().UTC().Format("20060102150405") + "-" + randomHex(8)
}

func randomHex(n int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = hex[time.Now().UnixNano()%16]
		time.Sleep(1) // ensure different nanos
	}
	return string(b)
}
