package k8s

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

func GetStatefulSet(name, namespace string) (*v1.StatefulSet, error) {
	return GetClient().CoreV1.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func WaitForStatefulSetReplica(name, namespace string, waitTimeout uint) error {
	watcher, err := GetClient().CoreV1.AppsV1().StatefulSets(namespace).Watch(context.TODO(), metav1.ListOptions{
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

			statefulSet, ok := event.Object.(*v1.StatefulSet)
			if !ok {
				continue
			}

			if statefulSet.Status.Replicas > 0 {
				ready <- true
			}
		}
	}()

	select {
	case <-ready:
		return nil
	case <-time.After(time.Duration(waitTimeout) * time.Second):
		return fmt.Errorf("Timeout occured after %d seconds while waiting for the workspace to become ready", waitTimeout)
	}
}

func WaitForStatefulSetReplicaReady(name, namespace string, waitTimeout uint) error {
	watcher, err := GetClient().CoreV1.AppsV1().StatefulSets(namespace).Watch(context.TODO(), metav1.ListOptions{
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

			statefulSet, ok := event.Object.(*v1.StatefulSet)
			if !ok {
				continue
			}

			if statefulSet.Status.ReadyReplicas > 0 {
				ready <- true
			}
		}
	}()

	select {
	case <-ready:
		return nil
	case <-time.After(time.Duration(waitTimeout) * time.Second):
		return fmt.Errorf("Timeout occured after %d seconds while waiting for the workspace to become ready", waitTimeout)
	}
}
