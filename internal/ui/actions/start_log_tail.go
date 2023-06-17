package actions

import (
	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"
)

// TriggerStartLogTail tells the tail view to start getting the logs.
type TriggerStartLogTailMsg struct {
	Pod *corev1.Pod
}

// Starts the process to tail initContainer logs
func StartLogTail(pod *corev1.Pod) tea.Cmd {
	return func() tea.Msg {
		return TriggerStartLogTailMsg{pod}
	}
}
