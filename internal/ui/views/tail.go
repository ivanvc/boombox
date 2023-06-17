package views

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/ivanvc/boombox/internal/ui/actions"
	"github.com/ivanvc/boombox/internal/ui/common"
)

type Tail struct {
	logLines     []string
	logLinesChan chan string
	spinner      spinner.Model
	viewport     viewport.Model
	ready        bool
	common       *common.Common
}

type logLineMsg string

// Returns a new Tail.
func NewTail(cmn *common.Common) *Tail {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = common.SecondaryTextStyle

	t := &Tail{
		logLinesChan: make(chan string),
		logLines:     make([]string, 0),
		spinner:      s,
		common:       cmn,
	}
	return t
}

// Init implements tea.Model.
func (t *Tail) Init() tea.Cmd {
	return tea.Batch(waitForLines(t.logLinesChan), t.spinner.Tick)
}

// Update implements tea.Model.
func (t *Tail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debugf("Tail got msg: %T", msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		verticalMarginHeight := len(strings.Split(common.MiniLogo, "\n"))
		if !t.ready {
			t.ready = true
			t.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			t.viewport.MouseWheelEnabled = false
			t.viewport.YPosition = verticalMarginHeight
			t.viewport.Style = common.BoxContainerStyle.Width(msg.Width)
			t.viewport.SetContent(t.viewportContent())
		} else {
			t.viewport.Width = msg.Width
			t.viewport.Style.Width(msg.Width)
			t.viewport.Height = msg.Height - verticalMarginHeight
		}
	case logLineMsg:
		cleanLine := strings.Map(func(r rune) rune {
			if unicode.IsGraphic(r) {
				return r
			}
			return -1
		}, string(msg))
		t.logLines = append(t.logLines, cleanLine)
		if len(t.logLines) > t.viewport.Height {
			t.logLines = t.logLines[len(t.logLines)-t.viewport.Height:]
		}
		return t, waitForLines(t.logLinesChan)
	case actions.TriggerStartLogTailMsg:
		return t, t.common.Actions.TailInitContainerLogs(msg.Pod, t.logLinesChan)
	case spinner.TickMsg:
		var cmd tea.Cmd
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd
	}

	return t, nil
}

// View implements tea.Model.
func (t *Tail) View() string {
	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		common.MiniLogo,
		" ",
		t.spinner.View(),
		" ",
		"Waiting for setup to complete",
	)
	if !t.ready {
		return header
	}

	t.viewport.SetContent(t.viewportContent())
	t.viewport.GotoBottom()

	return fmt.Sprintf("%s\n%s", header, t.viewport.View())
}

func (t *Tail) viewportContent() string {
	return strings.Join(t.logLines, "\n")
}

func waitForLines(sub chan string) tea.Cmd {
	return func() tea.Msg {
		line := <-sub
		log.Debug("Got tail line", "line", line)
		return logLineMsg(line)
	}
}
