package kubesleep

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func newParser(args []string, k8sFactory func() (K8S, error)) (*cobra.Command, *cliConfig) {
	slog.Debug("raw cli arguments", "args", args)

	config := &cliConfig{}

	rootCmd := &cobra.Command{
		Use:   "kubesleep",
		Short: "kubesleep can sleep and wake kubernetes namespaces by scaling workloads down to zero and back up",
	}

	rootCmd.SetArgs(args[1:])

	suspendCmd := &cobra.Command{
		Use:   "suspend",
		Short: "Suspend one or multiple kubernetes namespaces.",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Parsed cli arguments for the sleep subcommand", "config", config)
			config.suspend(k8sFactory)
			return nil
		},
	}
	suspendCmd.Flags().StringVarP(
		&config.namespace,
		"namespace",
		"n",
		"",
		"Kubernetes namespace",
	)
	suspendCmd.Flags().BoolVarP(
		&config.force,
		"force",
		"f",
		false,
		"Ignore the do-not-suspend label on the namespace",
	)

	resumeCmd := &cobra.Command{
		Use:   "wake",
		Short: "Wake a kubernetes namespace back up",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Parsed cli arguments for the wake subcommand", "config", config)
			config.wake(k8sFactory)
			return nil
		},
	}
	resumeCmd.Flags().StringVarP(
		&config.namespace,
		"namespace",
		"n",
		"",
		"Kubernetes namespace",
	)
	resumeCmd.MarkFlagRequired("namespace")

	rootCmd.AddCommand(suspendCmd, resumeCmd)
	return rootCmd, config
}

func Main(args []string, k8sFactory func() (K8S, error)) error {
	command, _ := newParser(args, k8sFactory)

	err := command.Execute()
	if err != nil {
		return err
	}

	return nil
}
