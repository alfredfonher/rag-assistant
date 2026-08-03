package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"rag-assistant/service/internal/domain"
)

func TestRegistryErrorContract(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{domain.ErrNotFound, http.StatusNotFound, "not_found"},
		{domain.ErrDuplicate, http.StatusConflict, "duplicate_id"},
		{domain.ErrAgentNotFound, http.StatusUnprocessableEntity, "agent_not_found"},
		{domain.ErrCollectionNotFound, http.StatusUnprocessableEntity, "collection_not_found"},
		{domain.ErrAgentInUse, http.StatusConflict, "agent_in_use"},
		{domain.ErrCollectionInUse, http.StatusConflict, "collection_in_use"},
		{errors.New("secret path"), http.StatusServiceUnavailable, "persistence_unavailable"},
	}
	for _, tt := range tests {
		recorder := httptest.NewRecorder()
		writeRegistryError(recorder, tt.err, "resource")
		var response domain.APIError
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}
		if recorder.Code != tt.status || response.Code != tt.code || response.Message == "secret path" {
			t.Fatalf("got %d %#v", recorder.Code, response)
		}
	}
}
