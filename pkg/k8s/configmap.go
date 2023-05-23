package k8s

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FileMap struct {
	LocalFilename string
	Id            string
	MountPath     string
}

type ConfigMapper struct {
	files []FileMap
}

func NewConfigMapper(files []FileMap) ConfigMapper {
	return ConfigMapper{
		files: files,
	}
}

func (o *ConfigMapper) BuildConfigMap(namespace, name string) error {
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	configMap.Name = name
	configMap.Data = map[string]string{}
	configMap.BinaryData = map[string][]byte{}

	for _, file := range o.files {
		data, err := os.ReadFile(file.LocalFilename)
		if err != nil {
			return err
		}
		configMap.BinaryData[file.Id] = data

	}

	return nil
}
