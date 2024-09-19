package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Executor struct {
	KubeClient    *kubernetes.Clientset
	LabelSelector string // LabelSelector: "bootstrap-manager=true"
	Namespace     string
	CMName        string
	NodeName      string
}

func NewExecutor(kcl *kubernetes.Clientset, ls, ns, cmn, nn string) *Executor {

	return &Executor{
		CMName:        cmn,
		KubeClient:    kcl,
		LabelSelector: ls,
		Namespace:     ns,
		NodeName:      nn,
	}
}

// Create Bootstrap Configmaps
func (e *Executor) CreateBootstrapData() error {

	delay := rand.Intn(5000) + 1000
	logrus.Infof("waiting for bootstrap delay of %dms ...", delay)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	cmObj := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: e.CMName,
			Labels: map[string]string{
				"bootstrap-manager": "true",
			},
		},
		Data: map[string]string{
			"apiServer":         "false",
			"controllerManager": "false",
			"scheduler":         "false",
			"owner":             e.NodeName,
		},
	}
	_, err := e.KubeClient.CoreV1().ConfigMaps(e.Namespace).Create(context.TODO(), &cmObj, metav1.CreateOptions{})

	return err
}

func (e *Executor) PatchConfigMap() {
	return
}

func (e *Executor) ListConfigMaps(desiredNumber, maxRetries, interval int) (*corev1.ConfigMapList, error) {

	var err error
	var objects *corev1.ConfigMapList

	for retry := 0; retry < maxRetries; retry++ {
		time.Sleep(time.Duration(interval) * time.Second)
		objects, err = e.KubeClient.CoreV1().ConfigMaps(e.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: e.LabelSelector})
		if err != nil {
			continue
		}
		if len(objects.Items) != desiredNumber && desiredNumber > 0 {
			continue
		}

		return objects, nil
	}

	if len(objects.Items) != desiredNumber && desiredNumber > 0 && err == nil {
		return nil, fmt.Errorf("failed to retrieve configmaps, desired number of %d doesn't match actual items length of %d", desiredNumber, len(objects.Items))
	}

	return nil, fmt.Errorf("error while listing configmaps: %v", err)
}

func (e *Executor) GetConfigMap(namespace, name string, maxRetries, interval int) (*corev1.ConfigMap, error) {

	var cmObj *corev1.ConfigMap
	var err error
	var retry = 0
	for {
		cmObj, err = e.KubeClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
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
