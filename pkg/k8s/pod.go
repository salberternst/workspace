package k8s

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/salberternst/workspace/pkg/utils"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

func WatchPodEvents(name string, namespace string) (watch.Interface, error) {
	pods, err := client.CoreV1.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("workspace-name=%s", name),
	})

	if err != nil {
		return nil, err
	}

	if len(pods.Items) > 1 {
		return nil, fmt.Errorf("Multiple pods found for workspace")
	}

	watcher, err := client.CoreV1.CoreV1().Events(namespace).Watch(context.TODO(),
		metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s", pods.Items[0].Name),
			TypeMeta: metav1.TypeMeta{
				Kind: "Pod",
			},
		})

	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	go func() {
		for event := range watcher.ResultChan() {
			if event.Object == nil {
				return
			}

			event, ok := event.Object.(*apiv1.Event)
			if !ok {
				continue
			}

			if event.LastTimestamp.After(startTime) {
				fmt.Printf("%s %s %s %s\n",
					event.Type,
					event.Reason,
					event.LastTimestamp.Time.Format(time.RFC1123Z),
					event.Message,
				)
			}
		}
	}()

	return watcher, nil
}

func GetWorkspacePod(namespace string, workspacName string) (*v1.Pod, error) {
	pods, err := GetClient().CoreV1.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "workspace-name=" + workspacName,
	})

	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, nil
	}

	return &pods.Items[0], nil
}

func GetPodLogs(pod v1.Pod, container string, follow bool) error {
	podLogOpts := v1.PodLogOptions{
		Follow:    follow,
		Container: container,
	}

	stream, err := GetClient().CoreV1.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts).Stream(context.TODO())
	if err != nil {
		return err
	}

	defer stream.Close()

	r := bufio.NewReader(stream)
	for {
		bytes, err := r.ReadBytes('\n')
		if _, err := os.Stdout.Write(bytes); err != nil {
			return err
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}

func ExecuteInPod(namespace string, name string, container string, command []string, terminal bool) error {
	req := GetClient().CoreV1.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       terminal,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(GetClient().Config, http.MethodPost, req.URL())
	if err != nil {
		return err
	}

	var sizeQueue remotecommand.TerminalSizeQueue
	if terminal {
		terminal, err := utils.NewTerminal()
		if err != nil {
			return err
		}

		sizeQueue = terminal.SizeQueue

		terminal.MonitorSize()

		defer terminal.Close()
	}

	if err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
		Tty:               terminal,
		TerminalSizeQueue: sizeQueue,
	}); err != nil {
		return err
	}

	return nil
}
