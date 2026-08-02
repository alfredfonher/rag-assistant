package readiness

import (
	"context"
	"os"
	"path/filepath"

	"rag-assistant/service/internal/domain"
)

type Checker interface {
	Check(context.Context) domain.ReadinessResponse
}

type Composite struct {
	Checkers []Checker
}

func (c Composite) Check(ctx context.Context) domain.ReadinessResponse {
	response := domain.ReadinessResponse{Ready: true}
	for _, checker := range c.Checkers {
		result := checker.Check(ctx)
		response.Ready = response.Ready && result.Ready
		response.Reasons = append(response.Reasons, result.Reasons...)
	}
	return response
}

type IngestRoot struct {
	Path string
}

func (r IngestRoot) Check(context.Context) domain.ReadinessResponse {
	absoluteRoot, err := filepath.Abs(r.Path)
	if err == nil {
		absoluteRoot, err = filepath.EvalSymlinks(absoluteRoot)
	}
	if err == nil {
		var info os.FileInfo
		info, err = os.Stat(absoluteRoot)
		if err == nil && !info.IsDir() {
			err = os.ErrInvalid
		}
	}
	if err != nil {
		return domain.ReadinessResponse{
			Ready: false,
			Reasons: []domain.DependencyReason{{
				Dependency: "ingest-root",
				Code:       domain.ReadinessCodeUnavailable,
				Detail:     "configured ingest root is unavailable",
			}},
		}
	}

	return domain.ReadinessResponse{Ready: true}
}
