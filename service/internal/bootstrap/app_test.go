package bootstrap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/llama"
)

func TestNewStartsWithMissingIngestRoot(t *testing.T) {
	llamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(llama.HealthResponse{
			Status:          "ok",
			EmbeddingLoaded: true,
			LLMLoaded:       true,
		})
	}))
	defer llamaServer.Close()

	dataDir := filepath.Join(t.TempDir(), "state")
	missingRoot := filepath.Join(t.TempDir(), "missing")
	t.Setenv("RAG_DATA_DIR", dataDir)
	t.Setenv("RAG_INGEST_ROOT", missingRoot)
	t.Setenv("RAG_LLAMA_SERVER_URL", llamaServer.URL)

	app, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if app == nil || app.Server == nil {
		t.Fatalf("expected initialized app, got %#v", app)
	}

	assertReadiness(t, app, http.StatusServiceUnavailable, false, "ingest-root")
	if err := os.Mkdir(missingRoot, 0o755); err != nil {
		t.Fatalf("create ingest root: %v", err)
	}
	assertReadiness(t, app, http.StatusOK, true, "")
}

func assertReadiness(t *testing.T, app *App, wantStatus int, wantReady bool, wantDependency string) {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	app.Server.Handler().ServeHTTP(recorder, request)
	if recorder.Code != wantStatus {
		t.Fatalf("expected readiness status %d, got %d", wantStatus, recorder.Code)
	}

	var response domain.ReadinessResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode readiness response: %v", err)
	}
	if response.Ready != wantReady {
		t.Fatalf("expected ready=%v, got %#v", wantReady, response)
	}
	if wantDependency != "" && (len(response.Reasons) != 1 || response.Reasons[0].Dependency != wantDependency) {
		t.Fatalf("expected dependency %q, got %#v", wantDependency, response.Reasons)
	}
}
