package kubernetes

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

// PodStatus represents if the Pod init container is ready, or if the Pod is ready.
type PodStatus int

const (
	PodStatusUnknown PodStatus = iota
	PodStatusInitContainerReady
	PodStatusReady
)

// Client holds a wrapped Kubernetes client.
type Client struct {
	*k8s.Clientset
	config    *rest.Config
	namespace string
}

// LoadClient creates a new Client singleton.
func LoadClient(namespace string) *Client {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		log.Fatal("Error loading kubeconfig", "error", err)
	}
	cs, err := k8s.NewForConfig(cfg)
	if err != nil {
		log.Fatal("Error initializing Kubernetes client", "error", err)
	}
	return &Client{cs, cfg, namespace}
}

// Gets client Config.
func (c *Client) GetConfig() *rest.Config {
	return c.config
}

// Get a Pod by name from the cluster.
func (c *Client) GetPod(name string) (*corev1.Pod, error) {
	pod, err := c.CoreV1().Pods(c.namespace).Get(
		context.Background(),
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return pod, nil
}

// Get a PVC by name from the cluster.
func (c *Client) GetPVC(name string) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := c.CoreV1().PersistentVolumeClaims(c.namespace).Get(
		context.Background(),
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return pvc, nil
}

// Creates a PVC by name with a given size.
func (c *Client) CreatePVC(name, size string) (*corev1.PersistentVolumeClaim, error) {
	pvc := getPVCPayload(c.namespace, name, size)
	if _, err := c.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(
		context.Background(),
		pvc,
		metav1.CreateOptions{},
	); err != nil {
		return nil, err
	}
	return pvc, nil
}

// Wait for a PersistentVolumeClaim to be ready.
func (c *Client) WaitForPVC(pvc *corev1.PersistentVolumeClaim) error {
	w, err := c.CoreV1().PersistentVolumeClaims(pvc.Namespace).Watch(
		context.Background(),
		metav1.SingleObject(pvc.ObjectMeta),
	)
	if err != nil {
		return err
	}

	for event := range w.ResultChan() {
		switch event.Type {
		case watch.Modified:
			w.Stop()
		case watch.Added:
			w.Stop()
		}
	}
	return nil
}

// Creates a Pod in the cluster with a given name, container image, and a pvc that will be mounted on /home.
func (c *Client) CreatePod(name, image string, pvc *corev1.PersistentVolumeClaim) (*corev1.Pod, error) {
	pod := getPodPayload(c.namespace, name, image, pvc)
	if err := c.createPod(pod); err != nil {
		return nil, err
	}
	return pod, nil
}

// Creates a Pod with an init container that provisions the user home, in the cluster with a given name, container image, and a pvc that will be mounted on /home.
func (c *Client) CreateInitialPod(name, image string, pvc *corev1.PersistentVolumeClaim) (*corev1.Pod, error) {
	pod := getInitialPodPayload(c.namespace, name, image, pvc)
	if err := c.createPod(pod); err != nil {
		return nil, err
	}
	return pod, nil
}

func (c *Client) createPod(pod *corev1.Pod) error {
	if _, err := c.CoreV1().Pods(pod.Namespace).Create(
		context.Background(),
		pod,
		metav1.CreateOptions{},
	); err != nil {
		return err
	}

	return nil
}

// Waits for a Pod to be scheduled.
func (c *Client) WaitForPodInitContainer(pod *corev1.Pod) (PodStatus, error) {
	w, err := c.CoreV1().Pods(pod.Namespace).Watch(
		context.Background(),
		metav1.SingleObject(pod.ObjectMeta),
	)
	if err != nil {
		return PodStatusUnknown, err
	}

	for event := range w.ResultChan() {
		switch event.Type {
		case watch.Modified:
			pod = event.Object.(*corev1.Pod)
			for _, status := range pod.Status.InitContainerStatuses {
				if status.Name == "init" && status.State.Running != nil {
					w.Stop()
					return PodStatusInitContainerReady, nil
				}
			}
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady &&
					cond.Status == corev1.ConditionTrue {
					w.Stop()
					return PodStatusReady, nil
				}
			}
		}
	}
	return PodStatusUnknown, nil
}

// Deletes a Pod.
func (c *Client) DeletePod(pod *corev1.Pod) error {
	err := c.CoreV1().Pods(pod.Namespace).Delete(
		context.Background(),
		pod.Name,
		metav1.DeleteOptions{},
	)
	return err
}

// Query for active PTYs in a Pod.
func (c *Client) GetActivePTYs(pod *corev1.Pod) (int, error) {
	stdout, err := c.execCommandInPod(pod, "/bin/sh", "-c", "find /dev/pts -group tty | wc -l")
	if err != nil {
		return -1, err
	}

	conn, err := strconv.Atoi(strings.TrimSpace(stdout))
	if err != nil {
		return -1, err
	}

	return conn, nil
}

// Send a wall about the shutdown to all session in the Pod.
func (c *Client) SendShutdownWallToPod(pod *corev1.Pod) error {
	_, err := c.execCommandInPod(pod, "/bin/sh", "-c", "for pty in $(find /dev/pts -group tty); do echo -e '\n###########################\n The system is going down! \n\n' > $pty; done")
	return err
}

func (c *Client) execCommandInPod(pod *corev1.Pod, command ...string) (string, error) {
	execOpts := &corev1.PodExecOptions{
		Container: pod.Spec.Containers[0].Name,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    false,
		TTY:       false,
	}
	req := c.CoreV1().RESTClient().Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(execOpts, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, http.MethodPost, req.URL())
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: nil,
		Tty:    false,
	})

	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}
