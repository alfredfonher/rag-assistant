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
	stream  bool
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

type streamFrameMsg struct {
	frame serviceapi.QueryResponse
	next  func() (serviceapi.QueryResponse, error)
	done  bool
}

func (m queryModel) submitStream() tea.Cmd {
	return func() tea.Msg {
		q := strings.TrimSpace(m.input.Value())
		if q == "" {
			return queryResultMsg{err: fmt.Errorf("empty query")}
		}
		stream, err := m.client.StreamQuery(context.Background(), serviceapi.QueryRequest{Query: q})
		if err != nil {
			return queryResultMsg{err: err}
		}
		// Read first frame
		frame, err := stream.Next()
		if err != nil {
			stream.Close()
			return queryResultMsg{err: err}
		}
		return streamFrameMsg{
			frame: frame,
			next:  stream.Next,
			done:  false,
		}
	}
}

func (m queryModel) Update(msg tea.Msg) (queryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.loading = true
			m.stream = true
			m.answer = ""
			m.cites = nil
			m.err = ""
			return m, m.submitStream()
		case tea.KeyEsc:
			m.input.Reset()
			m.answer = ""
			m.cites = nil
			m.err = ""
			m.stream = false
			return m, nil
		}
	case streamFrameMsg:
		m.loading = false
		// Accumulate answer from content frames
		if msg.frame.Answer != "" {
			m.answer = msg.frame.Answer
		}
		// Capture citations
		if len(msg.frame.Citations) > 0 {
			m.cites = msg.frame.Citations
		}
		// Check for errors
		if msg.frame.Error != nil {
			m.err = msg.frame.Error.Message
			m.stream = false
			return m, nil
		}
		// Done event
		if msg.frame.Event == "done" {
			m.stream = false
			return m, nil
		}
		// Read next frame
		if msg.next != nil {
			return m, func() tea.Msg {
				frame, err := msg.next()
				if err != nil {
					return queryResultMsg{err: err}
				}
				return streamFrameMsg{
					frame: frame,
					next:  msg.next,
					done:  false,
				}
			}
		}
		m.stream = false
	case queryResultMsg:
		m.loading = false
		m.stream = false
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
		b.WriteString(mutedStyle.Render("  connecting..."))
	} else if m.stream {
		b.WriteString(mutedStyle.Render("  streaming..."))
		if m.answer != "" {
			b.WriteString("\n" + answerStyle.Render("  "+m.answer+"▌"))
		}
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
