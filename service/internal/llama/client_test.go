package llama

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestEmbedderUsesDocumentAndQueryContracts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/embeddings":
			var request struct {
				Input []string `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode document request: %v", err)
				return
			}
			if len(request.Input) != 1 || request.Input[0] != "document text" {
				t.Errorf("unexpected document payload: %#v", request)
			}
			_, _ = w.Write([]byte(`{"data":[{"embedding":[1,2]}]}`))
		case "/v1/embeddings/query":
			var request EmbedRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode query request: %v", err)
				return
			}
			if request.Text != "query text" {
				t.Errorf("unexpected query payload: %#v", request)
			}
			_, _ = w.Write([]byte(`{"embedding":[3,4],"model":"nomic-embed-text"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	embedder := NewEmbedder(NewClient(server.URL + "/"))
	document, err := embedder.EmbedDocument(context.Background(), "document text")
	if err != nil || len(document) != 2 || document[0] != 1 {
		t.Fatalf("unexpected document embedding %v, err %v", document, err)
	}
	query, err := embedder.EmbedQuery(context.Background(), "query text")
	if err != nil || len(query) != 2 || query[0] != 3 {
		t.Fatalf("unexpected query embedding %v, err %v", query, err)
	}
}

func TestClientRejectsNon2xxAndPreservesResponseBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "redirect", status: http.StatusTemporaryRedirect, body: "embedding endpoint moved"},
		{name: "server error", status: http.StatusServiceUnavailable, body: "model unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			_, err := NewClient(server.URL).EmbedQuery(context.Background(), "query")
			if err == nil || !strings.Contains(err.Error(), strconv.Itoa(tt.status)) || !strings.Contains(err.Error(), tt.body) {
				t.Fatalf("expected status and response body, got %v", err)
			}
		})
	}
}

func TestClientRejectsEmptyResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/embeddings/query" {
			_, _ = w.Write([]byte(`{"embedding":[]}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if _, err := client.EmbedQuery(context.Background(), "query"); err == nil || !strings.Contains(err.Error(), "empty query embedding") {
		t.Fatalf("expected empty embedding error, got %v", err)
	}
	if _, err := client.ChatCompletion(context.Background(), ChatRequest{}); err == nil || !strings.Contains(err.Error(), "no choices") {
		t.Fatalf("expected empty choices error, got %v", err)
	}
}

func TestClientPropagatesCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(100 * time.Millisecond):
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := NewClient(server.URL).EmbedQuery(ctx, "query")
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
