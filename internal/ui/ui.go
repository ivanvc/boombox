package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/ivanvc/boombox/internal/config"
	k8s "github.com/ivanvc/boombox/internal/services/kubernetes"
	"github.com/ivanvc/boombox/internal/ui/actions"
	"github.com/ivanvc/boombox/internal/ui/common"
	"github.com/ivanvc/boombox/internal/ui/common/state"
	"github.com/ivanvc/boombox/internal/ui/views"
)

type view int

const (
	loadingView view = iota
	tailView
	completedView
)

// UI holds the main UI of the application.
type UI struct {
	common     *common.Common
	views      []tea.Model
	activeView view
	sizeChan   k8s.SizeChan
	error      error
	createdPVC bool
}

// New returns a new UI.
func New(common *common.Common) *UI {
	return &UI{
		common:   common,
		views:    make([]tea.Model, 3),
		sizeChan: make(k8s.SizeChan, 1),
	}
}

// Init implements tea.Model.
func (ui *UI) Init() tea.Cmd {
	ui.common.State = state.FetchingPod
	ui.views[loadingView] = views.NewLoading(ui.common)
	ui.views[tailView] = views.NewTail(ui.common)
	ui.views[completedView] = views.NewCompleted(ui.common)
	cmds := []tea.Cmd{ui.common.Actions.FetchPod(ui.common.User)}
	for _, v := range ui.views {
		cmds = append(cmds, v.Init())
	}
	ui.activeView = loadingView
	return tea.Batch(cmds...)
}

// Update implements tea.Model.
func (ui *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debugf("msg received: %T", msg)
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ui.common.Width = msg.Width
		ui.common.Height = msg.Height
		if ui.common.State == state.PodRunning {
			ui.sizeChan <- remotecommand.TerminalSize{
				Width:  uint16(msg.Width),
				Height: uint16(msg.Height),
			}
		}
	case tea.KeyMsg:
		if key := msg.String(); key == "ctrl+c" || key == "ctrl+d" || ui.error != nil {
			return ui, tea.Quit
		}
	case state.StateChangedMsg:
		log.Debug("Change in state", "state", msg.State)
		ui.common.State = msg.State
		switch msg.State {
		case state.FetchingPod:
			ui.activeView = loadingView
			cmds = append(cmds, ui.common.Actions.FetchPod(ui.common.User))
		case state.FetchingPVC:
			cmds = append(cmds, ui.common.Actions.FetchPVC(ui.common.User))
		case state.CreatingPVC:
			ui.createdPVC = true
			cmds = append(cmds, ui.common.Actions.CreatePVC(ui.common.User, config.Values.PVCSize))
		case state.WaitingForPVC:
			cmds = append(cmds, ui.common.Actions.WaitForPVC(msg.PVC))
		case state.CreatingPod:
			if ui.createdPVC {
				cmds = append(cmds, ui.common.Actions.CreateInitialPod(ui.common.User, config.Values.ContainerImage, msg.PVC))
			} else {
				cmds = append(cmds, ui.common.Actions.CreatePod(ui.common.User, config.Values.ContainerImage, msg.PVC))
			}
		case state.WaitingForPod:
			cmds = append(cmds, ui.common.Actions.WaitForPodInitContainer(msg.Pod))
		case state.WaitingForInitContainer:
			ui.activeView = tailView
			cmds = append(cmds, actions.StartLogTail(msg.Pod))
		case state.PodRunning:
			cmds = append(cmds, ui.common.Actions.AttachToPod(msg.Pod, ui.common.User, ui.sizeChan))
			ui.sizeChan <- remotecommand.TerminalSize{
				Width:  uint16(ui.common.Width),
				Height: uint16(ui.common.Height),
			}
		case state.PodTerminated:
			ui.activeView = completedView
		case state.Error:
			ui.error = msg.Error
		}
	}
	for i, v := range ui.views {
		m, cmd := v.Update(msg)
		ui.views[i] = m
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return ui, tea.Batch(cmds...)
}

// View implements tea.Model.
func (ui *UI) View() string {
	if ui.error != nil {
		return ui.common.RenderCentered(
			fmt.Sprintf(
				"%s\n\n%s\n%s\n%s",
				common.LogoSprite[0],
				common.ErrorStyle.Render("ERROR"),
				ui.error.Error(),
				"Press any key to exit",
			),
		)
	}
	if ui.common.State == state.Unknown {
		return fmt.Sprintf("%s\n%s", common.Splash, "Loading...")
	}
	return ui.views[ui.activeView].View()
}
