package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Kubeconfig     string `yaml:"kubeconfig"`
	LeaderElection struct {
		Namespace string `yaml:"namespace"`
		ConfigMap string `yaml:"configmap"`
	} `yaml:"leaderElection"`
}

var (
	rootCmd = cobra.Command{
		Use:   "bootstrap-manager",
		Short: "Utility used to 'convert' a kubernetes controlplane from systemd to pods",
		RunE:  start,
	}
	configPtr *string
)

func init() {
	configPtr = rootCmd.Flags().StringP("config", "c", "", "Pass config file for CLI")

	rootCmd.MarkFlagRequired("config")
}

func main() {
	rootCmd.Execute()
}

func start(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(*configPtr)
	if err != nil {
		return fmt.Errorf("failed to load configuration file: %v", err)
	}

	kclient := getKubeClient(cfg.Kubeconfig)

	return nil
}

func getKubeClient(kubeconfig string) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	// create clientset (set of muliple clients) for each Group (e.g. Core),
	// the Version (V1) of Group and Kind (e.g. Pods) so GVK.
	clientset, err := kubernetes.NewForConfig(config)

	return clientset
}

func orderManager() {
	leeway := rand.Intn(10)
	time.Sleep(leeway * time.Second)

	fmt.Println("trying to connect to the kubernetes api")
	// cm := corev1.ConfigMap{
	// 	TypeMeta: metav1.TypeMeta{
	// 	  Kind:       "ConfigMap",
	// 	  APIVersion: "v1",
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 	  Name:      "my-config-map",
	// 	  Namespace: "my-namespace",
	// 	},
	// 	Data: <config-map-data>,
	//   }

	//   clientset.CoreV1().ConfigMaps("my-namespace").Create(&cm)
}

func loadConfig(cfgFile string) (*Config, error) {
	var cfg *Config

	fstream, err := os.ReadFile(*configPtr)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(fstream, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
