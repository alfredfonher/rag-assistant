package query

import (
	"context"
	"testing"
	"time"

	"rag-assistant/service/internal/domain"
)

func TestConversationStoreAppendReloadsTurnsAndTimestamps(t *testing.T) {
	path := t.TempDir() + "/conversations.json"
	createdAt := time.Date(2026, time.July, 20, 13, 0, 0, 0, time.UTC)
	store, err := NewConversationStore(path)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	turn := ConversationTurn{
		Query:     "what is ready?",
		State:     domain.QueryStateAnswered,
		Answer:    "The service is ready.",
		Citations: []domain.Citation{{DocumentID: "doc-1", ChunkID: "chunk-1", Snippet: "ready"}},
		CreatedAt: createdAt,
	}
	if err := store.Append(context.Background(), "conv-1", turn); err != nil {
		t.Fatalf("append turn: %v", err)
	}

	reloaded, err := NewConversationStore(path)
	if err != nil {
		t.Fatalf("reload store: %v", err)
	}
	conversation, ok := reloaded.Get("conv-1")
	if !ok || len(conversation.Turns) != 1 {
		t.Fatalf("expected one reloaded turn, got %#v", conversation)
	}
	got := conversation.Turns[0]
	if got.Query != turn.Query || got.State != turn.State || got.Answer != turn.Answer || got.Citations[0] != turn.Citations[0] {
		t.Fatalf("reloaded turn changed: got %#v, want %#v", got, turn)
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Fatalf("timestamp changed: got %s, want %s", got.CreatedAt, createdAt)
	}
}

func TestNewConversationIDIsStableOnceAssigned(t *testing.T) {
	first := NewConversationID()
	second := NewConversationID()
	if first == "" || second == "" || first == second {
		t.Fatalf("expected distinct non-empty conversation ids, got %q and %q", first, second)
	}
}
