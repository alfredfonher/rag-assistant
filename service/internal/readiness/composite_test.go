package readiness

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"rag-assistant/service/internal/domain"
)

func TestIngestRootReadinessRecoversWhenRootAppears(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	checker := IngestRoot{Path: root}

	missing := checker.Check(context.Background())
	if missing.Ready {
		t.Fatal("expected missing ingest root to be not ready")
	}
	if len(missing.Reasons) != 1 || missing.Reasons[0] != (domain.DependencyReason{
		Dependency: "ingest-root",
		Code:       domain.ReadinessCodeUnavailable,
		Detail:     "configured ingest root is unavailable",
	}) {
		t.Fatalf("unexpected missing-root reasons: %#v", missing.Reasons)
	}

	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("create ingest root: %v", err)
	}
	recovered := checker.Check(context.Background())
	if !recovered.Ready || len(recovered.Reasons) != 0 {
		t.Fatalf("expected recovered ingest root to be ready, got %#v", recovered)
	}
}

func TestCompositePreservesDependencyReasons(t *testing.T) {
	llamaReason := domain.DependencyReason{
		Dependency: "llama-server",
		Code:       domain.ReadinessCodeUnavailable,
		Detail:     "llama unavailable",
	}
	checker := Composite{Checkers: []Checker{
		staticChecker{response: domain.ReadinessResponse{Ready: false, Reasons: []domain.DependencyReason{llamaReason}}},
		staticChecker{response: domain.ReadinessResponse{Ready: true}},
	}}

	response := checker.Check(context.Background())
	if response.Ready {
		t.Fatal("expected composite to preserve not-ready llama result")
	}
	if len(response.Reasons) != 1 || response.Reasons[0] != llamaReason {
		t.Fatalf("unexpected composite reasons: %#v", response.Reasons)
	}
}

type staticChecker struct {
	response domain.ReadinessResponse
}

func (s staticChecker) Check(context.Context) domain.ReadinessResponse {
	return s.response
}
