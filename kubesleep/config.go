package kubesleep

import (
	"log/slog"
	"slices"
)

var PROTECTED_NAMESPACES = []string{"default", "kube-node-lease", "kube-public", "kube-system", "ingress", "istio", "local-path"}

type cliConfig struct {
	namespaces    []string
	force         bool
	allNamespaces bool
}

func (c cliConfig) validate() {
	if c.allNamespaces && c.force {
		panic("allNamespaces and force can't both be specified")
	}
	if (c.namespaces == nil || len(c.namespaces) == 0) && !c.allNamespaces {
		panic("invalid namespaces value")
	}
	if slices.Contains(c.namespaces, "") {
		panic("Invalid namespace value")
	}
}

func (c cliConfig) suspend(k8sFactory func() (K8S, error)) error {
	c.validate()

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

	for _, namespace := range c.namespaces {
		ns, err := k8s.GetSuspendableNamespace(namespace)
		if err != nil {
			return err
		}
		if slices.Contains(PROTECTED_NAMESPACES, namespace) && (c.allNamespaces || !c.force) {
			slog.Info("Skipping automatically protected namespace", "namespace", ns.Name(), "force", c.force)
			continue
		}

		if !ns.Suspendable() && !c.force {
			slog.Info("Skipping manually protected namespace", "namespace", ns.Name(), "force", c.force, "Suspendable", ns.Suspendable())
			continue
		}
		err = ns.suspend(k8s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c cliConfig) wake(k8sFactory func() (K8S, error)) error {
	c.validate()

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
