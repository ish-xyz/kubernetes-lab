package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/config"
	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/executor"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Orchestrator struct {
	Executor *executor.Executor
	Config   *config.Config
}

func NewOrchestrator(e *executor.Executor, cfg *config.Config) *Orchestrator {
	return &Orchestrator{
		Executor: e,
		Config:   cfg,
	}
}

func (o *Orchestrator) fakeMigration(cmObj *corev1.ConfigMap) error {

	fmt.Println("(TODO) performing migration of this node! (3s)")

	time.Sleep(3 * time.Second)

	cmObj.Data["apiServer"] = "true"
	cmObj.Data["controllerManager"] = "true"
	cmObj.Data["scheduler"] = "true"

	patchBytes, err := json.Marshal(cmObj)
	if err != nil {
		return err
	}
	// TODO: Run via executor
	_, err = o.Executor.KubeClient.CoreV1().ConfigMaps(cmObj.Namespace).Patch(context.TODO(), cmObj.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	return err
}

func (o *Orchestrator) waitForMigration(namespace, name string, maxRetries, interval int) (bool, error) {

	retry := 0
	for {
		cmObj, err := o.Executor.GetConfigMap(namespace, name, 5, 3)
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

func (o *Orchestrator) RunMigrationWorkflow(namespace, nodeName string) error {

	objects, err := o.Executor.ListConfigMaps(3, 15, 5)
	if err != nil {
		return err
	}

	// Sort configmaps by creation date
	sort.Slice(objects.Items, func(i, j int) bool {
		return objects.Items[i].CreationTimestamp.Before(&objects.Items[j].CreationTimestamp)
	})

	for _, obj := range objects.Items {

		logrus.Infof("processing node: %s ...", obj.Name)
		cmObj, err := o.Executor.GetConfigMap(namespace, obj.Name, 5, 3)
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
			res, err := o.waitForMigration(namespace, cmObj.Name, 20, 15)
			if err != nil {
				return fmt.Errorf("failed while checking migration results for node %s with error: %v", obj.Name, err)
			}
			if !res {
				return fmt.Errorf("migration for node %s took too long, aborting", obj.Name)
			}
		}

		logrus.Infof("performing node migration ('%s')!", obj.Name)
		err = o.fakeMigration(cmObj)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Orchestrator) RunMainWorkflow() error {

	// TODO:
	// pre-flight checks
	// check systemd services
	// check for api-server to come up

	// PreMigration stesp
	// o.RunPreMigrationWorkflow()

	logrus.Infoln("starting pre migration workflow...")
	for _, pkg := range o.Config.PreMigration {
		logrus.Infof("package: %s with driver %s", pkg.Name, pkg.Driver)
		if pkg.Driver == "helm" {
			err := o.Executor.HelmInstall(pkg.Chart, o.Config.Kubeconfig)
			fmt.Println(err)
		}
	}

	return nil

	err := o.Executor.CreateBootstrapData()
	if err != nil {
		return err
	}

	return o.RunMigrationWorkflow(o.Config.Sync.Resources.Namespace, o.Config.NodeName)

	// PostMigration
	// o.RunPostMigrationWorkflow()

	// Final steps
	// o.RunFinalWorkflow()
}
