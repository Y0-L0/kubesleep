package kubesleep

import "maps"

type suspendableNamespace interface {
	suspendable() bool
	suspend(*K8Simpl) error
	wake(*K8Simpl) error
}

type suspendableNamespaceImpl struct {
	name         string
	_suspendable bool
}

func (n *suspendableNamespaceImpl) suspendable() bool {
	return n._suspendable
}

func (n *suspendableNamespaceImpl) wake(k8s *K8Simpl) error {
	stateFile, err := k8s.getStateFile(n.name)
	if err != nil {
		return err
	}

	for _, s := range stateFile.suspendables {
		err = s.wake(n.name, k8s)
		if err != nil {
			return err
		}
	}
	k8s.deleteStateFile(n.name)
	return nil
}

func (n *suspendableNamespaceImpl) suspend(k8s *K8Simpl) error {
	suspendables, err := k8s.getDeployments(n.name)
	if err != nil {
		return err
	}
	sus, err := k8s.getStatefulSets(n.name)
	if err != nil {
		return err
	}
	maps.Copy(suspendables, sus)

	stateFile := &suspendStateFile{
		suspendables: suspendables,
		finished:     false,
	}
	stateFile, err = k8s.createStateFile(n.name, stateFile)
	if err != nil {
		return err
	}

	for _, sus := range suspendables {
		err = sus.suspend()
		if err != nil {
			return err
		}
	}

	stateFile.finished = true
	stateFile, err = k8s.updateStateFile(n.name, stateFile)
	if err != nil {
		return err
	}

	return nil
}
