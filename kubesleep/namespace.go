package kubesleep

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
)

type SuspendableNamespace interface {
	Suspendable() bool
	suspend(K8S) error
	wake(K8S) error
}

type suspendableNamespaceImpl struct {
	name         string
	_suspendable bool
}

func NewSuspendableNamespace(name string, suspendable bool) SuspendableNamespace {
	return &suspendableNamespaceImpl{
		name:         name,
		_suspendable: suspendable,
	}
}

func (n *suspendableNamespaceImpl) Suspendable() bool {
	return n._suspendable
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

func (n *suspendableNamespaceImpl) ensureStateFile(k8s K8S, stateFile *SuspendStateFile) (*SuspendStateFile, StateFileActions, error) {
	var target StatefileAlreadyExistsError

	actions, err := k8s.CreateStateFile(n.name, stateFile)
	if err == nil {
		slog.Debug("No existnig statefile found. Creating a new one to save the starting conditions.")
		return stateFile, actions, nil
	}
	if !errors.As(err, &target) {
		slog.Error("Statefile creation failed for an unknown reason", "namespace", n.name)
		return nil, nil, err
	}

	slog.Debug("Statefile already exists. Reading exisitng statefile and merging it with the current state in the cluster.")
	var existingStateFile *SuspendStateFile
	existingStateFile, actions, err = k8s.GetStateFile(n.name)
	if err != nil {
		return nil, nil, err
	}
	return existingStateFile.merge(stateFile), actions, nil
}

func (n *suspendableNamespaceImpl) suspend(k8s K8S) error {
	suspendables, err := k8s.GetDeployments(n.name)
	if err != nil {
		return err
	}
	sus, err := k8s.GetStatefulSets(n.name)
	if err != nil {
		return err
	}
	maps.Copy(suspendables, sus)

	stateFile, actions, err := n.ensureStateFile(k8s, &SuspendStateFile{
		suspendables: suspendables,
		finished:     false,
	})
	if err != nil {
		return err
	}

	slog.Debug("Suspending workloads", "stateFile", stateFile)

	for _, sus := range suspendables {
		err = repeat(sus.Suspend)
		if err != nil {
			return err
		}
	}

	stateFile.finished = true
	return actions.Update(stateFile)
}
