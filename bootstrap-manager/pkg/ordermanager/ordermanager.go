package ordermanager

import "k8s.io/client-go/kubernetes"

type OrderManager struct {
	KubeClient    kubernetes.Clientset
	LabelSelector string
	Executor      *Executor
}

type Executor struct {
	Blah string
}
