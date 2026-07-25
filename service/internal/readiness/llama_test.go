package readiness

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rag-assistant/service/internal/llama"
)

func TestLlamaReadiness_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(llama.HealthResponse{
			Status:          "ok",
			EmbeddingLoaded: true,
			LLMLoaded:       true,
		})
	}))
	defer srv.Close()

	checker := LlamaReadiness{Client: llama.NewClient(srv.URL)}
	resp := checker.Check(context.Background())

	if !resp.Ready {
		t.Fatalf("expected ready=true, got ready=false, reasons=%v", resp.Reasons)
	}
	if len(resp.Reasons) != 0 {
		t.Fatalf("expected 0 reasons, got %d", len(resp.Reasons))
	}
}

func TestLlamaReadiness_Unreachable(t *testing.T) {
	checker := LlamaReadiness{Client: llama.NewClient("http://127.0.0.1:1")}
	resp := checker.Check(context.Background())

	if resp.Ready {
		t.Fatal("expected ready=false for unreachable server")
	}
	if len(resp.Reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d", len(resp.Reasons))
	}
	if resp.Reasons[0].Code != "unavailable" {
		t.Fatalf("expected code 'unavailable', got %q", resp.Reasons[0].Code)
	}
}

func TestLlamaReadiness_ModelsNotLoaded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(llama.HealthResponse{
			Status:          "ok",
			EmbeddingLoaded: false,
			LLMLoaded:       false,
		})
	}))
	defer srv.Close()

	checker := LlamaReadiness{Client: llama.NewClient(srv.URL)}
	resp := checker.Check(context.Background())

	if resp.Ready {
		t.Fatal("expected ready=false when models not loaded")
	}
	if len(resp.Reasons) != 2 {
		t.Fatalf("expected 2 reasons, got %d", len(resp.Reasons))
	}
}

func TestLlamaReadiness_NilClient(t *testing.T) {
	checker := LlamaReadiness{Client: nil}
	resp := checker.Check(context.Background())

	if resp.Ready {
		t.Fatal("expected ready=false for nil client")
	}
}
