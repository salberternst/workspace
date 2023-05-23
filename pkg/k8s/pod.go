package k8s

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"time"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

func WaitUntilPodIsReady(name, namespace string) error {
	client := GetClient()

	watcher, err := client.CoreV1.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
	})

	if err != nil {
		return err
	}

	defer watcher.Stop()

	// TODO: check also for errors to return immediatlty
	ready := make(chan bool, 1)
	go func() {
		for event := range watcher.ResultChan() {
			if event.Object == nil {
				return
			}

			pod, ok := event.Object.(*apiv1.Pod)
			if !ok {
				continue
			}

			if len(pod.Status.ContainerStatuses) > 0 && pod.Status.ContainerStatuses[0].Ready {
				ready <- true
			}
		}
	}()

	select {
	case <-ready:
		return nil
	case <-time.After(30 * time.Second):
		return errors.New("timeout occured")
	}
}

func GetNotebookPod(namespace string, notebookName string) (*v1.Pod, error) {
	pods, err := GetClient().CoreV1.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "notebook-name=" + notebookName,
	})

	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, nil
	}

	return &pods.Items[0], nil
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
