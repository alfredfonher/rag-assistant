package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"rag-assistant/client/internal/serviceapi"
)

type statusModel struct {
	client  *serviceapi.Client
	health  *serviceapi.HealthResponse
	ready   *serviceapi.ReadinessResponse
	err     string
	loading bool
	width   int
}

func newStatusModel(client *serviceapi.Client) statusModel {
	return statusModel{client: client}
}

type statusResultMsg struct {
	health *serviceapi.HealthResponse
	ready  *serviceapi.ReadinessResponse
	err    error
}

func (m statusModel) refresh() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		health, err := m.client.Health(ctx)
		if err != nil {
			return statusResultMsg{err: err}
		}
		ready, err := m.client.Ready(ctx)
		if err != nil {
			return statusResultMsg{health: &health, err: err}
		}
		return statusResultMsg{health: &health, ready: &ready}
	}
}

func (m statusModel) Init() tea.Cmd {
	return m.refresh()
}

func (m statusModel) Update(msg tea.Msg) (statusModel, tea.Cmd) {
	switch msg := msg.(type) {
	case statusResultMsg:
		m.loading = false
		if msg.health != nil {
			m.health = msg.health
		}
		if msg.ready != nil {
			m.ready = msg.ready
		}
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.err = ""
		}
		return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg { return refreshTick{} })
	case refreshTick:
		return m, m.refresh()
	}
	return m, nil
}

type refreshTick struct{}

func (m statusModel) View() string {
	var b strings.Builder

	b.WriteString("  Status Dashboard\n\n")

	if m.health != nil {
		alive := successStyle.Render("● UP")
		if !m.health.Alive {
			alive = errorStyle.Render("● DOWN")
		}
		b.WriteString(fmt.Sprintf("  Health:   %s  (%s)\n", alive, m.health.Service))
	} else if m.loading {
		b.WriteString("  Health:   " + mutedStyle.Render("loading...") + "\n")
	} else if m.err != "" {
		b.WriteString("  Health:   " + errorStyle.Render("● ERROR") + "\n")
	}

	if m.ready != nil {
		ready := successStyle.Render("● READY")
		if !m.ready.Ready {
			ready = errorStyle.Render("● NOT READY")
		}
		b.WriteString(fmt.Sprintf("  Ready:    %s\n", ready))
		if len(m.ready.Reasons) > 0 {
			for _, r := range m.ready.Reasons {
				b.WriteString(fmt.Sprintf("            %s %s: %s\n", mutedStyle.Render("·"), r.Dependency, r.Detail))
			}
		}
	} else if m.loading {
		b.WriteString("  Ready:    " + mutedStyle.Render("loading...") + "\n")
	}

	b.WriteString(helpStyle.Render("\n  auto-refresh 5s  tab: switch tab  q: quit"))
	return b.String()
}
