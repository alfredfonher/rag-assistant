package serviceapi

type HealthResponse struct {
	Alive   bool   `json:"alive"`
	Service string `json:"service"`
}

type DependencyReason struct {
	Dependency string `json:"dependency"`
	Code       string `json:"code"`
	Detail     string `json:"detail,omitempty"`
}

type ReadinessResponse struct {
	Ready   bool               `json:"ready"`
	Reasons []DependencyReason `json:"reasons,omitempty"`
}

type QueryRequest struct {
	Query          string `json:"query"`
	ConversationID string `json:"conversation_id,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Citation struct {
	DocumentID string `json:"document_id"`
	ChunkID    string `json:"chunk_id"`
}

type QueryResponse struct {
	State          string     `json:"state"`
	Answer         string     `json:"answer,omitempty"`
	Citations      []Citation `json:"citations,omitempty"`
	ConversationID string     `json:"conversation_id,omitempty"`
	Error          *APIError  `json:"error,omitempty"`
}

type IngestRequest struct {
	Path string `json:"path"`
}

type IngestResponse struct {
	State      string     `json:"state"`
	DocumentID string     `json:"document_id,omitempty"`
	Citations  []Citation `json:"citations,omitempty"`
	Error      *APIError  `json:"error,omitempty"`
}
