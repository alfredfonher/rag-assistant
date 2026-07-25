package readiness

import (
	"context"

	"rag-assistant/service/internal/domain"
	"rag-assistant/service/internal/llama"
)

// LlamaReadiness checks that the llama-server dependency is reachable and its
// receipt-backed verification fields confirm model integrity.
type LlamaReadiness struct {
	Client *llama.Client
}

func (r LlamaReadiness) Check(ctx context.Context) domain.ReadinessResponse {
	if r.Client == nil {
		return domain.ReadinessResponse{
			Ready: false,
			Reasons: []domain.DependencyReason{{
				Dependency: "llama-server",
				Code:       domain.ReadinessCodeUnavailable,
				Detail:     "client not configured",
			}},
		}
	}

	health, err := r.Client.Health(ctx)
	if err != nil {
		return domain.ReadinessResponse{
			Ready: false,
			Reasons: []domain.DependencyReason{{
				Dependency: "llama-server",
				Code:       domain.ReadinessCodeUnavailable,
				Detail:     err.Error(),
			}},
		}
	}

	reasons := []domain.DependencyReason{}

	if !health.EmbeddingLoaded {
		reasons = append(reasons, domain.DependencyReason{
			Dependency: "llama-server",
			Code:       "embedding_not_loaded",
			Detail:     "embedding model not loaded",
		})
	}

	if !health.LLMLoaded {
		reasons = append(reasons, domain.DependencyReason{
			Dependency: "llama-server",
			Code:       "llm_not_loaded",
			Detail:     "LLM model not loaded",
		})
	}

	ready := health.Status == "ok" && health.EmbeddingLoaded && health.LLMLoaded

	return domain.ReadinessResponse{
		Ready:   ready,
		Reasons: reasons,
	}
}
