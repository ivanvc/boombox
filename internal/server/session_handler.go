package server

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	bm "github.com/charmbracelet/wish/bubbletea"

	"github.com/ivanvc/boombox/internal/config"
	k8s "github.com/ivanvc/boombox/internal/services/kubernetes"
	"github.com/ivanvc/boombox/internal/ui"
	"github.com/ivanvc/boombox/internal/ui/actions"
	"github.com/ivanvc/boombox/internal/ui/common"
)

func sessionHandler(server *Server, cfg *config.Config, client *k8s.Client) bm.ProgramHandler {
	return func(sess ssh.Session) *tea.Program {
		pty, _, active := sess.Pty()
		if !active {
			log.Error("No active terminal", "session", sess)
			return nil
		}

		server.RegisterSession()
		common := &common.Common{
			Session: sess,
			User:    sess.User(),
			Width:   pty.Window.Width,
			Height:  pty.Window.Height,
			Client:  client,
			Config:  cfg,
			Actions: actions.New(client),
		}

		ctx := log.WithContext(sess.Context(), log.Default())
		p := tea.NewProgram(ui.New(common),
			tea.WithInput(sess),
			tea.WithOutput(sess),
			tea.WithAltScreen(),
			tea.WithContext(ctx),
		)

		go func() {
			defer server.DeregisterSession()
			<-ctx.Done()
			pod, err := client.GetPod(sess.User())
			if pod == nil || err != nil {
				return
			}
			if server.IsShuttingDown() {
				client.DeletePod(pod)
			}
			count, _ := client.GetActivePTYs(pod)
			log.Debug("Active PTYs", "pod", pod.Name, "count", count)
			if count == 1 {
				client.DeletePod(pod)
			}
		}()

		return p
	}
}
