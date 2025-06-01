package kubesleep

type suspendableNamespace interface {
	suspendable() bool
	suspend(k8simpl) error
}

type suspendableNamespaceImpl struct {
	name         string
	_suspendable bool
}

func (n *suspendableNamespaceImpl) suspendable() bool {
	return n._suspendable
}

func (n *suspendableNamespaceImpl) suspend(k8s k8simpl) error {
	suspendables, err := k8s.getDeployments(n.name)
	if err != nil {
		return err
	}
	sus, err := k8s.getStatefulsets(n.name)
	if err != nil {
		return err
	}
	suspendables = append(suspendables, sus...)

  stateFile := &suspendStateFile{
    suspendables: suspendables,
    finished: false,
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
