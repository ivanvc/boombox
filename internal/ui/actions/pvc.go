package actions

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"

	"github.com/ivanvc/boombox/internal/ui/common/state"
)

// FetchPVC tries to see if there's a Persitent Volume Claim with that name in the cluster.
func (a *Actions) FetchPVC(name string) tea.Cmd {
	return func() tea.Msg {
		pvc, err := a.k8sClient.GetPVC(name)
		if err != nil {
			log.Error("Error fetching PVC", err)
			return state.StateChangedMsg{
				State: state.Error,
				Error: err,
			}
		}
		if pvc == nil {
			return state.StateChangedMsg{State: state.CreatingPVC}
		}
		return state.StateChangedMsg{
			State: state.CreatingPod,
			PVC:   pvc,
		}
	}
}

// CreatePVC creates a new PersistentVolumeClaim with a given name and size in the cluster.
func (a *Actions) CreatePVC(name, size string) tea.Cmd {
	return func() tea.Msg {
		pvc, err := a.k8sClient.CreatePVC(name, size)
		if err != nil {
			log.Error("Error creating PVC", err)
			return state.StateChangedMsg{
				State: state.Error,
				Error: err,
			}
		}
		return state.StateChangedMsg{
			State: state.WaitingForPVC,
			PVC:   pvc,
		}
	}
}

// WaitForPVC waits until the PersistentVolumeClaim is in a ready state.
func (a *Actions) WaitForPVC(pvc *corev1.PersistentVolumeClaim) tea.Cmd {
	return func() tea.Msg {
		if err := a.k8sClient.WaitForPVC(pvc); err != nil {
			log.Error("Error waiting for PVC", err)
			return state.StateChangedMsg{
				State: state.Error,
				Error: err,
			}
		}
		return state.StateChangedMsg{
			State: state.CreatingPod,
			PVC:   pvc,
		}
	}
}
