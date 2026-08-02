package query

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"rag-assistant/service/internal/domain"
)

type Retriever interface {
	Retrieve(context.Context, domain.QueryRequest) ([]domain.Citation, error)
}

type AnswerComposer interface {
	Compose(context.Context, domain.QueryRequest, []domain.Citation) (string, error)
}

type Service struct {
	retriever Retriever
	composer  AnswerComposer
	store     ConversationStore
}

type resolvedQuery struct {
	response  domain.QueryResponse
	citations []domain.Citation
}

func New(retriever Retriever, composer AnswerComposer, stores ...ConversationStore) *Service {
	if retriever == nil {
		retriever = NewMemoryRetriever()
	}

	if composer == nil {
		composer = contextComposer{}
	}

	var store ConversationStore
	if len(stores) > 0 {
		store = stores[0]
	}
	return &Service{retriever: retriever, composer: composer, store: store}
}

func (s *Service) Query(ctx context.Context, request domain.QueryRequest) domain.QueryResponse {
	resolved := s.resolve(ctx, request)
	response := resolved.response
	_ = s.persistTerminalTurn(ctx, request, response)
	return response
}

func (s *Service) Stream(ctx context.Context, request domain.QueryRequest) []domain.QueryResponse {
	resolved := s.resolve(ctx, request)
	response := resolved.response
	if err := s.persistTerminalTurn(ctx, request, response); err != nil {
		return []domain.QueryResponse{
			streamFrame(domain.QueryStreamEventStart, domain.QueryStreamKindLifecycle, domain.QueryStateStreaming, response.ConversationID, "", nil, nil),
			streamFrame(domain.QueryStreamEventDone, domain.QueryStreamKindCompletion, domain.QueryStateUnsupported, response.ConversationID, "", nil, &domain.APIError{
				Code: "conversation_persistence_unavailable", Message: "conversation persistence unavailable",
			}),
		}
	}
	if response.State == domain.QueryStateUnsupported {
		return []domain.QueryResponse{
			streamFrame(domain.QueryStreamEventStart, domain.QueryStreamKindLifecycle, domain.QueryStateStreaming, response.ConversationID, "", nil, nil),
			streamFrame(domain.QueryStreamEventDone, domain.QueryStreamKindCompletion, response.State, response.ConversationID, "", nil, response.Error),
		}
	}

	return []domain.QueryResponse{
		streamFrame(domain.QueryStreamEventStart, domain.QueryStreamKindLifecycle, domain.QueryStateStreaming, response.ConversationID, "", nil, nil),
		streamFrame(domain.QueryStreamEventRetrieval, domain.QueryStreamKindRetrieval, domain.QueryStateRetrieving, response.ConversationID, "", resolved.citations, nil),
		streamFrame(domain.QueryStreamEventContent, domain.QueryStreamKindContent, domain.QueryStateStreaming, response.ConversationID, response.Answer, nil, nil),
		streamFrame(domain.QueryStreamEventCitation, domain.QueryStreamKindCitation, domain.QueryStateStreaming, response.ConversationID, "", response.Citations, nil),
		streamFrame(domain.QueryStreamEventDone, domain.QueryStreamKindCompletion, response.State, response.ConversationID, response.Answer, response.Citations, response.Error),
	}
}

func (s *Service) persistTerminalTurn(ctx context.Context, request domain.QueryRequest, response domain.QueryResponse) error {
	if s.store == nil || (response.State != domain.QueryStateAnswered && response.State != domain.QueryStateInsufficientContext) {
		return nil
	}
	return s.store.Append(ctx, response.ConversationID, ConversationTurn{
		Query: request.Query, State: response.State, Answer: response.Answer,
		Citations: response.Citations, CreatedAt: time.Now().UTC(),
	})
}

func (s *Service) resolve(ctx context.Context, request domain.QueryRequest) resolvedQuery {
	if strings.TrimSpace(request.Query) == "" {
		return resolvedQuery{response: domain.QueryResponse{
			State:          domain.QueryStateUnsupported,
			ConversationID: request.ConversationID,
			Error:          &domain.APIError{Code: "invalid_request", Message: "query is required"},
		}}
	}
	if request.ConversationID == "" {
		request.ConversationID = NewConversationID()
	}

	citations, err := s.retriever.Retrieve(ctx, request)
	if err != nil {
		return resolvedQuery{response: domain.QueryResponse{
			State:          domain.QueryStateUnsupported,
			ConversationID: request.ConversationID,
			Error:          &domain.APIError{Code: "retriever_unavailable", Message: err.Error()},
		}}
	}

	if len(citations) == 0 {
		response := domain.QueryResponse{
			State:          domain.QueryStateInsufficientContext,
			ConversationID: request.ConversationID,
			Error:          &domain.APIError{Code: "insufficient_context", Message: "no relevant context found"},
		}
		return resolvedQuery{response: response}
	}

	answer, err := s.composer.Compose(ctx, request, citations)
	if err != nil {
		return resolvedQuery{response: domain.QueryResponse{
			State:          domain.QueryStateUnsupported,
			ConversationID: request.ConversationID,
			Error:          &domain.APIError{Code: "answer_unavailable", Message: err.Error()},
		}}
	}

	response := domain.QueryResponse{
		State:          domain.QueryStateAnswered,
		Answer:         answer,
		Citations:      citations,
		ConversationID: request.ConversationID,
	}
	return resolvedQuery{response: response, citations: citations}
}

func streamFrame(event, kind, state, conversationID, answer string, citations []domain.Citation, err *domain.APIError) domain.QueryResponse {
	return domain.QueryResponse{
		State:          state,
		Event:          event,
		Kind:           kind,
		Answer:         answer,
		Citations:      citations,
		ConversationID: conversationID,
		Error:          err,
	}
}

type contextComposer struct{}

func (contextComposer) Compose(_ context.Context, request domain.QueryRequest, citations []domain.Citation) (string, error) {
	snippets := uniqueSnippets(citations)
	if len(snippets) == 0 {
		return "", fmt.Errorf("retrieved context is empty")
	}

	answer := composeAnswer(request.Query, snippets)
	if strings.TrimSpace(answer) == "" {
		return "", fmt.Errorf("retrieved context is empty")
	}

	return answer, nil
}

type sentenceCandidate struct {
	text          string
	score         int
	citationIndex int
	sentenceIndex int
}

func composeAnswer(query string, snippets []string) string {
	terms := queryTerms(query)
	candidates := make([]sentenceCandidate, 0, len(snippets))

	for citationIndex, snippet := range snippets {
		sentences := splitSentences(snippet)
		if len(sentences) == 0 {
			sentences = []string{snippet}
		}

		for sentenceIndex, sentence := range sentences {
			normalized := normalizeSentence(sentence)
			if normalized == "" {
				continue
			}

			score := scoreSentence(normalized, terms)

			candidates = append(candidates, sentenceCandidate{
				text:          normalized,
				score:         score,
				citationIndex: citationIndex,
				sentenceIndex: sentenceIndex,
			})
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		if candidates[i].citationIndex != candidates[j].citationIndex {
			return candidates[i].citationIndex < candidates[j].citationIndex
		}
		if candidates[i].sentenceIndex != candidates[j].sentenceIndex {
			return candidates[i].sentenceIndex < candidates[j].sentenceIndex
		}
		return candidates[i].text < candidates[j].text
	})

	selected := uniqueSentences(candidates, 2)
	if len(selected) == 0 {
		return ""
	}

	answer := sentenceText(selected[0])
	if len(selected) > 1 {
		answer += " Additional context: " + sentenceText(selected[1])
	}

	return answer
}

func queryTerms(query string) []string {
	query = strings.ToLower(query)
	terms := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	filtered := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if len(term) < 4 {
			continue
		}
		filtered = append(filtered, term)
	}

	return filtered
}

func scoreSentence(sentence string, terms []string) int {
	if len(terms) == 0 {
		return 0
	}

	score := 0
	lower := strings.ToLower(sentence)
	for _, term := range terms {
		if strings.Contains(lower, term) {
			score++
		}
	}

	return score
}

func uniqueSentences(candidates []sentenceCandidate, limit int) []sentenceCandidate {
	seen := make(map[string]struct{}, len(candidates))
	selected := make([]sentenceCandidate, 0, limit)
	for _, candidate := range candidates {
		key := strings.ToLower(candidate.text)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		selected = append(selected, candidate)
		if len(selected) == limit {
			break
		}
	}

	return selected
}

func splitSentences(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?' || r == '\n' || r == '\r'
	})
	if len(parts) == 0 {
		return nil
	}

	sentences := make([]string, 0, len(parts))
	for _, part := range parts {
		if sentence := normalizeSentence(part); sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

func normalizeSentence(sentence string) string {
	sentence = strings.TrimSpace(sentence)
	sentence = strings.TrimLeft(sentence, "-*• ")
	sentence = strings.Join(strings.Fields(sentence), " ")
	return sentence
}

func sentenceText(candidate sentenceCandidate) string {
	text := strings.TrimSpace(candidate.text)
	if text == "" {
		return text
	}
	if len(text) > 180 {
		text = strings.TrimSpace(text[:177]) + "..."
	}
	if !strings.HasSuffix(text, ".") && !strings.HasSuffix(text, "!") && !strings.HasSuffix(text, "?") {
		text += "."
	}
	return text
}

func uniqueSnippets(citations []domain.Citation) []string {
	seen := make(map[string]struct{}, len(citations))
	snippets := make([]string, 0, len(citations))
	for _, citation := range citations {
		snippet := strings.TrimSpace(citation.Snippet)
		if snippet == "" {
			continue
		}
		if _, ok := seen[snippet]; ok {
			continue
		}
		seen[snippet] = struct{}{}
		snippets = append(snippets, snippet)
	}
	return snippets
}
