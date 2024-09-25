package orchestrator

import (
	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/config"
	"github.com/ish-xyz/kubernetes-lab/bootstrap-manager/pkg/executor"
	corev1 "k8s.io/api/core/v1"
)

const (
	OWNER_KEY = "owner"
)

type Orchestrator struct {
	Executor    *executor.Executor
	Config      *config.Config
	Leader      string
	OrderedList []corev1.ConfigMap
}

func NewOrchestrator(e *executor.Executor, cfg *config.Config) *Orchestrator {
	return &Orchestrator{
		Executor: e,
		Config:   cfg,
	}
}
