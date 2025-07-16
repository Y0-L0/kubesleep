package main

import (
	"fmt"
	"log/slog"
	"os"

	k8s "github.com/Y0-L0/kubesleep/k8s"
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
)

func main() {
	kubesleep.SetupLogging(slog.LevelWarn)
	err := kubesleep.Main(os.Args, k8s.NewK8S)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}
