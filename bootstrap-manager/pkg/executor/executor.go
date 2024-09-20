package executor

import (
	"k8s.io/client-go/kubernetes"
)

type Executor struct {
	KubeClient *kubernetes.Clientset
	// Dynamic Client
	// Discovery Client
	// Rest Mapper
	LabelSelector string // LabelSelector: "bootstrap-manager=true"
	Namespace     string
	CMName        string
	NodeName      string
	TempFolder    string
}

func NewExecutor(kcl *kubernetes.Clientset, ls, ns, cmn, nn string) *Executor {

	return &Executor{
		CMName:        cmn,
		KubeClient:    kcl,
		LabelSelector: ls,
		Namespace:     ns,
		NodeName:      nn,
		TempFolder:    "/tmp/",
	}
}
