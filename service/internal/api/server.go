package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/query"
)

type ReadinessChecker interface {
	Check(context.Context) domain.ReadinessResponse
}

type QueryService interface {
	Query(context.Context, domain.QueryRequest) domain.QueryResponse
	Stream(context.Context, domain.QueryRequest) []domain.QueryResponse
}

type IngestService interface {
	Ingest(context.Context, domain.IngestRequest) domain.IngestResponse
}

type StaticReadiness struct {
	Response domain.ReadinessResponse
}

func (s StaticReadiness) Check(context.Context) domain.ReadinessResponse {
	return s.Response
}

type Server struct {
	serviceName   string
	readiness     ReadinessChecker
	ingestService IngestService
	queryService  QueryService
	mux           *http.ServeMux
}

type ServerDeps struct {
	Readiness     ReadinessChecker
	Ingest        IngestService
	Query         QueryService
	Agents        domain.AgentRepo
	Collections   domain.CollectionRepo
	Documents     domain.DocumentRepo
	Conversations *query.FileConversationStore
}

func NewServer(serviceName string, readiness ReadinessChecker, ingest IngestService, query QueryService) *Server {
	return NewServerWithDeps(serviceName, ServerDeps{
		Readiness: readiness,
		Ingest:    ingest,
		Query:     query,
	})
}

func NewServerWithDeps(serviceName string, deps ServerDeps) *Server {
	if deps.Readiness == nil {
		deps.Readiness = StaticReadiness{Response: domain.ReadinessResponse{Ready: true}}
	}
	if deps.Ingest == nil {
		deps.Ingest = StaticIngestService{}
	}
	if deps.Query == nil {
		deps.Query = StaticQueryService{}
	}

	s := &Server{
		serviceName:   serviceName,
		readiness:     deps.Readiness,
		ingestService: deps.Ingest,
		queryService:  deps.Query,
		mux:           http.NewServeMux(),
	}

	// Core endpoints
	s.mux.HandleFunc("/healthz", s.healthz)
	s.mux.HandleFunc("/readyz", s.readyz)
	s.mux.HandleFunc("/v1/documents/ingest", s.ingest)
	s.mux.HandleFunc("/v1/query", s.query)
	s.mux.HandleFunc("/v1/query/stream", s.queryStream)

	// CRUD endpoints
	if deps.Agents != nil {
		s.mux.Handle("/v1/agents/", NewAgentHandler(deps.Agents))
		s.mux.Handle("/v1/agents", NewAgentHandler(deps.Agents))
	}
	if deps.Collections != nil {
		s.mux.Handle("/v1/collections/", NewCollectionHandler(deps.Collections))
		s.mux.Handle("/v1/collections", NewCollectionHandler(deps.Collections))
	}
	if deps.Documents != nil {
		s.mux.Handle("/v1/documents/", NewDocumentHandler(deps.Documents))
		s.mux.Handle("/v1/documents", NewDocumentHandler(deps.Documents))
	}
	if deps.Conversations != nil {
		s.mux.Handle("/v1/conversations/", NewConversationHandler(deps.Conversations))
		s.mux.Handle("/v1/conversations", NewConversationHandler(deps.Conversations))
	}

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	writeJSON(w, http.StatusOK, domain.HealthResponse{Alive: true, Service: s.serviceName})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	response := s.readiness.Check(r.Context())
	status := http.StatusOK
	if !response.Ready {
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, response)
}

func (s *Server) query(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var request domain.QueryRequest
	if err := decodeJSON(r.Body, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, domain.QueryResponse{
			State: domain.QueryStateUnsupported,
			Error: &domain.APIError{Code: "invalid_request", Message: "query body must be valid JSON"},
		})
		return
	}

	if strings.TrimSpace(request.Query) == "" {
		writeJSON(w, http.StatusBadRequest, domain.QueryResponse{
			State: domain.QueryStateUnsupported,
			Error: &domain.APIError{Code: "invalid_request", Message: "query is required"},
		})
		return
	}

	response := s.queryService.Query(r.Context(), request)
	writeJSON(w, queryStatus(response), response)
}

func (s *Server) ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var request domain.IngestRequest
	if err := decodeJSON(r.Body, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, domain.IngestResponse{
			State: domain.DocumentStateUnsupported,
			Error: &domain.APIError{Code: "invalid_request", Message: "ingest body must be valid JSON"},
		})
		return
	}

	response := s.ingestService.Ingest(r.Context(), request)
	writeJSON(w, ingestStatus(response), publicIngestResponse(response))
}

func publicIngestResponse(response domain.IngestResponse) domain.IngestResponse {
	if len(response.Citations) == 0 {
		return response
	}

	citations := make([]domain.Citation, len(response.Citations))
	for i, citation := range response.Citations {
		citations[i] = domain.Citation{
			DocumentID: citation.DocumentID,
			ChunkID:    citation.ChunkID,
		}
	}
	response.Citations = citations
	return response
}

func (s *Server) queryStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var request domain.QueryRequest
	if err := decodeJSON(r.Body, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, domain.QueryResponse{
			State: domain.QueryStateUnsupported,
			Error: &domain.APIError{Code: "invalid_request", Message: "query body must be valid JSON"},
		})
		return
	}

	if strings.TrimSpace(request.Query) == "" {
		writeJSON(w, http.StatusBadRequest, domain.QueryResponse{
			State: domain.QueryStateUnsupported,
			Error: &domain.APIError{Code: "invalid_request", Message: "query is required"},
		})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	frames := s.queryService.Stream(r.Context(), request)
	if len(frames) == 0 {
		frames = []domain.QueryResponse{{
			State: domain.QueryStateUnsupported,
			Error: &domain.APIError{Code: "query_stream_empty", Message: "query stream returned no events"},
		}}
	}

	for i, frame := range frames {
		if err := writeSSE(w, i+1, eventName(frame), frame); err != nil {
			return
		}
	}
}

func decodeJSON(body io.ReadCloser, dst any) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSSE(w http.ResponseWriter, id int, event string, payload any) error {
	if _, err := fmt.Fprintf(w, "id: %d\n", id); err != nil {
		return err
	}

	if event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
			return err
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

func eventName(frame domain.QueryResponse) string {
	if frame.Event != "" {
		return frame.Event
	}
	if frame.Kind != "" {
		return frame.Kind
	}

	switch frame.State {
	case domain.QueryStateStreaming:
		return "start"
	case domain.QueryStateRetrieving:
		return "retrieval"
	case domain.QueryStateAnswered, domain.QueryStateInsufficientContext:
		return "done"
	case domain.QueryStateUnsupported:
		return "error"
	default:
		return "message"
	}
}

func queryStatus(response domain.QueryResponse) int {
	if response.State != domain.QueryStateUnsupported {
		return http.StatusOK
	}

	if response.Error != nil && response.Error.Code == "invalid_request" {
		return http.StatusBadRequest
	}

	return http.StatusServiceUnavailable
}

func ingestStatus(response domain.IngestResponse) int {
	switch response.State {
	case domain.DocumentStateIndexed:
		return http.StatusCreated
	case domain.DocumentStateUnindexed:
		if response.Error != nil {
			switch response.Error.Code {
			case "embedding_unavailable":
				return http.StatusServiceUnavailable
			}
		}
		return http.StatusServiceUnavailable
	case domain.DocumentStateUnsupported:
		if response.Error != nil {
			switch response.Error.Code {
			case "ingest_root_unavailable":
				return http.StatusServiceUnavailable
			case "document_not_found":
				return http.StatusNotFound
			case "unsupported_document":
				return http.StatusUnsupportedMediaType
			case "document_too_large":
				return http.StatusRequestEntityTooLarge
			case "document_unreadable", "document_not_regular", "empty_document":
				return http.StatusUnprocessableEntity
			}
		}
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

func methodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeJSON(w, http.StatusMethodNotAllowed, domain.APIError{Code: "method_not_allowed", Message: "unexpected HTTP method"})
}

type StaticQueryService struct {
	Response domain.QueryResponse
	Frames   []domain.QueryResponse
}

type StaticIngestService struct {
	Response domain.IngestResponse
}

func (s StaticIngestService) Ingest(context.Context, domain.IngestRequest) domain.IngestResponse {
	if s.Response.State != "" {
		return s.Response
	}

	return domain.IngestResponse{
		State: domain.DocumentStateUnsupported,
		Error: &domain.APIError{Code: "ingest_not_implemented", Message: "ingest handling is not configured"},
	}
}

func (s StaticQueryService) Query(context.Context, domain.QueryRequest) domain.QueryResponse {
	if s.Response.State != "" {
		return s.Response
	}

	return domain.QueryResponse{
		State: domain.QueryStateUnsupported,
		Error: &domain.APIError{Code: "query_not_implemented", Message: "query handling is not configured"},
	}
}

func (s StaticQueryService) Stream(ctx context.Context, request domain.QueryRequest) []domain.QueryResponse {
	if len(s.Frames) != 0 {
		return s.Frames
	}

	return []domain.QueryResponse{s.Query(ctx, request)}
}
