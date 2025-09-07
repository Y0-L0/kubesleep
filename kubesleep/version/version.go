package version

import (
	"fmt"
	"io"
	"runtime"
)

var (
	Version          = "v0.2.0"
	Commit           = "none"
	TreeState        = "unknown"
	BuildDate        = "unknown" // RFC3339
	latestVersionUrl = "https://api.github.com/repos/y0-l0/kubesleep/releases/latest"
)

func BuildInfo(w io.Writer) {
	fmt.Fprintf(w, "Version:    %s\n", Version)
	fmt.Fprintf(w, "Commit:     %s\n", Commit)
	fmt.Fprintf(w, "TreeState:  %s\n", TreeState)
	fmt.Fprintf(w, "BuildDate:  %s\n", BuildDate)
	fmt.Fprintf(w, "GoVersion:  %s\n", runtime.Version())
	fmt.Fprintf(w, "Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
