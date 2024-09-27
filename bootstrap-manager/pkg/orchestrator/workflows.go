package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/executor"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
	corev1 "k8s.io/api/core/v1"
)

func (o *Orchestrator) runInitialChecks() error {

	// pre migration checks
	for _, cfg := range o.Config.PreMigration {
		if cfg.Driver == executor.KUBECTL_DRIVER {
			if _, err := os.Stat(cfg.Manifest); err != nil {
				return fmt.Errorf("failed to stat '%s' => %v", cfg.Manifest, err)
			}
		}
		//TODO: check helm index.yaml
	}

	// migration checks
	for _, cfg := range o.Config.Migration {
		if exists, err := o.Executor.ServiceExists(cfg.SystemdUnit); !exists || err != nil {
			return fmt.Errorf("systemd unit '%s' doesn't not exists or systemd is not running properly '%v'", cfg.SystemdUnit, err)
		}
		if _, err := os.Stat(cfg.Manifest); err != nil {
			return fmt.Errorf("failed to stat '%s' => %v", cfg.Manifest, err)
		}
	}

	return nil
}

func (o *Orchestrator) leaderElection() error {

	delay := rand.Intn(5000) + 1000
	logrus.Infof("waiting for bootstrap delay of %dms ...", delay)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	if err := o.createBootstrapConfigMap(); err != nil {
		return fmt.Errorf("failed creating bootstrap configmap => %v", err)
	}

	configMapList, err := o.Executor.ListBootstrapConfigMaps(o.Config.Sync.NodesCount, 10, 6)
	if err != nil {
		return err
	}

	// Sort configmaps by creation date
	sort.Slice(configMapList.Items, func(i, j int) bool {
		return configMapList.Items[i].CreationTimestamp.Before(&configMapList.Items[j].CreationTimestamp)
	})

	o.Leader = configMapList.Items[0].Data[OWNER_KEY]

	return nil
}

func (o *Orchestrator) createBootstrapConfigMap() error {

	data := map[string]string{OWNER_KEY: o.Config.NodeName}
	for _, migrationConfig := range o.Config.Migration {
		data[migrationConfig.Key] = "false"
	}

	_, err := o.Executor.CreateBootstrapConfigMap(data)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) preMigrationWorkflow() error {

	logrus.Infoln("starting pre migration workflow...")
	for _, pkg := range o.Config.PreMigration {

		if pkg.LeaderOnly && o.Leader != o.Config.NodeName {
			logrus.Infof("not the leader, sleeping for 5 seconds to allow the leader to perform preMigration steps")
			continue
		}

		logrus.Infof(">>> processing package: %s with driver %s", pkg.Name, pkg.Driver)
		if pkg.Driver == "helm" {
			err := o.Executor.HelmInstall(pkg.Chart, o.Config.Kubeconfig)
			if err != nil {
				return fmt.Errorf("helm installation failed: %v", err)
			}

		} else if pkg.Driver == "kubectl" {
			err := o.Executor.KubectlApply(pkg.Manifest)
			if err != nil {
				return fmt.Errorf("kubectl apply failed: %v", err)
			}
		}

		// speculative sleep
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (o *Orchestrator) migrationWorkflow(namespace, nodeName string) error {

	configMapList, err := o.Executor.ListBootstrapConfigMaps(3, 100, 6)
	if err != nil {
		return err
	}

	// Sort configmaps by creation date
	sort.Slice(configMapList.Items, func(i, j int) bool {
		return configMapList.Items[i].CreationTimestamp.Before(&configMapList.Items[j].CreationTimestamp)
	})

	for _, obj := range configMapList.Items {

		logrus.Infof("processing node: %s ...", obj.Name)
		cmObj, err := o.Executor.GetConfigMap(namespace, obj.Name, 5, 3)
		if err != nil {
			return err
		}

		// 1. if migration for the node associated with the configmap is completed
		// we don't care and we move on.
		if o.isMigrationCompleted(cmObj) {
			// node already migrated, skip to next one
			logrus.Infof("node '%s' already migrated, skipping.", obj.Name)
			continue
		}

		// 2. if this node is not the owner of the configmap
		// we need to wait for the migration of the other node to finish
		if cmObj.Data["owner"] != nodeName {

			logrus.Infof("waiting for node's migration ('%s')...", obj.Name)

			res, err := o.waitForMigration(namespace, cmObj.Name, 150, 3) //TODO: set this as parameter
			if err != nil {
				return fmt.Errorf("failed while checking migration results for node %s with error: %v", obj.Name, err)
			}
			if !res {
				return fmt.Errorf("migration for node %s took too long, aborting", obj.Name)
			}

			continue
		}

		// 3. it's our turn to migrate
		logrus.Infof("performing node migration ('%s')!", obj.Name)
		err = o.execMigration(cmObj)
		if err != nil {
			return err
		}

		logrus.Infoln("migration completed successfully in node", o.Config.NodeName)
	}

	return nil
}

func (o *Orchestrator) isMigrationCompleted(bootstrapConfigMap *corev1.ConfigMap) bool {

	keys := func() []string {
		keys := []string{}
		for _, x := range o.Config.Migration {
			keys = append(keys, x.Key)
		}
		return keys
	}()

	res := true
	for _, key := range keys {
		if bootstrapConfigMap.Data[key] != "true" {
			res = false
		}
	}

	return res
}

func (o *Orchestrator) execMigration(cmObj *corev1.ConfigMap) error {

	for _, resource := range o.Config.Migration {

		err := o.Executor.StopService(resource.SystemdUnit)
		if err != nil {
			return err
		}

		err = o.Executor.DisableServices([]string{resource.SystemdUnit})
		if err != nil {
			return err
		}

		if resource.LeaderOnly && o.Leader == o.Config.NodeName {
			err = o.Executor.KubectlApply(resource.Manifest)
			if err != nil {
				return err
			}
		} else {
			logrus.Infoln("skipping apply migration step, should only be performed on leader", o.Leader)
		}

		for _, check := range resource.HTTPChecks {
			urlObj, err := url.Parse(check.URL)
			if err != nil {
				return err
			}
			o.Executor.URLCheck(
				check.URL,
				check.CA,
				check.Insecure,
				urlObj.Scheme == "https",
				check.MaxRetries,
				check.Interval,
			)
		}

		err = o.updateMigrationStatus(cmObj, resource.Key, "true")
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Orchestrator) updateMigrationStatus(cmObj *corev1.ConfigMap, key, val string) error {

	cmObj.Data[key] = val

	patchBytes, err := json.Marshal(cmObj)
	if err != nil {
		return err
	}

	o.Executor.PatchConfigMap(cmObj, patchBytes, 10, 3)
	return err
}

func (o *Orchestrator) waitForMigration(namespace, name string, maxRetries, interval int) (bool, error) {

	for retry := 0; retry <= maxRetries; retry++ {
		cmObj, err := o.Executor.GetConfigMap(namespace, name, 5, 3)
		if err != nil {
			return false, err
		}

		if o.isMigrationCompleted(cmObj) {
			return true, nil
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
	return false, nil
}

func (o *Orchestrator) postMigrationWorkflow() error {

	logrus.Infoln("starting pre migration workflow...")
	for _, pkg := range o.Config.PostMigration {

		if pkg.LeaderOnly && o.Leader != o.Config.NodeName {
			logrus.Infof("not the leader, skipping postMigration")
			continue
		}

		logrus.Infof(">>> processing package: %s with driver %s", pkg.Name, pkg.Driver)
		if pkg.Driver == "helm" {
			err := o.Executor.HelmInstall(pkg.Chart, o.Config.Kubeconfig)
			if err != nil {
				return fmt.Errorf("helm installation failed: %v", err)
			}

		} else if pkg.Driver == "kubectl" {
			err := o.Executor.KubectlApply(pkg.Manifest)
			if err != nil {
				return fmt.Errorf("kubectl apply failed: %v", err)
			}
		}

		// speculative sleep
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (o *Orchestrator) RunMainWorkflow() error {

	// used for dynamicValidation of the config
	err := o.runInitialChecks()
	if err != nil {
		return err
	}

	err = o.leaderElection()
	if err != nil {
		return err
	}

	err = o.preMigrationWorkflow()
	if err != nil {
		return err
	}

	err = o.migrationWorkflow(o.Config.Sync.Namespace, o.Config.NodeName)
	if err != nil {
		return err
	}

	err = o.postMigrationWorkflow()
	if err != nil {
		return err
	}

	return nil
}
