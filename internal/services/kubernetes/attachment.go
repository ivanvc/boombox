package kubernetes

import (
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

// Attachment represents the console (stdin/stdout) attachment to a given Pod.
type Attachment struct {
	*Client

	user string
	pod  *corev1.Pod

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	sizeChan SizeChan
}

// Returns a new Attachment.
func (c *Client) NewAttachment(pod *corev1.Pod, user string, sizeChan SizeChan) *Attachment {
	return &Attachment{Client: c, pod: pod, user: user, sizeChan: sizeChan}
}

// SetStdin implements tea.ExecCommand.
func (a *Attachment) SetStdin(reader io.Reader) {
	a.stdin = reader
}

// SetStdout implements tea.ExecCommand.
func (a *Attachment) SetStdout(writer io.Writer) {
	a.stdout = writer
}

// SetStderr implements tea.ExecCommand.
func (a *Attachment) SetStderr(writer io.Writer) {
	a.stderr = writer
}

// Run implements tea.ExecCommand.
func (a *Attachment) Run() error {
	execOpts := &corev1.PodExecOptions{
		Container: a.pod.Spec.Containers[0].Name,
		Command:   []string{"su", "-", a.user},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}

	req := a.CoreV1().RESTClient().Post().
		Namespace(a.pod.Namespace).
		Resource("pods").
		Name(a.pod.Name).
		SubResource("exec").
		VersionedParams(execOpts, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(
		a.GetConfig(),
		http.MethodPost,
		req.URL(),
	)
	if err != nil {
		return err
	}

	// Use stdout as stderr, because Bubble Tea assigns os.Stderr when calling
	// ExecCommand.SetStderr(io.Writer), which would then show the stderr output
	// on the server's screen rather than the client's.
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             a.stdin,
		Stdout:            a.stdout,
		Stderr:            a.stdout,
		Tty:               true,
		TerminalSizeQueue: a.sizeChan,
	})

	if err != nil {
		return err
	}

	return nil
}

type SizeChan chan remotecommand.TerminalSize

func (s SizeChan) Next() *remotecommand.TerminalSize {
	size, ok := <-s
	if !ok {
		return nil
	}
	return &size
}
