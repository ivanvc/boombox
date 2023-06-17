package views

import tea "github.com/charmbracelet/bubbletea"

// View is the interface of an UI view.
type View interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}
