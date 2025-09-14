package kubesleep

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
)

func Main(
	args []string,
	initialLogLevel slog.Level,
	k8sFactory func() (K8S, error),
	versionUpdateCheck func(*http.Client) (string, error),
	outWriter io.Writer,
	errWriter io.Writer,
) int {
	SetupLogging(slog.LevelInfo)

	var updateMsg atomic.Value
	updateMsg.Store("")
	go func() {
		msg, err := versionUpdateCheck(&http.Client{})
		if err != nil {
			slog.Info("Update check failed", "error", err)
			return
		}
		if msg != "" {
			updateMsg.Store(msg)
		}
	}()

	command, _ := NewParser(args, k8sFactory, SetupLogging)
	command.SetOut(outWriter)
	command.SetErr(errWriter)

	commandErr := command.Execute()

	updateMessage := updateMsg.Load().(string)
	if updateMessage != "" {
		fmt.Fprintln(outWriter)
		fmt.Fprintln(outWriter, updateMessage)
	}

	if commandErr != nil {
		fmt.Fprintln(errWriter, commandErr)
		return 1
	}
	return 0
}
