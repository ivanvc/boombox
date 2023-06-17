package common

import "github.com/charmbracelet/lipgloss"

// Lipgloss styles used across the UI
var (
	SecondaryTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	BoxContainerStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("4")).Padding(0, 1)
	LogoStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	LogoActivityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	ErrorStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("197"))

	CheckMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)
