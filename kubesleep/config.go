package kubesleep

import (
	"fmt"
	"io"
	"log/slog"
	"slices"
)

type cliConfig struct {
	namespaces    []string
	force         bool
	allNamespaces bool
	outWriter     io.Writer
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
	if c.outWriter == nil {
		panic("outWriter must be set")
	}
}

func (c cliConfig) getNamespaces(k8s K8S) ([]SuspendableNamespace, error) {
	if c.allNamespaces {
		return k8s.GetSuspendableNamespaces()
	}

	var namespaces []SuspendableNamespace
	for _, n := range c.namespaces {
		ns, err := k8s.GetSuspendableNamespace(n)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, ns)
	}
	return namespaces, nil
}

func (c cliConfig) suspend(k8sFactory func() (K8S, error)) error {
	c.validate()

	k8s, err := k8sFactory()
	if err != nil {
		return err
	}

	namespaces, err := c.getNamespaces(k8s)
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		if ns.autoProtected() && (c.allNamespaces || !c.force) {
			slog.Info("Skipping automatically protected namespace", "namespace", ns.Name(), "autoProtected", ns.autoProtected(), "force", c.force)
			fmt.Fprintf(c.outWriter, "Skipped auto-protected namespace %s\n", ns.Name())
			continue
		}

		if ns.Protected() && !c.force {
			slog.Info("Skipping manually protected namespace", "namespace", ns.Name(), "force", c.force, "Protected", ns.Protected())
			fmt.Fprintf(c.outWriter, "Skipped protected namespace %s\n", ns.Name())
			continue
		}
		err = ns.suspend(k8s)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.outWriter, "Suspended namespace %s\n", ns.Name())
	}
	return nil
}

func (c cliConfig) wake(k8sFactory func() (K8S, error)) error {
	c.validate()
	k8s, err := k8sFactory()
	if err != nil {
		return err
	}

	namespaces, err := c.getNamespaces(k8s)
	if err != nil {
		return err
	}
	for _, ns := range namespaces {
		err = ns.wake(k8s)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.outWriter, "Woke namespace %s\n", ns.Name())
	}
	return nil
}
