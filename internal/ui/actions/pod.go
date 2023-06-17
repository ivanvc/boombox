package actions

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	k8s "github.com/ivanvc/boombox/internal/services/kubernetes"
	"github.com/ivanvc/boombox/internal/ui/common/state"
	corev1 "k8s.io/api/core/v1"
)

// FetchPod tries to see if there's a Pod with that name in the cluster.
func (a *Actions) FetchPod(name string) tea.Cmd {
	return func() tea.Msg {
		pod, err := a.k8sClient.GetPod(name)
		if err != nil {
			log.Error("Error fetching pod", err)
			return state.StateChangedMsg{
				State: state.Error,
				Error: err,
			}
		}
		if pod == nil {
			return state.StateChangedMsg{State: state.FetchingPVC}
		}
		if pod.Status.Phase == corev1.PodRunning {
			return state.StateChangedMsg{
				State: state.PodRunning,
				Pod:   pod,
			}
		}
		return state.StateChangedMsg{
			State: state.WaitingForPod,
			Pod:   pod,
		}
	}
}

// CreateInitialPod creates a new Pod with the init container that provisions the user home.
func (a *Actions) CreateInitialPod(name, image string, pvc *corev1.PersistentVolumeClaim) tea.Cmd {
	return func() tea.Msg {
		pod, err := a.k8sClient.CreateInitialPod(name, image, pvc)
		if err != nil {
			log.Error("Error creating pod", err)
			return state.StateChangedMsg{
				State: state.Error,
				Error: err,
			}
		}
		return state.StateChangedMsg{
			State: state.WaitingForPod,
			Pod:   pod,
		}
	}
}

// CreatePod creates a new Pod with the given name, container image, and pvc.
func (a *Actions) CreatePod(name, image string, pvc *corev1.PersistentVolumeClaim) tea.Cmd {
	return func() tea.Msg {
		pod, err := a.k8sClient.CreatePod(name, image, pvc)
		if err != nil {
			log.Error("Error creating pod", err)
			return state.StateChangedMsg{
				State: state.Error,
				Error: err,
			}
		}
		return state.StateChangedMsg{
			State: state.WaitingForPod,
			Pod:   pod,
		}
	}
}

// WaitForPodInitContainer waits until the pod is ready.
func (a *Actions) WaitForPodInitContainer(pod *corev1.Pod) tea.Cmd {
	return func() tea.Msg {
		status, err := a.k8sClient.WaitForPodInitContainer(pod)
		if err != nil {
			log.Error("Error waiting for pod", err)
			return state.StateChangedMsg{
				State: state.Error,
				Error: err,
			}
		}
		if status == k8s.PodStatusInitContainerReady {
			return state.StateChangedMsg{
				State: state.WaitingForInitContainer,
				Pod:   pod,
			}
		}
		if status == k8s.PodStatusReady {
			return state.StateChangedMsg{
				State: state.PodRunning,
				Pod:   pod,
			}
		}
		return nil
	}
}

// Tail Pod's initContainer logs
func (a *Actions) TailInitContainerLogs(pod *corev1.Pod, linesChan chan string) tea.Cmd {
	return func() tea.Msg {
		eof := make(chan error)
		lt := a.k8sClient.NewLogTail(pod, linesChan)
		go func() {
			eof <- lt.Run("init")
		}()

		if err := <-eof; err != nil {
			log.Error("Error tailing log lines", err)
			return state.StateChangedMsg{
				State: state.Error,
				Error: err,
			}
		}

		return state.StateChangedMsg{
			State: state.WaitingForPod,
			Pod:   pod,
		}
	}
}

// Attach to a running Pod.
func (a *Actions) AttachToPod(pod *corev1.Pod, user string, sizeChan k8s.SizeChan) tea.Cmd {
	attachment := a.k8sClient.NewAttachment(pod, user, sizeChan)
	return tea.Exec(attachment, func(err error) tea.Msg {
		if conn, err := a.k8sClient.GetActivePTYs(pod); err != nil {
			return state.StateChangedMsg{
				State: state.Error,
				Error: err,
			}
		} else if conn == 1 {
			log.Debugf("Last PTY from Pod %q, deleting", pod.Name)
			if err := a.k8sClient.DeletePod(pod); err != nil {
				log.Errorf("Error deleting pod %T", err)
				return state.StateChangedMsg{
					State: state.Error,
					Error: err,
				}
			}
		}
		return state.StateChangedMsg{
			State: state.PodTerminated,
			Pod:   pod,
		}
	})
}
