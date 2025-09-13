package kubesleep

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Y0-L0/kubesleep/kubesleep/version"
)

func Main(args []string, initialLogLevel slog.Level, k8sFactory func() (K8S, error), outWriter, errWriter io.Writer) {
	SetupLogging(slog.LevelInfo)

	updateCh := make(chan string)
	go func() {
		msg, err := version.CheckForUpdate()
		if err != nil {
			slog.Info("Update check failed", "error", err)
		}
		updateCh <- msg
	}()

	command, _ := NewParser(args, k8sFactory, SetupLogging)
	command.SetOut(outWriter)
	command.SetErr(errWriter)

	commandErr := command.Execute()

	updateMessage := <-updateCh
	if updateMessage != "" {
		fmt.Fprintln(outWriter)
		fmt.Fprintln(outWriter, updateMessage)
	}

	if commandErr != nil {
		os.Exit(1)
	}
}
