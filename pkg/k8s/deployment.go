package k8s

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

func WaitForDeployment(name, namespace string, waitTimeout uint) error {
	watcher, err := GetClient().CoreV1.AppsV1().Deployments(namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
	})

	if err != nil {
		return err
	}

	defer watcher.Stop()

	ready := make(chan bool, 1)
	go func() {
		for event := range watcher.ResultChan() {
			if event.Object == nil {
				return
			}

			deployment, ok := event.Object.(*v1.Deployment)
			if !ok {
				continue
			}
			if deployment.Status.Replicas > 0 && deployment.Status.Replicas == deployment.Status.ReadyReplicas {
				ready <- true
			}
		}
	}()

	select {
	case <-ready:
		return nil
	case <-time.After(time.Duration(waitTimeout) * time.Second):
		return fmt.Errorf("Timeout occured after %d seconds while waiting for deployment to become ready", waitTimeout)
	}
}
