package kubernetes

import (
	"bufio"
	"context"

	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
)

// LogTail is the struct that tails the logs for a Pod's container.
type LogTail struct {
	*Client
	pod       *corev1.Pod
	linesChan chan string
}

// NewLogTail returns a new LogTail instance.
func (c *Client) NewLogTail(pod *corev1.Pod, linesChan chan string) *LogTail {
	return &LogTail{c, pod, linesChan}
}

// Run follows the log lines, and sends them to the linesChan.
func (lt *LogTail) Run(container string) error {
	count := int64(100)
	podLogOptions := corev1.PodLogOptions{
		Container: container,
		Follow:    true,
		TailLines: &count,
	}

	podLogRequest := lt.CoreV1().Pods(lt.pod.Namespace).
		GetLogs(lt.pod.Name, &podLogOptions)
	stream, err := podLogRequest.Stream(context.Background())
	if err != nil {
		return err
	}
	defer stream.Close()

	log.Debug("Tailing")
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		t := scanner.Text()
		log.Debugf("Got line: %q", t)
		lt.linesChan <- t
	}
	if err := scanner.Err(); err != nil {
		log.Error("Error reading standard input", "error", err)
		return err
	}
	log.Debug("End tailing")

	return nil
}
