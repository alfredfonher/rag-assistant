package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"rag-assistant/service/internal/domain"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	NewServer("rag-assistant", nil, nil, nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response domain.HealthResponse
	decodeResponse(t, rec.Body.Bytes(), &response)

	if !response.Alive || response.Service != "rag-assistant" {
		t.Fatalf("unexpected health response: %#v", response)
	}
}

func TestReadyz(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		checker    ReadinessChecker
		wantStatus int
		wantReady  bool
	}{
		{
			name:       "ready",
			checker:    StaticReadiness{Response: domain.ReadinessResponse{Ready: true}},
			wantStatus: http.StatusOK,
			wantReady:  true,
		},
		{
			name: "not ready",
			checker: StaticReadiness{Response: domain.ReadinessResponse{
				Ready: false,
				Reasons: []domain.DependencyReason{{
					Dependency: "vector_store",
					Code:       domain.ReadinessCodeUnavailable,
					Detail:     "qdrant is offline",
				}},
			}},
			wantStatus: http.StatusServiceUnavailable,
			wantReady:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

			NewServer("rag-assistant", tt.checker, nil, nil).Handler().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var response domain.ReadinessResponse
			decodeResponse(t, rec.Body.Bytes(), &response)

			if response.Ready != tt.wantReady {
				t.Fatalf("expected ready=%v, got %v", tt.wantReady, response.Ready)
			}
		})
	}
}

func TestQueryContract(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/query", strings.NewReader(`{"query":"what is ready?"}`))

	NewServer("rag-assistant", nil, nil, stubQueryService{
		response: domain.QueryResponse{
			State:          domain.QueryStateInsufficientContext,
			ConversationID: "conv-1",
			Error:          &domain.APIError{Code: "insufficient_context", Message: "no relevant context found"},
		},
	}).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response domain.QueryResponse
	decodeResponse(t, rec.Body.Bytes(), &response)

	if response.State != domain.QueryStateInsufficientContext {
		t.Fatalf("expected insufficient context state, got %q", response.State)
	}

	if response.Error == nil || response.Error.Code != "insufficient_context" {
		t.Fatalf("unexpected query error: %#v", response.Error)
	}
	if response.ConversationID != "conv-1" {
		t.Fatalf("expected conversation id to be preserved, got %q", response.ConversationID)
	}
}

func TestQueryStream(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/query/stream", strings.NewReader(`{"query":"what is ready?","conversation_id":"conv-1"}`))

	NewServer("rag-assistant", nil, nil, stubQueryService{
		frames: []domain.QueryResponse{
			{State: domain.QueryStateStreaming, Event: domain.QueryStreamEventStart, Kind: domain.QueryStreamKindLifecycle, ConversationID: "conv-1"},
			{State: domain.QueryStateRetrieving, Event: domain.QueryStreamEventRetrieval, Kind: domain.QueryStreamKindRetrieval, ConversationID: "conv-1", Citations: []domain.Citation{{DocumentID: "doc-1", ChunkID: "chunk-7"}}},
			{State: domain.QueryStateStreaming, Event: domain.QueryStreamEventContent, Kind: domain.QueryStreamKindContent, ConversationID: "conv-1", Answer: "ready"},
			{State: domain.QueryStateStreaming, Event: domain.QueryStreamEventCitation, Kind: domain.QueryStreamKindCitation, ConversationID: "conv-1", Citations: []domain.Citation{{DocumentID: "doc-1", ChunkID: "chunk-7"}}},
			{State: domain.QueryStateAnswered, Event: domain.QueryStreamEventDone, Kind: domain.QueryStreamKindCompletion, ConversationID: "conv-1", Answer: "ready", Citations: []domain.Citation{{DocumentID: "doc-1", ChunkID: "chunk-7"}}},
		},
	}).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("unexpected content type: %q", got)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: start") || !strings.Contains(body, "event: retrieval") || !strings.Contains(body, "event: content") || !strings.Contains(body, "event: citation") || !strings.Contains(body, "event: done") {
		t.Fatalf("unexpected SSE body: %q", body)
	}
	if strings.Index(body, "event: start") > strings.Index(body, "event: retrieval") || strings.Index(body, "event: retrieval") > strings.Index(body, "event: content") || strings.Index(body, "event: content") > strings.Index(body, "event: citation") || strings.Index(body, "event: citation") > strings.Index(body, "event: done") {
		t.Fatalf("events out of order: %q", body)
	}
	if !strings.Contains(body, `"conversation_id":"conv-1"`) {
		t.Fatalf("expected conversation id in SSE payloads, got %q", body)
	}
}

func TestIngestContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		response   domain.IngestResponse
		wantStatus int
	}{
		{
			name:       "indexed",
			response:   domain.IngestResponse{State: domain.DocumentStateIndexed, DocumentID: "guide"},
			wantStatus: http.StatusCreated,
		},
		{
			name: "unindexed",
			response: domain.IngestResponse{State: domain.DocumentStateUnindexed, Error: &domain.APIError{
				Code:    "embedding_unavailable",
				Message: "embedding offline",
			}},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/v1/documents/ingest", strings.NewReader(`{"path":"/tmp/guide.md"}`))

			NewServer("rag-assistant", nil, stubIngestService{response: tt.response}, nil).Handler().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var response domain.IngestResponse
			decodeResponse(t, rec.Body.Bytes(), &response)
			if response.State != tt.response.State {
				t.Fatalf("unexpected ingest response: %#v", response)
			}
		})
	}
}

func TestQueryStreamMonotonicIDs(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/query/stream", strings.NewReader(`{"query":"count ids"}`))

	NewServer("rag-assistant", nil, nil, stubQueryService{
		frames: []domain.QueryResponse{
			{State: domain.QueryStateStreaming, Event: domain.QueryStreamEventStart, Kind: domain.QueryStreamKindLifecycle},
			{State: domain.QueryStateRetrieving, Event: domain.QueryStreamEventRetrieval, Kind: domain.QueryStreamKindRetrieval},
			{State: domain.QueryStateStreaming, Event: domain.QueryStreamEventContent, Kind: domain.QueryStreamKindContent, Answer: "hello"},
			{State: domain.QueryStateStreaming, Event: domain.QueryStreamEventCitation, Kind: domain.QueryStreamKindCitation},
			{State: domain.QueryStateAnswered, Event: domain.QueryStreamEventDone, Kind: domain.QueryStreamKindCompletion},
		},
	}).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	for i := 1; i <= 5; i++ {
		expected := fmt.Sprintf("id: %d\n", i)
		if !strings.Contains(body, expected) {
			t.Errorf("missing %q in SSE body", expected)
		}
	}

	lines := strings.Split(body, "\n")
	var ids []int
	for _, line := range lines {
		if strings.HasPrefix(line, "id: ") {
			var id int
			if _, err := fmt.Sscanf(line, "id: %d", &id); err == nil {
				ids = append(ids, id)
			}
		}
	}
	for i := 1; i < len(ids); i++ {
		if ids[i] != ids[i-1]+1 {
			t.Fatalf("ids not monotonically increasing: %v", ids)
		}
	}
}

type stubQueryService struct {
	response domain.QueryResponse
	frames   []domain.QueryResponse
}

type stubIngestService struct {
	response domain.IngestResponse
}

func (s stubQueryService) Query(context.Context, domain.QueryRequest) domain.QueryResponse {
	return s.response
}

func (s stubQueryService) Stream(context.Context, domain.QueryRequest) []domain.QueryResponse {
	return s.frames
}

func (s stubIngestService) Ingest(context.Context, domain.IngestRequest) domain.IngestResponse {
	return s.response
}

func decodeResponse(t *testing.T, body []byte, dst any) {
	t.Helper()

	if err := json.Unmarshal(body, dst); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}
