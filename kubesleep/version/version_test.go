package version

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
)

func (s *Unittest) TestBuildInfoOutput() {
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

	s.Require().Equal(expected, buf.String())
}

func (s *Unittest) TestCheckForUpdate_NewerAvailable() {
	message, err := compareVersions("v0.2.0", "v0.3.0", "https://example.com")

	s.Require().NoError(err)
	s.Require().Equal("A new kubesleep version has been released: v0.2.0 -> v0.3.0 https://example.com", message)
}

func (s *Unittest) TestCheckForUpdate_UpToDate() {
	message, err := compareVersions("v0.3.0", "v0.3.0", "https://example.com")

	s.Require().NoError(err)
	s.Require().Equal(message, "")
}

func (s *Unittest) TestCheckForUpdate_InvalidRemoteVersion() {
	message, err := compareVersions("v0.2.0", "invalid version", "https://example.com")

	s.Require().Error(err)
	s.Require().Equal("", message)
}

func (s *Unittest) TestCheckForUpdate_InvalidLocalVersion() {
	Version = "not-semver"

	client := &http.Client{}
	message, err := CheckForUpdate(client)
	s.Require().Error(err)
	s.Require().Equal("", message)
}

func (s *Unittest) TestCheckForUpdate_SnapshotVersion() {
	message, err := compareVersions("v0.3.4-SNAPSHOT-421ef99", "v0.3.4", "https://example.com")

	s.Require().NoError(err)
	s.Require().Contains(message, "A new kubesleep version has been released:")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newMockClient(status int, body string) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if status == 0 {
				return nil, errors.New("simulated http error")
			}
			rec := httptest.NewRecorder()
			rec.WriteHeader(status)
			rec.Body.WriteString(body)
			return rec.Result(), nil
		}),
	}
}

func (s *Unittest) TestCheckForUpdate_HTTPErrorResponse() {
	Version = "v0.2.0"
	client := newMockClient(0, "")
	message, err := CheckForUpdate(client)
	s.Require().Error(err)
	s.Require().Equal("", message)
}

func (s *Unittest) TestCheckForUpdate_InvalidJSON() {
	Version = "v0.2.0"
	client := newMockClient(200, "not json")
	message, err := CheckForUpdate(client)
	s.Require().Error(err)
	s.Require().Equal("", message)
}

func (s *Unittest) TestCheckForUpdate_ValidResponse() {
	Version = "v0.2.0"
	body := `{"tag_name":"v0.3.0","html_url":"https://example.com/release/v0.3.0"}`
	client := newMockClient(200, body)
	message, err := CheckForUpdate(client)
	s.Require().NoError(err)
	s.Require().Equal("A new kubesleep version has been released: v0.2.0 -> v0.3.0 https://example.com/release/v0.3.0", message)
}
