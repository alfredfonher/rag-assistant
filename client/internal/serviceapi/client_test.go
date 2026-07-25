package serviceapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"alive":true,"service":"rag-assistant"}`))
		case "/readyz":
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"ready":false,"reasons":[{"dependency":"vector_store","code":"unavailable","detail":"offline"}]}`))
		case "/v1/query":
			w.WriteHeader(http.StatusNotImplemented)
			_, _ = w.Write([]byte(`{"state":"unsupported","error":{"code":"query_not_implemented","message":"query handling is not part of the bootstrap slice"}}`))
		case "/v1/query/stream":
			if got := r.Header.Get("Accept"); got != "text/event-stream" {
				t.Fatalf("unexpected accept header: %q", got)
			}

			var request QueryRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}

			if request.Query != "what is ready?" {
				t.Fatalf("unexpected stream request: %#v", request)
			}

			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("data: {\"state\":\"streaming\",\"conversation_id\":\"conv-1\"}\n\n"))
			_, _ = w.Write([]byte("data: {\"state\":\"answered\",\"answer\":\"ready\",\"conversation_id\":\"conv-1\",\"citations\":[{\"document_id\":\"doc-1\",\"chunk_id\":\"chunk-7\"}]}\n\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := New(server.URL)

	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if !health.Alive || health.Service != "rag-assistant" {
		t.Fatalf("unexpected health response: %#v", health)
	}

	ready, err := client.Ready(context.Background())
	if err == nil {
		t.Fatal("expected readiness error")
	}
	if statusErr, ok := err.(*HTTPStatusError); !ok || statusErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected readiness error: %#v", err)
	}
	if ready.Ready {
		t.Fatalf("expected readiness failure")
	}
	if len(ready.Reasons) != 1 || ready.Reasons[0].Code != "unavailable" {
		t.Fatalf("unexpected readiness response: %#v", ready)
	}

	query, err := client.Query(context.Background(), QueryRequest{Query: "what is ready?"})
	if err == nil {
		t.Fatal("expected query error")
	}
	if statusErr, ok := err.(*HTTPStatusError); !ok || statusErr.StatusCode != http.StatusNotImplemented {
		t.Fatalf("unexpected query error: %#v", err)
	}
	if query.State != "unsupported" || query.Error == nil || query.Error.Code != "query_not_implemented" {
		t.Fatalf("unexpected query response: %#v", query)
	}

	stream, err := client.StreamQuery(context.Background(), QueryRequest{Query: "what is ready?"})
	if err != nil {
		t.Fatalf("StreamQuery() error: %v", err)
	}
	defer stream.Close()

	first, err := stream.Next()
	if err != nil {
		t.Fatalf("Next() first event error: %v", err)
	}
	if first.State != "streaming" || first.ConversationID != "conv-1" {
		t.Fatalf("unexpected first stream event: %#v", first)
	}

	second, err := stream.Next()
	if err != nil {
		t.Fatalf("Next() second event error: %v", err)
	}
	if second.State != "answered" || second.Answer != "ready" || len(second.Citations) != 1 {
		t.Fatalf("unexpected second stream event: %#v", second)
	}

	if _, err := stream.Next(); err != io.EOF {
		t.Fatalf("expected EOF after stream end, got %v", err)
	}
}
