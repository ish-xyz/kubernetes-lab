package executor

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Create Bootstrap Configmaps
func (e *Executor) CreateBootstrapConfigMap(data map[string]string) (*corev1.ConfigMap, error) {

	cm := corev1.ConfigMap{
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
		Data: data,
	}
	cmObj, err := e.KubeClient.CoreV1().ConfigMaps(e.Namespace).Create(context.TODO(), &cm, metav1.CreateOptions{})

	return cmObj, err
}

func (e *Executor) PatchConfigMap(cmObj *corev1.ConfigMap, patch []byte, maxRetries, interval int) error {

	var err error
	for retry := 0; retry < maxRetries; retry++ {
		_, err = e.KubeClient.CoreV1().ConfigMaps(cmObj.Namespace).Patch(context.TODO(), cmObj.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
		if err == nil {
			break
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return err
}

// List boostrap configmaps and wait for all bootstrap-manager to post their own
func (e *Executor) ListBootstrapConfigMaps(desiredNumber, maxRetries, interval int) (*corev1.ConfigMapList, error) {

	var err error
	var objects *corev1.ConfigMapList

	for retry := 0; retry < maxRetries; retry++ {

		objects, err = e.KubeClient.CoreV1().ConfigMaps(e.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: e.LabelSelector})
		if err == nil && len(objects.Items) == desiredNumber {
			return objects, nil
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}

	if len(objects.Items) != desiredNumber {
		return nil, fmt.Errorf("failed to retrieve configmaps, desired number of %d doesn't match actual items length of %d", desiredNumber, len(objects.Items))
	}

	return nil, fmt.Errorf("error while listing configmaps: %v", err)
}

func (e *Executor) GetConfigMap(namespace, name string, maxRetries, interval int) (*corev1.ConfigMap, error) {

	var err error

	for retry := 0; retry <= maxRetries; retry++ {
		cmObj, err := e.KubeClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err == nil && cmObj != nil {
			return cmObj, nil
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}

	return nil, fmt.Errorf("operations took too long => %v", err)
}
