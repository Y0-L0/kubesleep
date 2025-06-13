package kubesleep

import "log/slog"

type cliConfig struct {
	namespace string
	force     bool
}

func (c cliConfig) suspend(k8sFactory func() (K8S, error)) error {
	k8s, err := k8sFactory()
	if err != nil {
		return err
	}

	ns, err := k8s.GetSuspendableNamespace(c.namespace)
	if err != nil {
		return err
	}

	if !ns.Suspendable() && !c.force {
		slog.Info("Skipping namespace", "namespace", c.namespace, "force", c.force, "Suspendable", ns.Suspendable())
		return nil
	}
	err = ns.suspend(k8s)
	if err != nil {
		return err
	}

	return nil
}

func (c cliConfig) wake(k8sFactory func() (K8S, error)) error {
	if c.namespace == "" {
		panic("Invalid namespace value")
	}
	k8s, err := k8sFactory()
	if err != nil {
		return err
	}
	ns, err := k8s.GetSuspendableNamespace(c.namespace)
	if err != nil {
		return err
	}
	err = ns.wake(k8s)
	if err != nil {
		return err
	}
	return nil
}
