package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primary   = lipgloss.Color("#7C3AED")
	secondary = lipgloss.Color("#06B6D4")
	success   = lipgloss.Color("#10B981")
	danger    = lipgloss.Color("#EF4444")
	muted     = lipgloss.Color("#6B7280")
	surface   = lipgloss.Color("#1E1E2E")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			MarginBottom(1)

	tabStyle = lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 2)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primary).
			Bold(true).
			Padding(0, 2)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondary).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(danger).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(success).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(muted)

	answerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginTop(1)

	citationStyle = lipgloss.NewStyle().
			Foreground(secondary).
			Italic(true).
			MarginLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(muted).
			MarginTop(1)
)
