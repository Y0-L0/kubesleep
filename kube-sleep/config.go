package kubesleep

import "log/slog"

type cliConfig struct {
	namespace string
	force     bool
}

func (c cliConfig) suspend(k8sFactory func() (*K8Simpl, error)) error {
	k8s, err := k8sFactory()
	if err != nil {
		return err
	}

	ns, err := k8s.suspendableNamespace(c.namespace)
	if err != nil {
		return err
	}

	if !ns.suspendable() && !c.force {
		slog.Info("Skipping namespace", "namespace", c.namespace, "force", c.force, "suspendable", ns.suspendable())
		return nil
	}
	err = ns.suspend(k8s)
	if err != nil {
		return err
	}

	return nil
}

func (c cliConfig) wake(k8sFactory func() (*K8Simpl, error)) error {
	if c.namespace == "" {
		panic("Invalid namespace value")
	}
	k8s, err := k8sFactory()
	if err != nil {
		return err
	}
	ns, err := k8s.suspendableNamespace(c.namespace)
	if err != nil {
		return err
	}
	err = ns.wake(k8s)
	if err != nil {
		return err
	}
	return nil
}
