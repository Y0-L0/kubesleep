package main

import (
	"fmt"
	"log/slog"
	"os"

	kubesleep "github.com/Y0-L0/kubesleep/kube-sleep"
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
	err := kubesleep.Main(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}
