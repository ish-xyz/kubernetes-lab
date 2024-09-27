package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/config"
	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/executor"
	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/orchestrator"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
	yaml "gopkg.in/yaml.v3"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	rootCmd = cobra.Command{
		Use:   "bootstrap-manager",
		Short: "Utility used to 'convert' a kubernetes controlplane from systemd to pods",
		RunE:  start,
	}
	kubeconfig *string
)

func init() {
	kubeconfig = rootCmd.Flags().StringP("config", "c", "", "Pass config file for CLI")

	rootCmd.MarkFlagRequired("config")
}

func main() {
	rootCmd.Execute()
}

func loadConfig(cfgFile string) (*config.Config, error) {
	var cfg *config.Config
	fstream, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(fstream, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func getKubeConfig(kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

func getKubeClient(kubeconfig *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(kubeconfig)
}

func getDynamicClient(kubeconfig *rest.Config) (*dynamic.DynamicClient, error) {
	return dynamic.NewForConfig(kubeconfig)
}

func getDiscoveryClient(kubeconfig *rest.Config) (*discovery.DiscoveryClient, error) {
	return discovery.NewDiscoveryClientForConfig(kubeconfig)
}

func getRestMapper(dsc *discovery.DiscoveryClient) *restmapper.DeferredDiscoveryRESTMapper {
	return restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dsc))
}

func start(cmd *cobra.Command, args []string) error {

	// TODO: add initial sleep time for original control plane to start
	rand.Seed(uint64(time.Now().UnixNano()))

	cfg, err := loadConfig(*kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load configuration file: %v", err)
	}

	kubeConfigObj, err := getKubeConfig(cfg.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to get rest.Config from kubeconfig path '%s' => %v", cfg.Kubeconfig, err)
	}

	// TODO: validate configuration

	kcl, err := getKubeClient(kubeConfigObj)
	if err != nil {
		return err
	}

	dvc, err := getDynamicClient(kubeConfigObj)
	if err != nil {
		return fmt.Errorf("failed to init the dynamic client: %v", err)
	}

	dsc, err := getDiscoveryClient(kubeConfigObj)
	if err != nil {
		return fmt.Errorf("failed to init the discovery client: %v", err)
	}

	rsm := getRestMapper(dsc)
	systemdConn, err := dbus.NewSystemConnectionContext(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to initiate connection to dbus for systemd management")
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	exec := executor.NewExecutor(
		systemdConn,
		kcl,
		dsc,
		dvc,
		rsm,
		"bootstrap-manager=true",
		cfg.Sync.Namespace,
		fmt.Sprintf("%s-%s", cfg.Sync.Prefix, cfg.NodeName),
		cfg.NodeName,
	)
	orch := orchestrator.NewOrchestrator(
		exec,
		cfg,
	)

	return orch.RunMainWorkflow()
}
