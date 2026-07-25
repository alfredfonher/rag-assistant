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

	// Tab -> status
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m3 := m2.(model)
	if m3.active != tabStatus {
		t.Fatal("should switch to status tab")
	}

	// Tab -> back to query
	m4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	m5 := m4.(model)
	if m5.active != tabQuery {
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

func TestStatusTabView(t *testing.T) {
	m := New(serviceapi.New("http://localhost:9999"))
	m.active = tabStatus

	view := m.View()
	if view == "" {
		t.Fatal("view should not be empty")
	}
}
