package version

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime"

	"golang.org/x/mod/semver"
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

func CheckForUpdate() (string, error) {
	if !semver.IsValid(Version) {
		return "", fmt.Errorf("invalid build version string: %s", Version)
	}
	response, err := http.Get(latestVersionUrl)
	if err != nil {
		return "", fmt.Errorf("failed update check http request: %w", err)
	}

	var payload struct {
		Version string `json:"tag_name"`
		Url     string `json:"html_url"`
	}
	err = json.NewDecoder(response.Body).Decode(&payload)
	if err != nil {
		return "", fmt.Errorf("failed to parse github release json: %w", err)
	}
	if !semver.IsValid(payload.Version) {
		return "", fmt.Errorf("invalid remote version string: %s", payload.Version)
	}
	if semver.Compare(Version, payload.Version) >= 0 {
		slog.Info("No new kubesleep version available on github", "build-version", Version, "latest-version", payload.Version)
		return "", nil
	}
	return fmt.Sprintf("A new kubesleep version has been released: %s -> %s %s", Version, payload.Version, payload.Url), nil
}
