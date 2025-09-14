package version

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LoggingSuite struct {
	suite.Suite
	logBuf     bytes.Buffer
	oldLogger  *slog.Logger
	oldVersion string
	oldURL     string
}

func (ls *LoggingSuite) SetupTest() {
	ls.oldVersion = Version
	ls.oldURL = LatestVersionURL
	ls.logBuf.Reset()

	handler := slog.NewTextHandler(&ls.logBuf, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func (ls *LoggingSuite) TearDownTest() {
	Version = ls.oldVersion
	LatestVersionURL = ls.oldURL
	if !ls.T().Failed() || !testing.Verbose() {
		return
	}
	ls.T().Log("=== Captured Production Logs ===\n")
	ls.T().Log(ls.logBuf.String())
}

type Unittest struct {
	LoggingSuite
}

func TestUnit(t *testing.T) {
	suite.Run(t, new(Unittest))
}
