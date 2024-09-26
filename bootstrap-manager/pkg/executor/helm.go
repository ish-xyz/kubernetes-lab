package executor

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
		return fmt.Errorf("error creating chart repository: %s", err)
	}

	index, err := r.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("error downloading index file: %v", err)
	}

	indexFile, err := repo.LoadIndexFile(index)
	if err != nil {
		return fmt.Errorf("error loading index file: %v", err)
	}

	cv, err := indexFile.Get(chartName, chartVersion)
	if err != nil {
		return fmt.Errorf("error finding chart version: %v", err)
	}

	// prepare for download
	var get getter.Getter
	if !strings.HasPrefix(url, "http://") {
		get, err = getter.All(cli.New()).ByScheme("http")
		if err != nil {
			return fmt.Errorf("failed to get getter by scheme 'http': %v", err)
		}
	} else if strings.HasPrefix(url, "https://") {
		get, err = getter.All(cli.New()).ByScheme("https")
		if err != nil {
			return fmt.Errorf("failed to get getter by scheme 'https': %v", err)
		}
	} else {
		return fmt.Errorf("unsupported protocol in URL '%s'", url)
	}

	chartAddress := cv.URLs[0]
	if !strings.HasPrefix(chartAddress, "http://") &&
		!strings.HasPrefix(chartAddress, "https://") {
		chartAddress = fmt.Sprintf(
			"%s/%s", strings.TrimRight(url, "/"),
			strings.TrimLeft(chartAddress, "/"),
		)
	}

	// download chart
	chartBytes, err := get.Get(chartAddress)
	if err != nil {
		return fmt.Errorf("error downloading chart: %v", err)
	}

	outFile, err := os.Create(outFilePath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, chartBytes)

	return err
}

func (e *Executor) HelmInstall(chart *config.ChartConfig, kubeconfigPath string) error {

	var err error
	for retry := 0; retry <= 10; retry++ {
		err = e.helmInstall(chart, kubeconfigPath)
		if err == nil {
			// no error exity early
			break
		}
		time.Sleep(5 * time.Second)
	}
	return err
}

func (e *Executor) helmInstall(chart *config.ChartConfig, kubeconfigPath string) error {

	if chart.Namespace == "" {
		chart.Namespace = DEFAULT_NAMESPACE
	}

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

	histClient := action.NewHistory(actionConfig)
	releases, err := histClient.Run(chart.ReleaseName)
	if err == nil && len(releases) > 0 {
		upgradeClient := action.NewUpgrade(actionConfig)
		upgradeClient.Namespace = chart.Namespace
		upgradeClient.ChartPathOptions.Version = chart.Version
		_, err = upgradeClient.Run(chart.ReleaseName, chartObj, chart.Values)

	} else {

		iCli := action.NewInstall(actionConfig)
		iCli.Namespace = chart.Namespace
		iCli.ReleaseName = chart.ReleaseName
		_, err = iCli.Run(chartObj, chart.Values)
	}

	return err
}
