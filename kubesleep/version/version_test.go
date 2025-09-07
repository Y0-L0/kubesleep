package version

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildInfoOutput(t *testing.T) {
	Version = "v1.2.3"
	Commit = "abc123"
	TreeState = "clean"
	BuildDate = "2024-01-02T03:04:05Z"

	var buf bytes.Buffer
	BuildInfo(&buf)

	expected := "" +
		"Version:    v1.2.3\n" +
		"Commit:     abc123\n" +
		"TreeState:  clean\n" +
		"BuildDate:  2024-01-02T03:04:05Z\n" +
		"GoVersion:  " + runtime.Version() + "\n" +
		"Platform:   " + runtime.GOOS + "/" + runtime.GOARCH + "\n"

	require.Equal(t, expected, buf.String())
}
