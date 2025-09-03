package main

import (
	"log/slog"
	"os"

	k8s "github.com/Y0-L0/kubesleep/k8s"
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
)

func main() {
	kubesleep.SetupLogging(slog.LevelWarn)

	command, _ := kubesleep.NewParser(os.Args, k8s.NewK8S, kubesleep.SetupLogging)
	command.SetOut(os.Stdout)
	command.SetErr(os.Stderr)

	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
