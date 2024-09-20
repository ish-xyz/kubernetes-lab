package executor

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/config"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	helmkube "helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/repo"
)

func (e *Executor) helmDownload(url, chartName, chartVersion, outFilePath string) error {

	logrus.Infoln("downloading chart ", chartName, chartVersion)
	r, err := repo.NewChartRepository(&repo.Entry{
		URL: url,
	}, getter.All(cli.New()))
	if err != nil {
		return fmt.Errorf("Error creating chart repository: %s\n", err)
	}

	index, err := r.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("Error downloading index file: %s\n", err)
	}

	indexFile, err := repo.LoadIndexFile(index)
	if err != nil {
		return fmt.Errorf("Error loading index file: %s\n", err)
	}

	cv, err := indexFile.Get(chartName, chartVersion)
	if err != nil {
		return fmt.Errorf("Error finding chart version: %s\n", err)
	}

	// Download the chart
	get, err := getter.All(cli.New()).ByScheme("https")
	if err != nil {
		return fmt.Errorf("failed to get getter by scheme 'https': %v", err)
	}

	chartUrl := cv.URLs[0]
	if !strings.HasPrefix(chartUrl, "http://") &&
		!strings.HasPrefix(chartUrl, "https://") {
		chartUrl = fmt.Sprintf(
			"%s/%s", strings.TrimRight(url, "/"),
			strings.TrimLeft(chartUrl, "/"),
		)
	}

	chartBytes, err := get.Get(chartUrl)
	if err != nil {
		return fmt.Errorf("Error downloading chart: %s\n", err)
	}

	outFile, err := os.Create(outFilePath)
	if err != nil {
		return fmt.Errorf("Error creating output file: %v\n", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, chartBytes)

	return err
}

func (e *Executor) HelmInstall(chart *config.ChartConfig, kubeconfigPath string) error {

	outFilePath := fmt.Sprintf("%s/%s-%s.tgz", e.TempFolder, chart.Name, chart.Version)
	err := e.helmDownload(chart.Url, chart.Name, chart.Version, outFilePath)
	if err != nil {
		return fmt.Errorf("helm download error => %v", err)
	}

	logrus.Infof("loading chart %s-%s ...", chart.Name, chart.Version)
	chartObj, err := loader.Load(outFilePath)
	if err != nil {
		return fmt.Errorf("helm load error => %v", err)
	}

	actionConfig := new(action.Configuration)
	err = actionConfig.Init(
		helmkube.GetConfig(kubeconfigPath, "", chart.Namespace),
		chart.Namespace,
		os.Getenv("HELM_DRIVER"),
		func(format string, v ...interface{}) {
			logrus.Infof(format, v)
		},
	)
	if err != nil {
		return fmt.Errorf("helm init error => %v", err)
	}

	logrus.Infof("installing chart %s-%s ...", chart.Name, chart.Version)
	iCli := action.NewInstall(actionConfig)
	iCli.Namespace = chart.Namespace
	iCli.ReleaseName = chart.ReleaseName
	iCli.IsUpgrade = true
	_, err = iCli.Run(chartObj, nil)
	if err != nil {
		return fmt.Errorf("helm install error => %v", err)
	}
	return nil
}
