package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"rag-assistant/client/internal/serviceapi"
)

type ingestModel struct {
	input   textinput.Model
	client  *serviceapi.Client
	docID   string
	status  string
	err     string
	loading bool
}

func newIngestModel(client *serviceapi.Client) ingestModel {
	ti := textinput.New()
	ti.Placeholder = "guides/document.md"
	ti.CharLimit = 512
	ti.Width = 60

	return ingestModel{
		input:  ti,
		client: client,
	}
}

func (m ingestModel) Init() tea.Cmd {
	return nil
}

type ingestResultMsg struct {
	docID  string
	status string
	err    error
}

func (m ingestModel) submit() tea.Cmd {
	return func() tea.Msg {
		path := strings.TrimSpace(m.input.Value())
		if path == "" {
			return ingestResultMsg{err: fmt.Errorf("empty path")}
		}
		resp, err := m.client.Ingest(context.Background(), serviceapi.IngestRequest{Path: path})
		if err != nil {
			return ingestResultMsg{err: err}
		}
		if resp.Error != nil {
			return ingestResultMsg{err: fmt.Errorf("%s", resp.Error.Message)}
		}
		return ingestResultMsg{docID: resp.DocumentID, status: resp.State}
	}
}

func (m ingestModel) Update(msg tea.Msg) (ingestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.loading = true
			m.docID = ""
			m.status = ""
			m.err = ""
			return m, m.submit()
		case tea.KeyEsc:
			m.input.Reset()
			m.docID = ""
			m.status = ""
			m.err = ""
			return m, nil
		}
	case ingestResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.docID = msg.docID
			m.status = msg.status
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m ingestModel) View() string {
	var b strings.Builder

	b.WriteString("Path relative to the configured ingest root\n")
	b.WriteString(m.input.View() + "\n")

	if m.loading {
		b.WriteString(mutedStyle.Render("  ingesting..."))
	} else if m.err != "" {
		b.WriteString(errorStyle.Render("  ✗ " + m.err))
	} else if m.docID != "" {
		b.WriteString(successStyle.Render(fmt.Sprintf("  ✓ %s (%s)", m.docID, m.status)))
	}

	b.WriteString(helpStyle.Render("\n  enter: ingest  esc: clear  tab: switch tab  q: quit"))
	return b.String()
}
