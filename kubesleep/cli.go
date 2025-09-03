package kubesleep

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/Y0-L0/kubesleep/kubesleep/version"
	"github.com/spf13/cobra"
)

type CliArgumentError string

func (e CliArgumentError) Error() string { return string(e) }

func SetupLogging(logLevel slog.Level) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     logLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func NewParser(args []string, k8sFactory func() (K8S, error), setupLogging func(slog.Level)) (*cobra.Command, *cliConfig) {
	slog.Debug("raw cli arguments", "args", args)

	config := &cliConfig{}
	var verbosity int

	rootCmd := &cobra.Command{
		Use:           "kubesleep",
		Short:         "kubesleep can sleep and wake kubernetes namespaces by scaling workloads down to zero and back up",
		SilenceErrors: true,
		SilenceUsage:  true,
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
	// Direct human-facing output through the command's stdout writer
	config.outWriter = rootCmd.OutOrStdout()

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			version.BuildInfo(cmd.OutOrStdout())
		},
	}

	rootCmd.PersistentFlags().CountVarP(
		&verbosity,
		"verbose",
		"v",
		"Increase the log level. Can be specified multiple times.",
	)
	rootCmd.PersistentFlags().StringArrayVarP(
		&config.namespaces,
		"namespace",
		"n",
		nil,
		"Kubernetes namespace. Can be specified multiple times",
	)

	suspendCmd := &cobra.Command{
		Use:   "suspend",
		Short: "Suspend one or multiple kubernetes namespaces.",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Parsed cli arguments for the suspend subcommand", "config", config)
			if !config.allNamespaces && len(config.namespaces) == 0 {
				cmd.PrintErrln("either --all-namespaces or --namespace (-n) must be specified")
				return CliArgumentError("missing namespace or all-namespaces argument")
			}
			if config.allNamespaces && (len(config.namespaces) > 0 || config.force) {
				cmd.PrintErrln("--all-namespaces cannot be combined with --namespace or --force")
				return CliArgumentError("missing namespace or all-namespaces argument")
			}
			return config.suspend(cmd.Context(), k8sFactory)
		},
	}
	suspendCmd.Flags().BoolVarP(
		&config.force,
		"force",
		"f",
		false,
		"Ignore the do-not-suspend label on the namespace",
	)
	suspendCmd.Flags().BoolVar(
		&config.allNamespaces,
		"all-namespaces",
		false,
		"Suspend all unprotected namespaces",
	)

	wakeCmd := &cobra.Command{
		Use:   "wake",
		Short: "Wake a kubernetes namespace back up",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Parsed cli arguments for the wake subcommand", "config", config)
			if len(config.namespaces) == 0 {
				cmd.PrintErrln("--namespace (-n) must be specified")
				return CliArgumentError("missing namespace argument")
			}
			return config.wake(cmd.Context(), k8sFactory)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Display the suspend status of kubernetes namespaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Parsed cli arguments for the status subcommand", "config", config)
			if !config.allNamespaces && len(config.namespaces) == 0 {
				cmd.PrintErrln("either --all-namespaces or --namespace (-n) must be specified")
				return CliArgumentError("missing namespace or all-namespaces argument")
			}
			statusTable, err := config.status(cmd.Context(), k8sFactory)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\n", err)
				return err
			}
			printStatus(statusTable, cmd.OutOrStdout())
			return nil
		},
	}
	statusCmd.Flags().BoolVar(
		&config.allNamespaces,
		"all-namespaces",
		false,
		"Suspend all unprotected namespaces",
	)

	rootCmd.AddCommand(versionCmd, suspendCmd, wakeCmd, statusCmd)
	return rootCmd, config
}

func printStatus(statusTable []status, stdout io.Writer) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "name\tstatus\tprotected\t")
	for _, row := range statusTable {
		fmt.Fprintf(w, "%s\t%s\t%t\t\n", row.name, row.status, row.protected)
	}
	w.Flush()
}
