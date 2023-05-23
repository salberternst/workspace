package utils

import (
	"os"
	"path/filepath"
)

const PrivateKeyFileName = "id_devspace_ecdsa"

func EnsureConfigFolder(name string, namespace string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homedir, ".workspace", namespace, name)

	err = os.MkdirAll(configPath, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func WritePrivateKey(name string, namespace string, privateKey []byte) (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	err = EnsureConfigFolder(name, namespace)
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(homedir, ".workspace", namespace, name, PrivateKeyFileName)
	return configPath, os.WriteFile(configPath, privateKey, 0600)
}
