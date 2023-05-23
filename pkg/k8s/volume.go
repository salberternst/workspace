package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetWorkspaceVolumes(namespace string, notebookName string) ([]v1.PersistentVolumeClaim, error) {
	volumes, err := GetClient().CoreV1.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "workspace/name=" + notebookName,
	})

	if err != nil {
		return nil, err
	}

	if len(volumes.Items) == 0 {
		return nil, nil
	}

	return volumes.Items, nil
}
