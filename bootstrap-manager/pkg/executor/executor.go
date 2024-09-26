package executor

import (
	"github.com/coreos/go-systemd/v22/dbus"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

const (
	KUBECTL_DRIVER    = "kubectl"
	HELM_DRIVER       = "helm"
	DEFAULT_NAMESPACE = "default"
	YAML_DELIMITER    = "---"
	SYSTEMD_DONE      = "done"
)

type Executor struct {
	SystemdConn     *dbus.Conn
	KubeClient      *kubernetes.Clientset
	DynamicClient   *dynamic.DynamicClient
	DiscoveryClient *discovery.DiscoveryClient
	RestMapper      *restmapper.DeferredDiscoveryRESTMapper
	LabelSelector   string
	Namespace       string
	CMName          string
	NodeName        string
	TempFolder      string
}

func NewExecutor(
	sc *dbus.Conn,
	kcl *kubernetes.Clientset,
	dsc *discovery.DiscoveryClient,
	dvc *dynamic.DynamicClient,
	rsm *restmapper.DeferredDiscoveryRESTMapper,
	ls, ns, cmn, nn string,
) *Executor {
	return &Executor{
		CMName:          cmn,
		KubeClient:      kcl,
		DiscoveryClient: dsc,
		DynamicClient:   dvc,
		RestMapper:      rsm,
		LabelSelector:   ls,
		Namespace:       ns,
		NodeName:        nn,
		SystemdConn:     sc,
		TempFolder:      "/tmp/",
	}
}
