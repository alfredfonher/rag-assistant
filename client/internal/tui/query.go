package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"rag-assistant/client/internal/serviceapi"
)

type queryModel struct {
	input   textinput.Model
	client  *serviceapi.Client
	answer  string
	cites   []serviceapi.Citation
	err     string
	loading bool
	width   int
}

func newQueryModel(client *serviceapi.Client) queryModel {
	ti := textinput.New()
	ti.Placeholder = "Ask something..."
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 60

	return queryModel{
		input:  ti,
		client: client,
	}
}

func (m queryModel) Init() tea.Cmd {
	return textinput.Blink
}

type queryResultMsg struct {
	answer string
	cites  []serviceapi.Citation
	err    error
}

func (m queryModel) submit() tea.Cmd {
	return func() tea.Msg {
		q := strings.TrimSpace(m.input.Value())
		if q == "" {
			return queryResultMsg{err: fmt.Errorf("empty query")}
		}
		resp, err := m.client.Query(context.Background(), serviceapi.QueryRequest{Query: q})
		if err != nil {
			return queryResultMsg{err: err}
		}
		if resp.Error != nil {
			return queryResultMsg{err: fmt.Errorf("%s", resp.Error.Message)}
		}
		return queryResultMsg{answer: resp.Answer, cites: resp.Citations}
	}
}

func (m queryModel) Update(msg tea.Msg) (queryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.loading = true
			m.answer = ""
			m.cites = nil
			m.err = ""
			return m, m.submit()
		case tea.KeyEsc:
			m.input.Reset()
			m.answer = ""
			m.cites = nil
			m.err = ""
			return m, nil
		}
	case queryResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.answer = msg.answer
			m.cites = msg.cites
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m queryModel) View() string {
	var b strings.Builder

	b.WriteString(m.input.View() + "\n")

	if m.loading {
		b.WriteString(mutedStyle.Render("  thinking..."))
	} else if m.err != "" {
		b.WriteString(errorStyle.Render("  ✗ "+m.err))
	} else if m.answer != "" {
		b.WriteString(answerStyle.Render("  "+m.answer))
		if len(m.cites) > 0 {
			b.WriteString("\n")
			for _, c := range m.cites {
				b.WriteString(citationStyle.Render(fmt.Sprintf("  → %s/%s", c.DocumentID, c.ChunkID)))
			}
		}
	}

	b.WriteString(helpStyle.Render("\n  enter: submit  esc: clear  tab: switch tab  q: quit"))
	return b.String()
}
