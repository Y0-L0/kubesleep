package kubesleep

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LoggingSuite struct {
	suite.Suite
	logBuf    bytes.Buffer
	oldLogger *slog.Logger
}

func (ls *LoggingSuite) SetupTest() {
	ls.logBuf.Reset()

	handler := slog.NewTextHandler(&ls.logBuf, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func (loggingSuite *LoggingSuite) TearDownTest() {
	if !loggingSuite.T().Failed() || !testing.Verbose() {
		return
	}
	loggingSuite.T().Log("=== Captured Production Logs ===\n")
	loggingSuite.T().Log(loggingSuite.logBuf.String())
}
