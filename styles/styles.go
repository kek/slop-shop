package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Primary   = lipgloss.Color("#7D56F4")
	Secondary = lipgloss.Color("#8B5CF6")
	Accent    = lipgloss.Color("#A855F7")
	Success   = lipgloss.Color("#10B981")
	Warning   = lipgloss.Color("#F59E0B")
	ErrColor  = lipgloss.Color("#EF4444")
	Info      = lipgloss.Color("#3B82F6")
	Muted     = lipgloss.Color("#6B7280")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			MarginLeft(2).
			MarginBottom(1)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true).
			MarginBottom(1)

	PromptStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true).
			MarginLeft(2)

	ResponseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB")).
			MarginLeft(2).
			MarginTop(1).
			MarginBottom(1)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Info).
			MarginLeft(2)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			MarginLeft(2)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			MarginLeft(2)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrColor).
			MarginLeft(2)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginLeft(2)

	SeparatorStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginLeft(2).
			MarginTop(1).
			MarginBottom(1)

	ToolStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true).
			MarginLeft(2)

	ToolResultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB")).
			MarginLeft(4).
			MarginTop(1)

	REPLPromptStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	REPLInputStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	UserStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")). // Green
			Bold(true)

	AssistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")). // Blue
			Italic(true)
)
