package common

import (
	"github.com/charmbracelet/ssh"
	k8s "github.com/ivanvc/boombox/internal/services/kubernetes"

	"github.com/ivanvc/boombox/internal/config"
	"github.com/ivanvc/boombox/internal/ui/actions"
	"github.com/ivanvc/boombox/internal/ui/common/state"
)

// Common holds elements used by all the UI components and views.
type Common struct {
	Session ssh.Session
	User    string

	Width  int
	Height int

	Client  *k8s.Client
	Config  *config.Config
	Actions *actions.Actions
	State   state.State
}
