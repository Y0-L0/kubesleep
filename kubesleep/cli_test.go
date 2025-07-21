package kubesleep

import (
	"log/slog"

	"github.com/stretchr/testify/mock"
)

func (s *Unittest) TestMainHelp() {
	err := Main([]string{"kubesleep", "--help"}, nil)
	s.Require().NoError(err)
}

func (s *Unittest) TestMainError() {
	err := Main([]string{"kubesleep", "invalidArgument"}, nil)
	s.Require().Error(err)
}

func (s *Unittest) TestValidCliArguments() {
	tests := []struct {
		name    string
		args    []string
		command string
		config  *cliConfig
	}{
		{
			"suspend with ns",
			[]string{"kubesleep", "suspend", "-n", "test-ns"},
			"suspend",
			&cliConfig{[]string{"test-ns"}, false, false},
		},
		{
			"suspend verbose",
			[]string{"kubesleep", "suspend", "-n", "test-ns"},
			"suspend",
			&cliConfig{[]string{"test-ns"}, false, false},
		},
		{
			"suspend multiple namespaces",
			[]string{"kubesleep", "suspend", "-n", "test-ns", "-n", "other-test-ns"},
			"suspend",
			&cliConfig{[]string{"test-ns", "other-test-ns"}, false, false},
		},
		// TODO: make the --all-namespaces argument functional
		// {
		// 	"suspend all namespaces",
		// 	[]string{"kubesleep", "suspend", "--all-namespaces"},
		// 	"suspend",
		// 	&cliConfig{nil, false, true},
		// },
		{
			"suspend with force",
			[]string{"kubesleep", "suspend", "-n", "test-ns", "-f"},
			"suspend",
			&cliConfig{[]string{"test-ns"}, true, false},
		},
		{
			"wake with ns",
			[]string{"kubesleep", "wake", "-n", "test-ns"},
			"wake",
			&cliConfig{[]string{"test-ns"}, false, false},
		},
	}

	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			k8s, factory := NewMockK8S()
			k8s.On("GetSuspendableNamespace", mock.Anything).Return(&suspendableNamespaceImpl{}, errExpected)

			command, config := newParser(
				testCase.args,
				factory,
				SetupLogging,
			)
			err := command.Execute()
			s.Require().Equal(errExpected, err)
			k8s.AssertExpectations(s.T())

			s.Equal("kubesleep", command.Name())
			s.Equal(testCase.config, config)
		})
	}
}

func (s *Unittest) TestInvalidCliArguments() {
	tests := []struct {
		name   string
		args   []string
		config *cliConfig
	}{
		{"wake no namespace", []string{"kubesleep", "wake"}, &cliConfig{}},
		{"suspend no namespace", []string{"kubesleep", "suspend"}, &cliConfig{}},
		{"suspend no namespace force", []string{"kubesleep", "suspend", "--force"}, &cliConfig{force: true}},
		{"suspend all namespaces force", []string{"kubesleep", "suspend", "--all-namespaces", "--force"}, &cliConfig{allNamespaces: true, force: true}},
		{"suspend all namespaces namespace colision", []string{"kubesleep", "suspend", "--all-namespaces", "--namespace", "foo"}, &cliConfig{allNamespaces: true, namespaces: []string{"foo"}}},
		{"unknown command", []string{"kubesleep", "unknown"}, &cliConfig{}},
	}

	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			command, config := newParser(testCase.args, nil, SetupLogging)

			err := command.Execute()

			s.Require().Error(err)
			s.Require().Equal(testCase.config, config)
		})
	}
}

func (s *Unittest) TestLogLevel() {
	tests := []struct {
		name     string
		args     []string
		logLevel slog.Level
	}{
		{"suspend no verbosity", []string{"kubesleep", "suspend", "-n", "foo"}, slog.LevelWarn},
		{"wake no verbosity", []string{"kubesleep", "wake", "-n", "foo"}, slog.LevelWarn},
		{"suspend some verbosity", []string{"kubesleep", "wake", "-n", "foo", "-v"}, slog.LevelInfo},
		{"wake some verbosity", []string{"kubesleep", "wake", "-n", "foo", "-v"}, slog.LevelInfo},
		{"suspend full verbosity", []string{"kubesleep", "wake", "-n", "foo", "-vv"}, slog.LevelDebug},
		{"suspend excessive verbosity", []string{"kubesleep", "wake", "-n", "foo", "-vvv"}, slog.LevelDebug},
		{"suspend different excessive verbosity", []string{"kubesleep", "wake", "-n", "foo", "-v", "-v", "-v", "-v"}, slog.LevelDebug},
	}

	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			k8s, factory := NewMockK8S()
			k8s.On("GetSuspendableNamespace", mock.Anything).Return(&suspendableNamespaceImpl{}, errExpected)

			var logLevel slog.Level

			mockSetupLogging := func(l slog.Level) { logLevel = l }

			command, config := newParser(testCase.args, factory, mockSetupLogging)
			err := command.Execute()

			s.Require().Equal(errExpected, err)
			k8s.AssertExpectations(s.T())
			s.Require().Equal(&cliConfig{namespaces: []string{"foo"}}, config)
			s.Require().Equal(testCase.logLevel, logLevel)
		})
	}
}
