package common

import "github.com/charmbracelet/lipgloss"

func (c *Common) RenderCentered(text string) string {
	return lipgloss.PlaceHorizontal(
		c.Width,
		lipgloss.Center,
		lipgloss.PlaceVertical(
			c.Height,
			lipgloss.Center,
			text,
		),
	)
}
