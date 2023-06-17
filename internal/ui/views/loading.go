package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ivanvc/boombox/internal/ui/common"
	"github.com/ivanvc/boombox/internal/ui/common/state"
)

type spriteSub chan struct{}
type spriteIndexChangeMsg struct{}

// Loading is the view that displays a loading message and a spinner.
type Loading struct {
	states      []state.State
	common      *common.Common
	spriteSub   spriteSub
	spriteIndex uint8
	durations   []time.Duration
	startTime   time.Time
	spinner     spinner.Model
}

// NewLoading returns a new Loading instance.
func NewLoading(cmn *common.Common) *Loading {
	const statesLen = 5
	states := make([]state.State, statesLen)
	states[statesLen-1] = cmn.State
	s := spinner.New()
	s.Style = common.SecondaryTextStyle

	return &Loading{
		states:    states,
		durations: make([]time.Duration, statesLen),
		common:    cmn,
		spriteSub: make(spriteSub),
		startTime: time.Now(),
		spinner:   s,
	}
}

// Init implements tea.Model.
func (l *Loading) Init() tea.Cmd {
	return tea.Batch(generateSpriteChanges(l.spriteSub),
		waitForSpriteChanges(l.spriteSub), l.spinner.Tick)
}

// Update implements tea.Model.
func (l *Loading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spriteIndexChangeMsg:
		l.spriteIndex = (l.spriteIndex + 1) % uint8(len(common.LogoSprite))
		return l, waitForSpriteChanges(l.spriteSub)
	case state.StateChangedMsg:
		l.durations[len(l.durations)-1] = time.Now().Sub(l.startTime).Round(time.Millisecond)
		l.startTime = time.Now()
		l.states = append(l.states[1:], msg.State)
		l.durations = append(l.durations[1:], 0)
	case spinner.TickMsg:
		var cmd tea.Cmd
		l.spinner, cmd = l.spinner.Update(msg)
		return l, cmd
	}

	return l, nil
}

// View implements tea.Model.
func (l *Loading) View() string {
	return l.common.RenderCentered(
		fmt.Sprintf(
			"%s\n\n%s\n",
			common.LogoSprite[l.spriteIndex],
			common.BoxContainerStyle.Width(50).Render(
				l.renderStates(),
				//lipgloss.PlaceHorizontal(50, lipgloss.Left, l.renderStates()),
			),
		),
	)
}

func (l *Loading) renderStates() string {
	var b strings.Builder

	for i := len(l.states) - 1; i >= 0; i-- {
		var indicator, text, duration string
		if i == len(l.states)-1 {
			indicator = l.spinner.View()
			text = l.states[i].String()
			duration = time.Now().Sub(l.startTime).Round(time.Millisecond).String()
		} else if l.states[i] > state.Unknown {
			indicator = common.CheckMark.Render()
			text = l.states[i].String()
			duration = l.durations[i].String()
		}
		b.WriteString(fmt.Sprintf("%s %s %s", indicator, text, common.SecondaryTextStyle.Render(duration)))
		if i > 0 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

func generateSpriteChanges(sub spriteSub) tea.Cmd {
	return func() tea.Msg {
		for {
			<-time.After(time.Millisecond * 500)
			sub <- struct{}{}
		}
	}
}

func waitForSpriteChanges(sub spriteSub) tea.Cmd {
	return func() tea.Msg {
		return spriteIndexChangeMsg(<-sub)
	}
}
