package llama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"rag-assistant/service/internal/domain"
)

func TestComposerBuildsGroundedChatRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		var request ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode chat request: %v", err)
			return
		}
		if request.Model != "gemma3" || len(request.Messages) != 2 {
			t.Errorf("unexpected chat payload: %#v", request)
		}
		if got := request.Messages[1].Content; !strings.Contains(got, "Context:\nalpha\n\nbeta") || !strings.Contains(got, "Question: What?") {
			t.Errorf("unexpected user prompt %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"  grounded answer  "}}]}`))
	}))
	defer server.Close()

	composer := NewComposer(NewClient(server.URL))
	answer, err := composer.Compose(context.Background(), domain.QueryRequest{Query: "What?"}, []domain.Citation{
		{Snippet: "alpha"}, {Snippet: "alpha"}, {Snippet: "beta"},
	})
	if err != nil || answer != "grounded answer" {
		t.Fatalf("unexpected answer %q, err %v", answer, err)
	}
}

func TestComposerRejectsEmptyContext(t *testing.T) {
	t.Parallel()

	_, err := NewComposer(nil).Compose(context.Background(), domain.QueryRequest{Query: "What?"}, nil)
	if err == nil || !strings.Contains(err.Error(), "context is empty") {
		t.Fatalf("expected empty context error, got %v", err)
	}
}
