package helm

import (
	"io/fs"

	"github.com/salberternst/workspace/pkg/charts"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func GetFilesForChart(chart fs.FS, chartName string) ([]string, error) {
	fnames := []string{}
	err := fs.WalkDir(chart, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fnames = append(fnames, path)
		return nil
	})

	return fnames, err
}

func LoadChart(chartName string) (*chart.Chart, error) {
	chart, err := fs.Sub(charts.EmbeddedCharts, chartName)
	if err != nil {
		return nil, err
	}

	fnames, err := GetFilesForChart(chart, chartName)
	if err != nil {
		return nil, err
	}

	var files []*loader.BufferedFile
	for _, fname := range fnames {
		data, err := fs.ReadFile(chart, fname)
		if err != nil {
			return nil, err
		}

		// Helm expects unix / separator, but on windows this will be \
		files = append(files, &loader.BufferedFile{
			Name: fname,
			Data: data,
		})
	}

	return loader.LoadFiles(files)
}

func ReleaseExists(namespace string, name string) error {
	helmConfiguration, err := GetConfiguration(namespace)
	if err != nil {
		return err
	}

	getAction := action.NewGet(helmConfiguration)

	_, err = getAction.Run(name)
	return err
}
