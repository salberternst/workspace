package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"text/template"
)

const HostConfigTemplate = `# workspace start {{.Name}}.{{.Namespace}}.workspace
Host {{.Name}}.{{.Namespace}}.workspace
  HostName {{.Hostname}}
  LogLevel error
  Port {{.Port}}
  IdentityFile "{{.PrivateKeyPath}}"
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  User workspace
# workspace end {{.Name}}.{{.Namespace}}.workspace`
const HostConfigRegex = `(?s)(# workspace start {{.Name}}.{{.Namespace}}.workspace)(.*)(# workspace end {{.Name}}.{{.Namespace}}.workspace)`

type HostConfig struct {
	Name           string
	Namespace      string
	Hostname       string
	Port           uint16
	PrivateKeyPath string
	template       *template.Template
}

func NewHostConfig(name string, namespace string, privateKeyPath string) (string, error) {
	template, err := template.New("ssh_conf").Parse(HostConfigTemplate)
	if err != nil {
		return "", err
	}

	hostConfig := &HostConfig{
		Name:           name,
		Namespace:      namespace,
		template:       template,
		Port:           2222,
		Hostname:       "localhost",
		PrivateKeyPath: privateKeyPath,
	}

	return hostConfig.serialize()
}

func (o *HostConfig) serialize() (string, error) {
	var data bytes.Buffer
	if err := o.template.Execute(&data, o); err != nil {
		return "", err
	}

	return data.String(), nil
}

func buildRegex(name string, namespace string) (string, error) {
	template, err := template.New("host-regex").Parse(HostConfigRegex)
	if err != nil {
		return "", err
	}

	var data bytes.Buffer
	if err := template.Execute(&data, struct {
		Name      string
		Namespace string
	}{
		Name:      name,
		Namespace: namespace,
	}); err != nil {
		return "", err
	}

	return data.String(), nil
}

func DeleteSshConfEntry(name string, namespace string) error {
	sshConfPath := filepath.Join(os.Getenv("HOME"), ".ssh", "config")

	file, err := os.ReadFile(sshConfPath)
	if err != nil {
		return err
	}

	expression, err := buildRegex(name, namespace)
	if err != nil {
		return err
	}

	regex, err := regexp.Compile(expression)
	if err != nil {
		return err
	}

	err = os.WriteFile(sshConfPath, []byte(regex.ReplaceAllString(string(file), "")), 0644)
	if err != nil {
		return err
	}

	return nil
}

func AppendSshConfEntry(name string, namespace string, privateKeyPath string) error {
	file, err := os.OpenFile(filepath.Join(os.Getenv("HOME"), ".ssh", "config"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer file.Close()

	hostConfig, err := NewHostConfig(name, namespace, privateKeyPath)
	if err != nil {
		return err
	}

	if _, err := file.WriteString(hostConfig); err != nil {
		return err
	}

	return nil
}
