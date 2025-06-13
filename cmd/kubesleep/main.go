package main

import (
	"fmt"
	"log/slog"
	"os"

	k8s "github.com/Y0-L0/kubesleep/k8s"
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
)

func setupLogging(logLevel slog.Level) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     logLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	setupLogging(slog.LevelDebug)
	err := kubesleep.Main(os.Args, k8s.NewK8S)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}
