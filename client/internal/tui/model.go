package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"rag-assistant/client/internal/serviceapi"
)

type tab int

const (
	tabQuery tab = iota
	tabIngest
	tabStatus
)

type model struct {
	active  tab
	query   queryModel
	ingest  ingestModel
	status  statusModel
}

func New(client *serviceapi.Client) model {
	return model{
		active: tabQuery,
		query:  newQueryModel(client),
		ingest: newIngestModel(client),
		status: newStatusModel(client),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.query.Init(),
		m.status.Init(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
		// Tab switching only when not in text input
		if msg.Type == tea.KeyTab {
			switch m.active {
			case tabQuery:
				m.active = tabIngest
			case tabIngest:
				m.active = tabStatus
			case tabStatus:
				m.active = tabQuery
			}
			return m, nil
		}
		if msg.String() == "q" && m.active == tabStatus {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.query.width = msg.Width
		m.status.width = msg.Width
	}

	// Delegate to active tab
	switch m.active {
	case tabQuery:
		newQuery, qCmd := m.query.Update(msg)
		m.query = newQuery
		return m, qCmd
	case tabIngest:
		newIngest, iCmd := m.ingest.Update(msg)
		m.ingest = newIngest
		return m, iCmd
	case tabStatus:
		newStatus, sCmd := m.status.Update(msg)
		m.status = newStatus
		return m, sCmd
	}
	return m, nil
}

func (m model) View() string {
	// Tabs
	tabs := "  "
	for i, name := range []string{"Query", "Ingest", "Status"} {
		if tab(i) == m.active {
			tabs += activeTabStyle.Render(name)
		} else {
			tabs += tabStyle.Render(name)
		}
	}

	// Content
	var content string
	switch m.active {
	case tabQuery:
		content = m.query.View()
	case tabIngest:
		content = m.ingest.View()
	case tabStatus:
		content = m.status.View()
	}

	return "\n" + titleStyle.Render("rag-assistant") + "\n\n" + tabs + "\n\n" + content
}

func Run() {
	serviceURL := os.Getenv("RAG_SERVICE_URL")
	if serviceURL == "" {
		serviceURL = "http://127.0.0.1:8080"
	}
	client := serviceapi.New(serviceURL)
	p := tea.NewProgram(New(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
