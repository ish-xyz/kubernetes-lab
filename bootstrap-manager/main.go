package main

import (
	"context"
	"encoding/json"
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
	"k8s.io/apimachinery/pkg/types"
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

func createBootStrapData(kcl *kubernetes.Clientset, ns, cm, nodeName string) error {

	cmName := fmt.Sprintf("%s-%s", cm, nodeName)
	delay := rand.Intn(5000) + 1000

	logrus.Infof("waiting for bootstrap delay of %dms ...", delay)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	cmObj := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cmName,
			Labels: map[string]string{
				"bootstrap-manager": "true",
			},
		},
		Data: map[string]string{
			"apiServer":         "false",
			"controllerManager": "false",
			"scheduler":         "false",
			"owner":             nodeName,
		},
	}
	_, err := kcl.CoreV1().ConfigMaps(ns).Create(context.TODO(), &cmObj, metav1.CreateOptions{})

	return err
}

func getConfigMapWithRetries(kcl *kubernetes.Clientset, namespace, name string, maxRetries, interval int) (*corev1.ConfigMap, error) {

	var cmObj *corev1.ConfigMap
	var err error
	var retry = 0
	for {
		cmObj, err = kcl.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil && retry >= maxRetries {
			return nil, err
		}
		if err != nil {
			retry++
		} else {
			break
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}

	return cmObj, err
}

func waitForMigration(kcl *kubernetes.Clientset, namespace, name string, maxRetries, interval int) (bool, error) {

	retry := 0
	for {
		cmObj, err := getConfigMapWithRetries(kcl, namespace, name, 5, 3)
		if err != nil {
			return false, err
		}

		if cmObj.Data["apiServer"] == "true" &&
			cmObj.Data["controllerManager"] == "true" &&
			cmObj.Data["scheduler"] == "true" {
			return true, nil
		}

		if retry >= maxRetries {
			break
		} else {
			retry++
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return false, nil
}

func performMigration(kcl *kubernetes.Clientset, cmObj *corev1.ConfigMap) error {

	fmt.Println("(TODO) performing migration of this node! (3s)")

	time.Sleep(3 * time.Second)

	cmObj.Data["apiServer"] = "true"
	cmObj.Data["controllerManager"] = "true"
	cmObj.Data["scheduler"] = "true"

	patchBytes, err := json.Marshal(cmObj)
	if err != nil {
		return err
	}
	_, err = kcl.CoreV1().ConfigMaps(cmObj.Namespace).Patch(context.TODO(), cmObj.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	return err
}

func migration(kcl *kubernetes.Clientset, namespace, nodeName string) error {

	// TODO: check against desired number of configmaps and retry in case
	objects, err := kcl.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "bootstrap-manager=true"})
	if err != nil {
		return err
	}

	// Sort configmaps by creation date
	sort.Slice(objects.Items, func(i, j int) bool {
		return objects.Items[i].CreationTimestamp.Before(&objects.Items[j].CreationTimestamp)
	})

	for _, obj := range objects.Items {

		logrus.Infof("processing node: %s ...", obj.Name)
		cmObj, err := getConfigMapWithRetries(kcl, namespace, obj.Name, 5, 3)
		if err != nil {
			return err
		}

		// check until completed
		if cmObj.Data["apiServer"] == "true" &&
			cmObj.Data["controllerManager"] == "true" &&
			cmObj.Data["scheduler"] == "true" {

			// node already migrated, skip to next one
			logrus.Infof("node '%s' already migrated, skipping.", obj.Name)
			continue
		}

		if cmObj.Data["owner"] != nodeName {
			logrus.Infof("waiting for node's migration ('%s')...", obj.Name)
			res, err := waitForMigration(kcl, namespace, cmObj.Name, 20, 15)
			if err != nil {
				return fmt.Errorf("failed while checking migration results for node %s with error: %v", obj.Name, err)
			}
			if !res {
				return fmt.Errorf("migration for node %s took too long, aborting", obj.Name)
			}
		}

		logrus.Infof("performing node migration ('%s')!", obj.Name)
		err = performMigration(kcl, cmObj)
		if err != nil {
			return err
		}
	}

	return nil
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

	err = createBootStrapData(kcl, cfg.OrderManager.Namespace, cfg.OrderManager.ConfigMap, cfg.NodeName)
	if err != nil {
		return err
	}

	return migration(kcl, cfg.OrderManager.Namespace, cfg.NodeName)
}
