package query

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"rag-assistant/service/internal/domain"
)

type ConversationTurn struct {
	Query     string            `json:"query"`
	State     string            `json:"state"`
	Answer    string            `json:"answer"`
	Citations []domain.Citation `json:"citations,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type Conversation struct {
	ID    string             `json:"id"`
	Turns []ConversationTurn `json:"turns"`
}

type ConversationStore interface {
	Append(context.Context, string, ConversationTurn) error
}

type conversationFile struct {
	Conversations []Conversation `json:"conversations"`
}

type FileConversationStore struct {
	mu            sync.Mutex
	path          string
	conversations map[string]Conversation
}

func NewConversationStore(path string) (*FileConversationStore, error) {
	store := &FileConversationStore{path: path, conversations: make(map[string]Conversation)}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *FileConversationStore) Append(ctx context.Context, id string, turn ConversationTurn) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("conversation id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	conversation := s.conversations[id]
	conversation.ID = id
	conversation.Turns = append(conversation.Turns, cloneTurn(turn))
	s.conversations[id] = conversation
	return s.saveLocked()
}

func (s *FileConversationStore) Get(id string) (Conversation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conversation, ok := s.conversations[id]
	if !ok {
		return Conversation{}, false
	}
	conversation.Turns = append([]ConversationTurn(nil), conversation.Turns...)
	return conversation, true
}

func (s *FileConversationStore) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var file conversationFile
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	for _, conversation := range file.Conversations {
		s.conversations[conversation.ID] = conversation
	}
	return nil
}

func (s *FileConversationStore) saveLocked() error {
	file := conversationFile{Conversations: make([]Conversation, 0, len(s.conversations))}
	for _, conversation := range s.conversations {
		file.Conversations = append(file.Conversations, conversation)
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(s.path), ".conversation-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, s.path)
}

// List returns summaries of all conversations.
func (s *FileConversationStore) List() []Conversation {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Conversation, 0, len(s.conversations))
	for _, c := range s.conversations {
		cpy := Conversation{ID: c.ID, Turns: append([]ConversationTurn(nil), c.Turns...)}
		out = append(out, cpy)
	}
	return out
}

// Rename is a no-op placeholder (conversations don't have names in the current model).
func (s *FileConversationStore) Rename(_, _ string) error {
	return nil
}

// Delete removes a conversation by ID.
func (s *FileConversationStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.conversations[id]; !ok {
		return fmt.Errorf("conversation not found")
	}
	delete(s.conversations, id)
	return s.saveLocked()
}

func NewConversationID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("conv-%d", time.Now().UnixNano())
	}
	return "conv-" + hex.EncodeToString(bytes[:])
}

func cloneTurn(turn ConversationTurn) ConversationTurn {
	turn.Citations = append([]domain.Citation(nil), turn.Citations...)
	return turn
}
