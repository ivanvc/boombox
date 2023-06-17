package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ivanvc/boombox/internal/ui/common"
	"github.com/ivanvc/boombox/internal/ui/common/state"
)

const timeout = time.Second * 5

// The Completed view holds the last screen after the pod has been terminated.
type Completed struct {
	timer  timer.Model
	common *common.Common
}

// NewCompleted returns a new Completed instance.
func NewCompleted(common *common.Common) *Completed {
	return &Completed{common: common}
}

// Init implements tea.Model.
func (c *Completed) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (c *Completed) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case state.StateChangedMsg:
		if msg.State == state.PodTerminated {
			c.timer = timer.NewWithInterval(timeout, time.Second)
			return c, c.timer.Init()
		}
	case tea.KeyMsg:
		return c, tea.Quit
	case timer.TimeoutMsg:
		return c, tea.Quit
	case timer.TickMsg:
		var cmd tea.Cmd
		c.timer, cmd = c.timer.Update(msg)
		return c, cmd
	}

	return c, nil
}

// View implements tea.Model.
func (c *Completed) View() string {
	var timerView string

	if c.timer.Timedout() {
		timerView = "Buh bye!"
	} else {
		timerView = "Exiting in " + c.timer.View()
	}

	return c.common.RenderCentered(
		fmt.Sprintf("%s\n\n%s\n%s\n",
			common.LogoSprite[0],
			timerView,
			"Press any key to exit",
		),
	)
}
