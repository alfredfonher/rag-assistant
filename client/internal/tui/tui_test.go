package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"rag-assistant/client/internal/serviceapi"
)

func TestTabSwitching(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))

	if m.active != tabQuery {
		t.Fatal("default tab should be query")
	}

	// Tab -> ingest
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m3 := m2.(model)
	if m3.active != tabIngest {
		t.Fatal("should switch to ingest tab")
	}

	// Tab -> status
	m4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	m5 := m4.(model)
	if m5.active != tabStatus {
		t.Fatal("should switch to status tab")
	}

	// Tab -> back to query
	m6, _ := m5.Update(tea.KeyMsg{Type: tea.KeyTab})
	m7 := m6.(model)
	if m7.active != tabQuery {
		t.Fatal("should switch back to query tab")
	}
}

func TestCtrlCQuits(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Fatal("Ctrl+C should produce a quit command")
	}
}

func TestQueryClearOnEsc(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))

	// Type something
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h', 'i'}})
	m3 := m2.(model)

	// Esc should clear
	m4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m5 := m4.(model)

	if m5.query.input.Value() != "" {
		t.Fatal("Esc should clear the input")
	}
}

func TestQueryErrorDisplay(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))

	// Simulate error result
	m.query, _ = m.query.Update(queryResultMsg{err: errors.New("connection refused")})

	if m.query.err == "" {
		t.Fatal("error should be displayed")
	}
	if m.query.loading {
		t.Fatal("loading should be false after error")
	}
}

func TestQueryAnswerDisplay(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))

	// Simulate answer result
	m.query, _ = m.query.Update(queryResultMsg{
		answer: "cobalt-seven",
		cites:  []serviceapi.Citation{{DocumentID: "doc1", ChunkID: "c1"}},
	})

	if m.query.answer != "cobalt-seven" {
		t.Fatalf("expected 'cobalt-seven', got %q", m.query.answer)
	}
	if len(m.query.cites) != 1 {
		t.Fatal("expected 1 citation")
	}
}

func TestQueryLoadingState(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))
	m.query.loading = true

	if m.query.View() == "" {
		t.Fatal("loading view should not be empty")
	}
}

func TestStreamFrameAccumulatesAnswer(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))
	m.query.stream = true

	callCount := 0
	fakeNext := func() (serviceapi.QueryResponse, error) {
		callCount++
		switch callCount {
		case 1:
			return serviceapi.QueryResponse{Kind: "content", Answer: "Hello "}, nil
		case 2:
			return serviceapi.QueryResponse{Kind: "content", Answer: "Hello World"}, nil
		default:
			return serviceapi.QueryResponse{Event: "done"}, nil
		}
	}

	// Content frame
	m.query, _ = m.query.Update(streamFrameMsg{
		frame: serviceapi.QueryResponse{Kind: "content", Answer: "Hello "},
		next:  fakeNext,
	})
	if m.query.answer != "Hello " {
		t.Fatalf("expected 'Hello ', got %q", m.query.answer)
	}
	if !m.query.stream {
		t.Fatal("stream should still be active")
	}

	// More content
	m.query, _ = m.query.Update(streamFrameMsg{
		frame: serviceapi.QueryResponse{Kind: "content", Answer: "Hello World"},
		next:  fakeNext,
	})
	if m.query.answer != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", m.query.answer)
	}

	// Done frame
	m.query, _ = m.query.Update(streamFrameMsg{
		frame: serviceapi.QueryResponse{Event: "done"},
		next:  fakeNext,
	})
	if m.query.stream {
		t.Fatal("stream should be done")
	}
}

func TestStreamFrameError(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))

	m.query, _ = m.query.Update(streamFrameMsg{
		frame: serviceapi.QueryResponse{
			Error: &serviceapi.APIError{Message: "model unavailable"},
		},
	})

	if m.query.err == "" {
		t.Fatal("error should be set")
	}
	if m.query.stream {
		t.Fatal("stream should stop on error")
	}
}

func TestStreamFrameCitations(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))

	m.query, _ = m.query.Update(streamFrameMsg{
		frame: serviceapi.QueryResponse{
			Citations: []serviceapi.Citation{{DocumentID: "doc1", ChunkID: "c1"}},
		},
	})

	if len(m.query.cites) != 1 {
		t.Fatal("citations should be captured")
	}
}

func TestStatusTabView(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))
	m.active = tabStatus

	view := m.View()
	if view == "" {
		t.Fatal("view should not be empty")
	}
}

func TestIngestSuccess(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))

	m.ingest, _ = m.ingest.Update(ingestResultMsg{
		docID:  "readme",
		status: "indexed",
	})

	if m.ingest.docID != "readme" {
		t.Fatalf("expected 'readme', got %q", m.ingest.docID)
	}
	if m.ingest.status != "indexed" {
		t.Fatalf("expected 'indexed', got %q", m.ingest.status)
	}
}

func TestIngestError(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))

	m.ingest, _ = m.ingest.Update(ingestResultMsg{
		err: errors.New("unsupported file type"),
	})

	if m.ingest.err == "" {
		t.Fatal("error should be displayed")
	}
	if m.ingest.loading {
		t.Fatal("loading should be false after error")
	}
}

func TestIngestClearOnEsc(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))

	// Set some state
	m.ingest, _ = m.ingest.Update(ingestResultMsg{docID: "test", status: "indexed"})
	m.ingest.input.SetValue("/some/path.md")

	// Esc should clear
	m2, _ := m.ingest.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m2.input.Value() != "" {
		t.Fatal("Esc should clear the input")
	}
	if m2.docID != "" {
		t.Fatal("Esc should clear the doc ID")
	}
}

func TestIngestTabView(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))
	m.active = tabIngest

	view := m.View()
	if view == "" {
		t.Fatal("view should not be empty")
	}
}
