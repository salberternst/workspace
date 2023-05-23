package helm

import (
	"log"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type HelmClient struct {
	Install   *action.Install
	Uninstall *action.Uninstall
}

const (
	HelmDriver = "secret"
)

func GetConfiguration(namespace string) (*action.Configuration, error) {
	settings := cli.New()
	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, HelmDriver, log.Printf); err != nil {
		return nil, err
	}

	return actionConfig, nil
}
