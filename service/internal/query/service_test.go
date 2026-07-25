package query

import (
	"context"
	"errors"
	"strings"
	"testing"

	"rag-assistant/service/internal/domain"
)

type stubRetriever struct {
	citations []domain.Citation
	err       error
}

func (s stubRetriever) Retrieve(context.Context, domain.QueryRequest) ([]domain.Citation, error) {
	return s.citations, s.err
}

type stubComposer struct {
	answer string
	err    error
}

func (s stubComposer) Compose(context.Context, domain.QueryRequest, []domain.Citation) (string, error) {
	return s.answer, s.err
}

func TestQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		retriever Retriever
		composer  AnswerComposer
		request   domain.QueryRequest
		wantState string
		wantCode  string
		wantAns   string
		wantCites int
		wantChunk string
	}{
		{
			name: "insufficient context when no indexed chunks match",
			retriever: seededRetriever(t, scriptedEmbeddingProvider{vectors: map[string][]float64{
				"alpha beta":     {1, 0},
				"gamma delta":    {0, 1},
				"what is ready?": {0, 0},
			}}, "doc-1", "alpha beta\n\ngamma delta"),
			request:   domain.QueryRequest{Query: "what is ready?"},
			wantState: domain.QueryStateInsufficientContext,
			wantCode:  "insufficient_context",
		},
		{
			name: "answered when indexed chunks match",
			retriever: seededRetriever(t, scriptedEmbeddingProvider{vectors: map[string][]float64{
				"alpha beta":  {1, 0},
				"gamma delta": {0, 1},
				"delta":       {0, 1},
			}}, "doc-1", "alpha beta\n\ngamma delta"),
			request:   domain.QueryRequest{Query: "delta", ConversationID: "conv-1"},
			wantState: domain.QueryStateAnswered,
			wantCites: 1,
			wantChunk: "doc-1-chunk-2",
		},
		{
			name:      "retriever errors become unsupported responses",
			retriever: stubRetriever{err: errors.New("offline")},
			request:   domain.QueryRequest{Query: "what is ready?"},
			wantState: domain.QueryStateUnsupported,
			wantCode:  "retriever_unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := New(tt.retriever, tt.composer)
			response := service.Query(context.Background(), tt.request)

			if response.State != tt.wantState {
				t.Fatalf("expected state %q, got %q", tt.wantState, response.State)
			}

			if tt.wantCode != "" {
				if response.Error == nil || response.Error.Code != tt.wantCode {
					t.Fatalf("expected error code %q, got %#v", tt.wantCode, response.Error)
				}
			}

			if tt.wantState == domain.QueryStateAnswered && !strings.Contains(response.Answer, "gamma delta") {
				t.Fatalf("expected answer to use retrieved context, got %q", response.Answer)
			}

			if tt.wantState == domain.QueryStateAnswered && strings.Contains(response.Answer, "Based on the retrieved context") {
				t.Fatalf("expected extractive answer instead of placeholder text, got %q", response.Answer)
			}

			if response.Event != "" || response.Kind != "" {
				t.Fatalf("expected non-streaming response to omit stream-only fields, got %#v", response)
			}

			if got := len(response.Citations); got != tt.wantCites {
				t.Fatalf("expected %d citations, got %d", tt.wantCites, got)
			}

			if tt.wantChunk != "" {
				if len(response.Citations) == 0 || response.Citations[0].ChunkID != tt.wantChunk {
					t.Fatalf("expected chunk %q, got %#v", tt.wantChunk, response.Citations)
				}
			}
		})
	}
}

func TestQueryUsesEmbeddingSimilarity(t *testing.T) {
	t.Parallel()

	provider := scriptedEmbeddingProvider{vectors: map[string][]float64{
		"alpha chunk": {1, 0},
		"beta chunk":  {0, 1},
		"banana":      {1, 0},
	}}
	retriever := seededRetriever(t, provider, "doc-1", "alpha chunk\n\nbeta chunk")

	service := New(retriever, nil)
	response := service.Query(context.Background(), domain.QueryRequest{Query: "banana"})

	if response.State != domain.QueryStateAnswered {
		t.Fatalf("expected answered state, got %q", response.State)
	}

	if !strings.Contains(response.Answer, "alpha chunk") {
		t.Fatalf("expected answer from retrieved context, got %q", response.Answer)
	}

	if len(response.Citations) != 1 || response.Citations[0].ChunkID != "doc-1-chunk-1" {
		t.Fatalf("expected vector-selected citation, got %#v", response.Citations)
	}
}

func TestContextComposerProducesExtractiveAnswer(t *testing.T) {
	t.Parallel()

	answer, err := contextComposer{}.Compose(context.Background(), domain.QueryRequest{Query: "What explains readiness?"}, []domain.Citation{
		{DocumentID: "doc-1", ChunkID: "doc-1-chunk-1", Snippet: "General background. The readiness check depends on a live vector store."},
		{DocumentID: "doc-2", ChunkID: "doc-2-chunk-1", Snippet: "Streaming stays coherent when the final answer is composed from retrieved citations."},
	})
	if err != nil {
		t.Fatalf("compose failed: %v", err)
	}

	if strings.Contains(answer, "Based on the retrieved context") {
		t.Fatalf("expected extractive answer instead of placeholder text, got %q", answer)
	}

	if !strings.Contains(answer, "readiness check depends on a live vector store") {
		t.Fatalf("expected answer to use retrieved context, got %q", answer)
	}
}

func TestStream(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		service    *Service
		request    domain.QueryRequest
		wantStates []string
		wantEvents []string
		wantKinds  []string
		wantAnswer string
		wantCites  int
	}{
		{
			name: "answered query emits ordered lifecycle frames",
			service: New(seededRetriever(t, scriptedEmbeddingProvider{vectors: map[string][]float64{
				"alpha beta":  {1, 0},
				"gamma delta": {0, 1},
				"delta":       {0, 1},
			}}, "doc-1", "alpha beta\n\ngamma delta"), nil),
			request:    domain.QueryRequest{Query: "delta", ConversationID: "conv-1"},
			wantStates: []string{domain.QueryStateStreaming, domain.QueryStateRetrieving, domain.QueryStateStreaming, domain.QueryStateStreaming, domain.QueryStateAnswered},
			wantEvents: []string{domain.QueryStreamEventStart, domain.QueryStreamEventRetrieval, domain.QueryStreamEventContent, domain.QueryStreamEventCitation, domain.QueryStreamEventDone},
			wantKinds:  []string{domain.QueryStreamKindLifecycle, domain.QueryStreamKindRetrieval, domain.QueryStreamKindContent, domain.QueryStreamKindCitation, domain.QueryStreamKindCompletion},
			wantAnswer: "gamma delta",
			wantCites:  1,
		},
		{
			name:       "insufficient context still emits stable stream frames",
			service:    New(stubRetriever{}, nil),
			request:    domain.QueryRequest{Query: "delta", ConversationID: "conv-2"},
			wantStates: []string{domain.QueryStateStreaming, domain.QueryStateRetrieving, domain.QueryStateStreaming, domain.QueryStateStreaming, domain.QueryStateInsufficientContext},
			wantEvents: []string{domain.QueryStreamEventStart, domain.QueryStreamEventRetrieval, domain.QueryStreamEventContent, domain.QueryStreamEventCitation, domain.QueryStreamEventDone},
			wantKinds:  []string{domain.QueryStreamKindLifecycle, domain.QueryStreamKindRetrieval, domain.QueryStreamKindContent, domain.QueryStreamKindCitation, domain.QueryStreamKindCompletion},
			wantCites:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			frames := tt.service.Stream(context.Background(), tt.request)
			if len(frames) != len(tt.wantStates) {
				t.Fatalf("expected %d frames, got %d", len(tt.wantStates), len(frames))
			}

			for i, frame := range frames {
				if frame.State != tt.wantStates[i] {
					t.Fatalf("frame %d: expected state %q, got %q", i, tt.wantStates[i], frame.State)
				}
				if frame.Event != tt.wantEvents[i] {
					t.Fatalf("frame %d: expected event %q, got %q", i, tt.wantEvents[i], frame.Event)
				}
				if frame.Kind != tt.wantKinds[i] {
					t.Fatalf("frame %d: expected kind %q, got %q", i, tt.wantKinds[i], frame.Kind)
				}
				if frame.ConversationID != tt.request.ConversationID {
					t.Fatalf("frame %d: expected conversation id %q, got %q", i, tt.request.ConversationID, frame.ConversationID)
				}
			}

			if tt.wantAnswer != "" && !strings.Contains(frames[2].Answer, tt.wantAnswer) {
				t.Fatalf("expected content frame to include answer text %q, got %#v", tt.wantAnswer, frames[2])
			}
			if got := len(frames[1].Citations); got != tt.wantCites {
				t.Fatalf("expected retrieval frame to contain %d citations, got %d", tt.wantCites, got)
			}
			if got := len(frames[3].Citations); got != tt.wantCites {
				t.Fatalf("expected citation frame to contain %d citations, got %d", tt.wantCites, got)
			}
		})
	}
}

func TestQueryPersistsNewConversationAndFollowUp(t *testing.T) {
	store, err := NewConversationStore(t.TempDir() + "/conversations.json")
	if err != nil {
		t.Fatalf("create conversation store: %v", err)
	}
	retriever := stubRetriever{citations: []domain.Citation{{DocumentID: "doc-1", ChunkID: "chunk-1", Snippet: "retrieved context"}}}
	service := New(retriever, stubComposer{answer: "first answer"}, store)

	first := service.Query(context.Background(), domain.QueryRequest{Query: "first question"})
	if first.State != domain.QueryStateAnswered || first.ConversationID == "" {
		t.Fatalf("expected answered response with generated conversation id, got %#v", first)
	}

	second := New(retriever, stubComposer{answer: "follow-up answer"}, store).Query(context.Background(), domain.QueryRequest{
		Query: "follow-up question", ConversationID: first.ConversationID,
	})
	if second.ConversationID != first.ConversationID {
		t.Fatalf("expected conversation id to be reused, got %q and %q", first.ConversationID, second.ConversationID)
	}

	conversation, ok := store.Get(first.ConversationID)
	if !ok || len(conversation.Turns) != 2 {
		t.Fatalf("expected two persisted turns, got %#v", conversation)
	}
	if conversation.Turns[0].Query != "first question" || conversation.Turns[1].Answer != "follow-up answer" {
		t.Fatalf("unexpected persisted turns: %#v", conversation.Turns)
	}
}

func TestQueryPersistsInsufficientContextAndReusesConversationID(t *testing.T) {
	store, err := NewConversationStore(t.TempDir() + "/conversations.json")
	if err != nil {
		t.Fatalf("create conversation store: %v", err)
	}

	first := New(stubRetriever{}, nil, store).Query(context.Background(), domain.QueryRequest{Query: "first question"})
	if first.State != domain.QueryStateInsufficientContext || first.ConversationID == "" {
		t.Fatalf("expected insufficient-context response with generated conversation id, got %#v", first)
	}

	second := New(stubRetriever{citations: []domain.Citation{{DocumentID: "doc-1", ChunkID: "chunk-1", Snippet: "retrieved context"}}}, stubComposer{answer: "follow-up answer"}, store).Query(context.Background(), domain.QueryRequest{
		Query: "follow-up question", ConversationID: first.ConversationID,
	})
	if second.State != domain.QueryStateAnswered {
		t.Fatalf("expected answered follow-up, got %#v", second)
	}
	if second.ConversationID != first.ConversationID {
		t.Fatalf("expected conversation id to be reused, got %q and %q", first.ConversationID, second.ConversationID)
	}

	conversation, ok := store.Get(first.ConversationID)
	if !ok || len(conversation.Turns) != 2 {
		t.Fatalf("expected two persisted turns, got %#v", conversation)
	}
	if conversation.Turns[0].State != domain.QueryStateInsufficientContext || conversation.Turns[0].Query != "first question" {
		t.Fatalf("unexpected persisted insufficient-context turn: %#v", conversation.Turns[0])
	}
	if len(conversation.Turns[0].Citations) != 0 {
		t.Fatalf("expected zero fabricated citations, got %#v", conversation.Turns[0].Citations)
	}
	if conversation.Turns[0].CreatedAt.IsZero() {
		t.Fatalf("expected persisted insufficient-context turn to have a timestamp, got %#v", conversation.Turns[0].CreatedAt)
	}
	if conversation.Turns[1].State != domain.QueryStateAnswered || conversation.Turns[1].Answer != "follow-up answer" {
		t.Fatalf("unexpected persisted follow-up turn: %#v", conversation.Turns[1])
	}
}

func seededRetriever(t *testing.T, provider EmbeddingProvider, documentID, content string) *MemoryRetriever {
	t.Helper()

	retriever, err := NewMemoryRetrieverWithProvider(provider)
	if err != nil {
		t.Fatalf("failed to create retriever: %v", err)
	}
	if _, err := retriever.IndexDocument(context.Background(), documentID, content); err != nil {
		t.Fatalf("failed to seed retriever: %v", err)
	}

	return retriever
}

type scriptedEmbeddingProvider struct {
	vectors map[string][]float64
	err     error
}

func (p scriptedEmbeddingProvider) EmbedDocument(ctx context.Context, text string) ([]float64, error) {
	return p.embed(ctx, text)
}

func (p scriptedEmbeddingProvider) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	return p.embed(ctx, text)
}

func (p scriptedEmbeddingProvider) embed(_ context.Context, text string) ([]float64, error) {
	if p.err != nil {
		return nil, p.err
	}

	if vector, ok := p.vectors[text]; ok {
		return append([]float64(nil), vector...), nil
	}

	return []float64{0, 0}, nil
}
