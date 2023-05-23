package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

type Chart struct {
	chart     *chart.Chart
	chartName string
}

func NewChart(chartName string) (Chart, error) {
	chart, err := LoadChart(chartName)
	if err != nil {
		return Chart{}, err
	}

	return Chart{
		chart:     chart,
		chartName: chartName,
	}, nil
}

func (o *Chart) Install(namespace string, releaseName string, dryRun bool, values map[string]interface{}) (*release.Release, error) {
	helmConfiguration, err := GetConfiguration(namespace)
	if err != nil {
		return nil, err
	}

	installAction := action.NewInstall(helmConfiguration)
	installAction.Namespace = namespace
	installAction.ReleaseName = releaseName
	installAction.DryRun = dryRun

	return installAction.Run(o.chart, values)
}

func (o *Chart) Update(namespace string, releaseName string, dryRun bool, values map[string]interface{}) (*release.Release, error) {
	helmConfiguration, err := GetConfiguration(namespace)
	if err != nil {
		return nil, err
	}

	upgradeAction := action.NewUpgrade(helmConfiguration)
	upgradeAction.Namespace = namespace
	upgradeAction.DryRun = dryRun
	upgradeAction.ReuseValues = true

	return upgradeAction.Run(releaseName, o.chart, values)
}

func (o *Chart) Delete(namespace string, releaseName string, dryRun bool) (*release.UninstallReleaseResponse, error) {
	helmConfiguration, err := GetConfiguration(namespace)
	if err != nil {
		return nil, err
	}

	uninstallAction := action.NewUninstall(helmConfiguration)
	uninstallAction.DryRun = false

	return uninstallAction.Run(releaseName)
}

func (o *Chart) List(namespace string) ([]*release.Release, error) {
	helmConfiguration, err := GetConfiguration(namespace)
	if err != nil {
		return nil, err
	}

	listAction := action.NewList(helmConfiguration)

	return listAction.Run()
}

func (o *Chart) Get(namespace string, name string) error {
	helmConfiguration, err := GetConfiguration(namespace)
	if err != nil {
		return err
	}

	getAction := action.NewGet(helmConfiguration)

	release, err := getAction.Run(name)
	if err != nil {
		return err
	}

	if release.Chart.Metadata.Name != o.chartName {
		return fmt.Errorf("%s in project %s is not a %s chart", name, namespace, o.chartName)
	}

	return nil
}
