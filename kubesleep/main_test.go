package kubesleep

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"

	"github.com/stretchr/testify/mock"
)

func (s *Unittest) TestMain_Success() {
	_, factory := NewMockK8S()

	var outWriter, errWriter bytes.Buffer
	mockVersionUpdateCheck := func(_ *http.Client) (string, error) {
		return "", nil
	}
	code := Main(
		[]string{"kubesleep", "version"},
		slog.LevelInfo,
		factory,
		mockVersionUpdateCheck,
		&outWriter,
		&errWriter,
	)
	s.Require().Equal(0, code)
	s.Require().Contains(outWriter.String(), "Version:")
	s.Require().Equal("", errWriter.String())
}

func (s *Unittest) TestMain_UpdateCheck_IgnoreError() {
	_, factory := NewMockK8S()

	var outWriter, errWriter bytes.Buffer
	mockVersionUpdateCheck := func(_ *http.Client) (string, error) {
		return "updated!", errExpected
	}
	code := Main(
		[]string{"kubesleep", "version"},
		slog.LevelInfo,
		factory,
		mockVersionUpdateCheck,
		&outWriter,
		&errWriter,
	)
	s.Require().Equal(0, code)
	s.Require().Contains(outWriter.String(), "Version:")
	s.Require().NotContains(outWriter.String(), "updated!")
	s.Require().Equal("", errWriter.String())
}

func (s *Unittest) TestMain_UpdateCheck_NonBlocking() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", mock.Anything).Return(&suspendableNamespaceImpl{}, errExpected)

	var outWriter, errWriter bytes.Buffer
	release := make(chan struct{})
	mockVersionUpdateCheck := func(_ *http.Client) (string, error) {
		<-release
		return "update!", nil
	}

	start := time.Now()
	code := Main(
		[]string{"kubesleep", "suspend", "-n", "blub"},
		slog.LevelInfo,
		factory,
		mockVersionUpdateCheck,
		&outWriter,
		&errWriter,
	)
	elapsed := time.Since(start)

	s.Require().Equal(1, code)
	s.Require().Less(elapsed, 100*time.Millisecond)
	s.Require().Equal("", outWriter.String())
	s.Require().Equal("broken k8s factory\n", errWriter.String())

	close(release)
}

func (s *Unittest) TestMain_PrintsUpdateWhenReady() {
	k8s, factory := NewMockK8S()
	updateDone := make(chan struct{})
	mockVersionUpdateCheck := func(_ *http.Client) (string, error) {
		close(updateDone)
		return "update!", nil
	}

	// Hold command execution until the update goroutine signaled readiness
	k8s.On("GetSuspendableNamespace", mock.Anything).
		Run(func(_ mock.Arguments) { <-updateDone }).
		Return(&suspendableNamespaceImpl{}, errExpected)

	var outWriter, errWriter bytes.Buffer
	code := Main(
		[]string{"kubesleep", "suspend", "-n", "blub"},
		slog.LevelInfo,
		factory,
		mockVersionUpdateCheck,
		&outWriter,
		&errWriter,
	)

	s.Require().Equal(1, code)
	s.Require().Equal("\nupdate!\n", outWriter.String())
	s.Require().Equal("broken k8s factory\n", errWriter.String())
}
