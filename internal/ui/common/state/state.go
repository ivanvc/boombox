package state

import corev1 "k8s.io/api/core/v1"

// State holds the current state of the UI.
type State int

const (
	Unknown State = iota
	FetchingPod
	FetchingPVC
	CreatingPVC
	WaitingForPVC
	CreatingPod
	WaitingForPod
	WaitingForInitContainer
	PodTerminated
	PodRunning
	AttachedToPod
	Error
)

// StateChangedMsg is the message sent when there's a change in the UI State.
type StateChangedMsg struct {
	State State
	Pod   *corev1.Pod
	PVC   *corev1.PersistentVolumeClaim
	Error error
}

func (s State) String() string {
	switch s {
	case FetchingPod:
		return "Communicating to the Kubernetes cluster"
	case FetchingPVC:
		return "Fetching volume"
	case CreatingPVC:
		return "Creating volume"
	case WaitingForPVC:
		return "Waiting for volume to be ready"
	case CreatingPod:
		return "Creating pod"
	case WaitingForPod:
		return "Waiting for pod to be ready"
	case PodRunning:
		return "Attaching to pod..."
	}
	return ""
}
