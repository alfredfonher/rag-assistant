package query

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMutationRollback(t *testing.T) {
	tests := []struct {
		name    string
		seed    string
		mutate  func(*MemoryRetriever) (*MutationToken, error)
		after   string
		want    string
		deleted bool
	}{
		{"replace insert", "", func(r *MemoryRetriever) (*MutationToken, error) {
			_, token, err := r.ReplaceDocument(context.Background(), "doc", "new")
			return token, err
		}, "new", "", true},
		{"replace existing", "old", func(r *MemoryRetriever) (*MutationToken, error) {
			_, token, err := r.ReplaceDocument(context.Background(), "doc", "new")
			return token, err
		}, "new", "old", false},
		{"delete existing", "old", func(r *MemoryRetriever) (*MutationToken, error) {
			return r.DeleteDocument(context.Background(), "doc")
		}, "", "old", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMemoryRetriever()
			if tt.seed != "" {
				_, _ = r.IndexDocument(context.Background(), "doc", tt.seed)
			}
			token, err := tt.mutate(r)
			if err != nil || token == nil {
				t.Fatalf("mutation failed: token=%v err=%v", token, err)
			}
			mutated := r.documents["doc"]
			if (tt.after == "" && mutated != nil) || (tt.after != "" && (len(mutated) != 1 || mutated[0].text != tt.after)) {
				t.Fatalf("mutation did not publish expected state: %#v", r.documents)
			}
			if err := token.Rollback(context.Background()); err != nil {
				t.Fatalf("rollback failed: %v", err)
			}
			chunks, exists := r.documents["doc"]
			if exists == tt.deleted || (!tt.deleted && (len(chunks) != 1 || chunks[0].text != tt.want)) {
				t.Fatalf("unexpected restored state: %#v", r.documents)
			}
		})
	}
}

func TestMutationValidationAndAbsentDelete(t *testing.T) {
	r := NewMemoryRetriever()
	if _, _, err := r.ReplaceDocument(context.Background(), "doc", " \n "); !errors.Is(err, ErrEmptyDocument) {
		t.Fatalf("expected empty document error, got %v", err)
	}
	store := &recordingStore{}
	r, _ = newMemoryRetriever(nil, store)
	token, err := r.DeleteDocument(context.Background(), "missing")
	if err != nil || token != nil || store.saves != 0 {
		t.Fatalf("absent delete was not a no-op: token=%v saves=%d err=%v", token, store.saves, err)
	}
}

func TestMutationPersistenceFailuresAreAtomic(t *testing.T) {
	storeErr := errors.New("disk full")
	store := &recordingStore{fail: storeErr}
	r, _ := newMemoryRetriever(nil, store)
	if _, _, err := r.ReplaceDocument(context.Background(), "doc", "new"); !errors.Is(err, ErrPersistenceUnavailable) || !errors.Is(err, storeErr) {
		t.Fatalf("expected both persistence errors, got %v", err)
	}
	if len(r.documents) != 0 || r.revision != 0 {
		t.Fatalf("failed replace changed state: %#v revision=%d", r.documents, r.revision)
	}

	store.fail = nil
	_, token, _ := r.ReplaceDocument(context.Background(), "doc", "new")
	store.fail = storeErr
	if err := token.Rollback(context.Background()); !errors.Is(err, storeErr) {
		t.Fatalf("expected rollback save failure, got %v", err)
	}
	if r.documents["doc"][0].text != "new" || r.revision != 1 {
		t.Fatalf("failed rollback changed state: %#v revision=%d", r.documents, r.revision)
	}
	if err := token.Rollback(context.Background()); !errors.Is(err, ErrMutationUsed) {
		t.Fatalf("failed rollback must consume token, got %v", err)
	}
}

func TestMutationTokenStaleAndUsed(t *testing.T) {
	r := NewMemoryRetriever()
	_, token, _ := r.ReplaceDocument(context.Background(), "doc", "first")
	_, _ = r.IndexDocument(context.Background(), "other", "later")
	if err := token.Rollback(context.Background()); !errors.Is(err, ErrMutationStale) {
		t.Fatalf("expected stale token, got %v", err)
	}
	if err := token.Rollback(context.Background()); !errors.Is(err, ErrMutationUsed) {
		t.Fatalf("stale attempt must consume token, got %v", err)
	}
	_, token, _ = r.ReplaceDocument(context.Background(), "doc", "second")
	if err := token.Rollback(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := token.Rollback(context.Background()); !errors.Is(err, ErrMutationUsed) {
		t.Fatalf("expected used token, got %v", err)
	}
}

func TestMutationContext(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	r := NewMemoryRetriever()
	if _, _, err := r.ReplaceDocument(canceled, "doc", "new"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected entry cancellation, got %v", err)
	}
	if _, err := r.DeleteDocument(canceled, "doc"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected delete cancellation, got %v", err)
	}
	_, token, _ := r.ReplaceDocument(context.Background(), "doc", "new")
	if err := token.Rollback(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected rollback cancellation, got %v", err)
	}
	if err := token.Rollback(context.Background()); err != nil {
		t.Fatalf("entry cancellation consumed token: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	r, _ = NewMemoryRetrieverWithProvider(cancelingProvider{cancel: cancel})
	if _, _, err := r.ReplaceDocument(ctx, "doc", "new"); !errors.Is(err, context.Canceled) || len(r.documents) != 0 {
		t.Fatalf("expected post-embedding cancellation without publish, state=%#v err=%v", r.documents, err)
	}
}

func TestMutationRestartUsesVersionOneAndFreshRevision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.json")
	r, _ := NewPersistentMemoryRetriever(path)
	_, _, _ = r.ReplaceDocument(context.Background(), "doc", "text")
	data, _ := os.ReadFile(path)
	var state persistedRetrieverState
	if err := json.Unmarshal(data, &state); err != nil || state.Version != 1 {
		t.Fatalf("expected version 1 persistence, state=%#v err=%v", state, err)
	}
	reloaded, err := NewPersistentMemoryRetriever(path)
	if err != nil || reloaded.revision != 0 || reloaded.documents["doc"][0].text != "text" {
		t.Fatalf("unexpected reload: revision=%d documents=%#v err=%v", reloaded.revision, reloaded.documents, err)
	}
}

type recordingStore struct {
	saves int
	fail  error
}

func (*recordingStore) Load(context.Context) (map[string][]memoryChunk, error) {
	return map[string][]memoryChunk{}, nil
}

func (s *recordingStore) Save(context.Context, map[string][]memoryChunk) error {
	s.saves++
	return s.fail
}

type cancelingProvider struct{ cancel context.CancelFunc }

func (p cancelingProvider) EmbedDocument(context.Context, string) ([]float64, error) {
	p.cancel()
	return []float64{1}, nil
}

func (cancelingProvider) EmbedQuery(context.Context, string) ([]float64, error) {
	return []float64{1}, nil
}
