package api

import (
	"errors"
	"net/http"

	"rag-assistant/service/internal/domain"
)

func writeRegistryError(w http.ResponseWriter, err error, resource string) {
	switch {
	case errors.Is(err, domain.ErrAgentNotFound):
		writeError(w, http.StatusUnprocessableEntity, "agent_not_found", "agent does not exist")
	case errors.Is(err, domain.ErrCollectionNotFound):
		writeError(w, http.StatusUnprocessableEntity, "collection_not_found", "collection does not exist")
	case errors.Is(err, domain.ErrAgentInUse):
		writeError(w, http.StatusConflict, "agent_in_use", "agent is referenced by a collection")
	case errors.Is(err, domain.ErrCollectionInUse):
		writeError(w, http.StatusConflict, "collection_in_use", "collection is referenced by a document")
	case errors.Is(err, domain.ErrDuplicate):
		writeError(w, http.StatusConflict, "duplicate_id", resource+" id already exists")
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", resource+" not found")
	default:
		writeError(w, http.StatusServiceUnavailable, "persistence_unavailable", "registry persistence is unavailable")
	}
}
