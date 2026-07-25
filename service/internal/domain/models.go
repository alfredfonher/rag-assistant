package domain

import "time"

// Document lifecycle states.
const (
	DocStatusPending   = "pending"
	DocStatusIndexing  = "indexing"
	DocStatusReady     = "ready"
	DocStatusError     = "error"
	DocStatusOutdated  = "outdated"
)

// Agent represents a query configuration (model, prompt, parameters).
type Agent struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Model        string    `json:"model"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
	Temperature  float64   `json:"temperature"`
	MaxTokens    int       `json:"max_tokens"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Collection groups documents under a common agent.
type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	AgentID     string    `json:"agent_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Document represents an ingested file with lifecycle metadata.
type Document struct {
	ID           string    `json:"id"`
	CollectionID string    `json:"collection_id"`
	Filename     string    `json:"filename"`
	Path         string    `json:"path"`
	Status       string    `json:"status"`
	ChunksCount  int       `json:"chunks_count"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
