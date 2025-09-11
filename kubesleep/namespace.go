package kubesleep

import (
	"errors"
	"fmt"
	"log/slog"

	"k8s.io/utils/strings/slices"
)

var PROTECTED_NAMESPACES = []string{"default", "kube-node-lease", "kube-public", "kube-system", "ingress-nginx", "istio", "local-path"}

type SuspendableNamespace interface {
	Name() string
	Protected() bool
	autoProtected() bool
	suspend(K8S) error
	wake(K8S) error
}

type suspendableNamespaceImpl struct {
	name      string
	protected bool
}

func NewSuspendableNamespace(name string, protected bool) SuspendableNamespace {
	return &suspendableNamespaceImpl{
		name:      name,
		protected: protected,
	}
}

func (n *suspendableNamespaceImpl) Protected() bool {
	return n.protected || n.autoProtected()
}
func (n *suspendableNamespaceImpl) autoProtected() bool {
	return slices.Contains(PROTECTED_NAMESPACES, n.name)
}
func (n *suspendableNamespaceImpl) Name() string {
	return n.name
}

func (n *suspendableNamespaceImpl) wake(k8s K8S) error {
	stateFile, actions, err := k8s.GetStateFile(n.name)
	if err != nil {
		return err
	}

	if !stateFile.finished {
		return fmt.Errorf("cannot wake the namespace %s because the namespace is partially suspended. Please first resume / retry the suspend operation", n.name)
	}

	for _, s := range stateFile.suspendables {
		err = repeat(func() error { return s.wake(n.name, k8s) })
		if err != nil {
			return err
		}
	}
	return actions.Delete()
}

func (n *suspendableNamespaceImpl) ensureStateFile(k8s K8S, stateFile *SuspendState) (*SuspendState, SuspendStateActions, error) {
	var alreadyExists StatefileAlreadyExistsError

	actions, err := k8s.CreateStateFile(n.name, stateFile.Write())
	if err == nil {
		slog.Debug("No existing statefile found. Creating a new one to save the starting conditions.", "namespace", n.name)
		return stateFile, actions, nil
	}
	if !errors.As(err, &alreadyExists) {
		slog.Error("Statefile creation failed for an unknown reason", "namespace", n.name)
		return nil, nil, err
	}

	slog.Debug("Statefile already exists. Reading existing statefile and merging it with the current state in the cluster.", "namespace", n.name)
	var existingStateFile *SuspendState
	existingStateFile, actions, err = k8s.GetStateFile(n.name)
	if err != nil {
		return nil, nil, err
	}
	return existingStateFile.merge(stateFile), actions, nil
}

func (n *suspendableNamespaceImpl) suspend(k8s K8S) error {
	suspendables, err := k8s.GetSuspendables(n.name)
	if err != nil {
		return err
	}

	stateFile, actions, err := n.ensureStateFile(k8s, &SuspendState{
		suspendables: suspendables,
		finished:     false,
	})
	if err != nil {
		return err
	}

	slog.Debug("Suspending workloads", "stateFile", stateFile, "namespace", n.name)

	for _, sus := range suspendables {
		err = repeat(sus.Suspend)
		if err != nil {
			return err
		}
	}

	stateFile.finished = true
	return actions.Update(stateFile.Write())
}
