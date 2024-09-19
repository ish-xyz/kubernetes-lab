package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Kubeconfig   string `yaml:"kubeconfig"`
	NodeName     string `yaml:"nodeName"`
	OrderManager struct {
		Namespace string `yaml:"namespace"`
		ConfigMap string `yaml:"configmap"`
	} `yaml:"orderManager"`
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

func start(cmd *cobra.Command, args []string) error {
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

	createBootStrapData(kcl, cfg.OrderManager.Namespace, cfg.OrderManager.ConfigMap, cfg.NodeName)
	checkForMigration(kcl, cfg.OrderManager.Namespace)
	return nil
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

func createBootStrapData(kcl *kubernetes.Clientset, ns, cm, nodeName string) error {

	cmName := fmt.Sprintf("%s-%s", cm, nodeName)
	delay := rand.Intn(5)

	logrus.Infof("waiting for bootstrap delay of %ds ...", delay)
	time.Sleep(time.Duration(delay) * time.Second)

	cmObj := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cmName,
			Labels: map[string]string{
				"bootstrap-manager": "true",
				"owner":             nodeName,
			},
		},
		Data: map[string]string{
			"apiServer":         "false",
			"controllerManager": "false",
			"scheduler":         "false",
		},
	}
	_, err := kcl.CoreV1().ConfigMaps(ns).Create(context.TODO(), &cmObj, metav1.CreateOptions{})

	return err
}

func checkForMigration(kcl *kubernetes.Clientset, ns string) error {

	objects, err := kcl.CoreV1().ConfigMaps(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: "bootstrap-manager=true"})
	if err != nil {
		return err
	}

	// Sort configmaps by creation date
	sort.Slice(objects.Items, func(i, j int) bool {
		return objects.Items[i].CreationTimestamp.Before(&objects.Items[j].CreationTimestamp)
	})

	retry, retryLimit := 0, 10
	interval := 5
	for _, obj := range objects.Items {

		var cmObj *corev1.ConfigMap
		var err error
		for {
			cmObj, err = kcl.CoreV1().ConfigMaps(ns).Get(context.TODO(), obj.ObjectMeta.Name, metav1.GetOptions{})
			if err != nil && retry >= retryLimit {
				return err
			}
			if err != nil {
				retry++
			} else {
				break
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}

		if cmObj.Data["apiServer"] == "true" &&
			cmObj.Data["controllerManager"] == "true" &&
			cmObj.Data["scheduler"] == "true" {

			// node already migrated, skip to next one
			continue
		}

		// check status, if all true -> skip
		// check object ownership
		// if not ownership =>
	}

	return nil
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
