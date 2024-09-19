package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/executor"
	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/orchestrator"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Kubeconfig string `yaml:"kubeconfig"`
	NodeName   string `yaml:"nodeName"`
	NodesCount int    `yaml:"nodesCount"`
	Namespace  string `yaml:"namespace"`
	Prefix     string `yaml:"prefix"`
}

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

func loadConfig(cfgFile string) (*Config, error) {
	var cfg *Config
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

func getKubeClient(kubeconfig string) (*kubernetes.Clientset, error) {

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	// create clientset (set of muliple clients) for each Group (e.g. Core),
	// the Version (V1) of Group and Kind (e.g. Pods) so GVK.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func start(cmd *cobra.Command, args []string) error {

	// TODO: add initial sleep time for original control plane to start
	rand.Seed(uint64(time.Now().UnixNano()))

	cfg, err := loadConfig(*kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load configuration file: %v", err)
	}

	// TODO: validate configuration

	kcl, err := getKubeClient(cfg.Kubeconfig)
	if err != nil {
		return err
	}

	exec := executor.NewExecutor(
		kcl,
		"bootstrap-manager=true",
		cfg.Namespace,
		fmt.Sprintf("%s-%s", cfg.Prefix, cfg.NodeName),
		cfg.NodeName,
	)
	orch := orchestrator.NewOrchestrator(
		exec,
		cfg.Namespace,
		cfg.NodeName,
	)

	return orch.RunMainWorkflow()
}
