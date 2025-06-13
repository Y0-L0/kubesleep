package kubesleep

import "maps"

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
	stateFile, err := k8s.GetStateFile(n.name)
	if err != nil {
		return err
	}

	for _, s := range stateFile.suspendables {
		err = s.wake(n.name, k8s)
		if err != nil {
			return err
		}
	}
	k8s.DeleteStateFile(n.name)
	return nil
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

	stateFile := &SuspendStateFile{
		suspendables: suspendables,
		finished:     false,
	}
	stateFile, err = k8s.CreateStateFile(n.name, stateFile)
	if err != nil {
		return err
	}

	for _, sus := range suspendables {
		err = sus.Suspend()
		if err != nil {
			return err
		}
	}

	stateFile.finished = true
	stateFile, err = k8s.UpdateStateFile(n.name, stateFile)
	if err != nil {
		return err
	}

	return nil
}
