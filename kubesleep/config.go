package kubesleep

import "log/slog"

type cliConfig struct {
	namespaces    []string
	force         bool
	allNamespaces bool
}

var PROTECTED_NAMESPACES = []string{"default", "kube-node-lease", "kube-public", "kube-system", "ingress", "istio", "local-path"}

func (c cliConfig) suspend(k8sFactory func() (K8S, error)) error {
	k8s, err := k8sFactory()
	if err != nil {
		return err
	}

	if c.allNamespaces {
		c.namespaces, err = k8s.GetNamespaces()
		if err != nil {
			return err
		}
	}

	if c.namespaces == nil || len(c.namespaces) == 0 {
		panic("invalid namespaces value")
	}
	for _, namespace := range c.namespaces {
		ns, err := k8s.GetSuspendableNamespace(namespace)
		if err != nil {
			return err
		}

		if !ns.Suspendable() && !c.force {
			slog.Info("Skipping namespace", "namespace", c.namespaces, "force", c.force, "Suspendable", ns.Suspendable())
			return nil
		}
		err = ns.suspend(k8s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c cliConfig) wake(k8sFactory func() (K8S, error)) error {
	if c.namespaces == nil {
		panic("Invalid namespaces value")
	}
	for _, namespace := range c.namespaces {
		if namespace == "" {
			panic("Invalid namespace value")
		}
		k8s, err := k8sFactory()
		if err != nil {
			return err
		}
		ns, err := k8s.GetSuspendableNamespace(namespace)
		if err != nil {
			return err
		}
		err = ns.wake(k8s)
		if err != nil {
			return err
		}
	}
	return nil
}
