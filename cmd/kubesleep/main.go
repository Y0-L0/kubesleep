package main

import (
	"log/slog"
	"os"

	k8s "github.com/Y0-L0/kubesleep/k8s"
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	"github.com/Y0-L0/kubesleep/kubesleep/version"
)

func main() {
	os.Exit(kubesleep.Main(
		os.Args,
		slog.LevelWarn,
		k8s.NewK8S,
		version.CheckForUpdate,
		os.Stdout,
		os.Stderr,
	))
}
