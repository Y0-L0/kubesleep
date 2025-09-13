package main

import (
	"log/slog"
	"os"

	k8s "github.com/Y0-L0/kubesleep/k8s"
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
)

func main() {
	kubesleep.Main(
		os.Args,
		slog.LevelWarn,
		k8s.NewK8S,
		os.Stdout,
		os.Stderr,
	)
}
