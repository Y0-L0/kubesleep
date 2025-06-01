package kubesleep

import (
	"errors"
	"log/slog"

	flag "github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type cliConfig struct {
	namespace string
  force bool
}

func parseFlags(args []string) (cliConfig, error) {
	slog.Debug("raw cli arguments", "args", args)

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	namespace := flags.StringP("namespace", "n", "", "Kubernetes namespace")
  force := flags.BoolP("force", "f", false, "Ignore the do-not-suspend label on the namespace")

	err := flags.Parse(args[1:])
	if err != nil {
		return cliConfig{}, err
	}

	if *namespace == "" {
		return cliConfig{}, errors.New("-n or --namespace is required")
	}
	config := cliConfig{
		namespace: *namespace,
    force: *force,
	}
	slog.Debug("Parsed cli arguments", "config", config)
	return config, nil
}

func Main(args []string) error {
	config, err := parseFlags(args)
	if err != nil {
		return err
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	k8s := k8simpl{clientset: clientset}

	ns, err := k8s.suspendableNamespace(config.namespace)
	if err != nil {
		return err
	}

	if !ns.suspendable() && !config.force {
    slog.Info("Skipping namespace", "namespace", config.namespace, "force", config.force, "suspendable", ns.suspendable())
		return nil
	}
	err = ns.suspend(k8s)
	if err != nil {
		return err
	}

	return nil
}
