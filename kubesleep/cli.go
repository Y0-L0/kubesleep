package kubesleep

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func SetupLogging(logLevel slog.Level) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     logLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func newParser(args []string, k8sFactory func() (K8S, error), setupLogging func(slog.Level)) (*cobra.Command, *cliConfig) {
	slog.Debug("raw cli arguments", "args", args)

	config := &cliConfig{}
	var verbosity int

	rootCmd := &cobra.Command{
		Use:   "kubesleep",
		Short: "kubesleep can sleep and wake kubernetes namespaces by scaling workloads down to zero and back up",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var logLevel slog.Level
			switch verbosity {
			case 0:
				logLevel = slog.LevelWarn
			case 1:
				logLevel = slog.LevelInfo
			default:
				logLevel = slog.LevelDebug
			}
			setupLogging(logLevel)
		},
	}

	rootCmd.SetArgs(args[1:])

	rootCmd.PersistentFlags().CountVarP(
		&verbosity,
		"verbose",
		"v",
		"Increase the log level. Can be specified multiple times.",
	)

	suspendCmd := &cobra.Command{
		Use:   "suspend",
		Short: "Suspend one or multiple kubernetes namespaces.",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Parsed cli arguments for the sleep subcommand", "config", config)
			if !config.allNamespaces && len(config.namespaces) == 0 {
				return fmt.Errorf("either --all-namespaces or --namespace (-n) must be specified")
			}
			if config.allNamespaces && (len(config.namespaces) > 0 || config.force) {
				return fmt.Errorf("--all-namespaces cannot be combined with --namespace or --force")
			}
			return config.suspend(k8sFactory)
		},
	}
	suspendCmd.Flags().StringArrayVarP(
		&config.namespaces,
		"namespace",
		"n",
		nil,
		"Kubernetes namespace. Can be specified multiple times",
	)
	suspendCmd.Flags().BoolVarP(
		&config.force,
		"force",
		"f",
		false,
		"Ignore the do-not-suspend label on the namespace",
	)
	suspendCmd.MarkFlagRequired("namespace")
	// suspendCmd.Flags().BoolVar(
	// 	&config.allNamespaces,
	// 	"all-namespaces",
	// 	false,
	// 	"Suspend all non-protected namespaces",
	// )

	resumeCmd := &cobra.Command{
		Use:   "wake",
		Short: "Wake a kubernetes namespace back up",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Parsed cli arguments for the wake subcommand", "config", config)
			return config.wake(k8sFactory)
		},
	}
	resumeCmd.Flags().StringArrayVarP(
		&config.namespaces,
		"namespace",
		"n",
		nil,
		"Kubernetes namespace. Can be specified multiple times",
	)
	resumeCmd.MarkFlagRequired("namespace")

	rootCmd.AddCommand(suspendCmd, resumeCmd)
	return rootCmd, config
}

func Main(args []string, k8sFactory func() (K8S, error)) error {
	command, _ := newParser(args, k8sFactory, SetupLogging)

	err := command.Execute()
	if err != nil {
		return err
	}

	return nil
}
