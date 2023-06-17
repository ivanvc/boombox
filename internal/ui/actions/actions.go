package actions

import k8s "github.com/ivanvc/boombox/internal/services/kubernetes"

// Actions is the bridge between the UI and the services.
type Actions struct {
	k8sClient *k8s.Client
}

// Returns a new Actions instance.
func New(client *k8s.Client) *Actions {
	return &Actions{client}
}
