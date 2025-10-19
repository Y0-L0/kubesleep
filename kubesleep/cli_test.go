package kubesleep

import (
	"log/slog"

	"github.com/stretchr/testify/mock"
)

func NewTestParser(args []string, k8sFactory K8SFactory) (*cliConfig, error) {
	command, config := NewParser(args, k8sFactory, SetupLogging)
	err := command.Execute()
	return config, err
}

func (s *Unittest) TestMainHelp() {
	_, err := NewTestParser([]string{"kubesleep", "--help"}, nil)
	s.Require().NoError(err)
}

func (s *Unittest) TestMainError() {
	_, err := NewTestParser([]string{"kubesleep", "blub"}, nil)
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
			&cliConfig{namespaces: []string{"test-ns"}, force: false, allNamespaces: false},
		},
		{
			"suspend verbose",
			[]string{"kubesleep", "suspend", "-n", "test-ns"},
			"suspend",
			&cliConfig{namespaces: []string{"test-ns"}, force: false, allNamespaces: false},
		},
		{
			"suspend multiple namespaces",
			[]string{"kubesleep", "suspend", "-n", "test-ns", "-n", "other-test-ns"},
			"suspend",
			&cliConfig{namespaces: []string{"test-ns", "other-test-ns"}, force: false, allNamespaces: false},
		},
		{
			"suspend all namespaces",
			[]string{"kubesleep", "suspend", "--all-namespaces"},
			"suspend",
			&cliConfig{namespaces: nil, force: false, allNamespaces: true},
		},
		{
			"suspend with force",
			[]string{"kubesleep", "suspend", "-n", "test-ns", "-f"},
			"suspend",
			&cliConfig{namespaces: []string{"test-ns"}, force: true, allNamespaces: false},
		},
		{
			"wake with ns",
			[]string{"kubesleep", "wake", "-n", "test-ns"},
			"wake",
			&cliConfig{namespaces: []string{"test-ns"}, force: false, allNamespaces: false},
		},
		{
			"status with ns",
			[]string{"kubesleep", "status", "-n", "test-ns"},
			"status",
			&cliConfig{namespaces: []string{"test-ns"}, force: false, allNamespaces: false},
		},
		{
			"status multiple namespaces",
			[]string{"kubesleep", "status", "-n", "test-ns", "-n", "other-test-ns"},
			"status",
			&cliConfig{namespaces: []string{"test-ns", "other-test-ns"}, force: false, allNamespaces: false},
		},
		{
			"status all namespaces",
			[]string{"kubesleep", "status", "--all-namespaces"},
			"status",
			&cliConfig{namespaces: nil, force: false, allNamespaces: true},
		},
	}

	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			k8s, factory := NewMockK8S()
			if testCase.config.allNamespaces {
				k8s.On("GetSuspendableNamespaces", mock.Anything).Return([]SuspendableNamespace{}, errExpected)
			} else {
				k8s.On("GetSuspendableNamespace", mock.Anything, mock.Anything).Return(&suspendableNamespaceImpl{}, errExpected)
			}

			command, config := NewParser(
				testCase.args,
				factory,
				SetupLogging,
			)
			err := command.Execute()
			s.Require().Equal(errExpected, err)
			k8s.AssertExpectations(s.T())

			s.Equal("kubesleep", command.Name())
			// outWriter is initialized by the parser, set expected to match
			testCase.config.outWriter = command.OutOrStdout()
			s.Equal(testCase.config, config)
		})
	}
}

func (s *Unittest) TestVersionSubcommand() {
	tests := []struct {
		name   string
		args   []string
		config *cliConfig
	}{
		{
			"print version information",
			[]string{"kubesleep", "version"},
			&cliConfig{namespaces: nil, force: false, allNamespaces: false},
		},
		{
			"print version information ignoring any namespace arguments",
			[]string{"kubesleep", "version", "-n", "test-ns", "-n", "other-test-ns"},
			&cliConfig{namespaces: []string{"test-ns", "other-test-ns"}, force: false, allNamespaces: false},
		},
	}

	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			command, config := NewParser(
				testCase.args,
				nil,
				SetupLogging,
			)
			err := command.Execute()
			s.Require().NoError(err)

			s.Equal("kubesleep", command.Name())
			testCase.config.outWriter = command.OutOrStdout()
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
		{"wake empty namespace", []string{"kubesleep", "wake", "-n", ""}, &cliConfig{namespaces: []string{""}}},
		{"suspend no namespace", []string{"kubesleep", "suspend"}, &cliConfig{}},
		{"suspend empty namespace", []string{"kubesleep", "suspend", "-n", ""}, &cliConfig{namespaces: []string{""}}},
		{"suspend no namespace force", []string{"kubesleep", "suspend", "--force"}, &cliConfig{force: true}},
		{"suspend all namespaces force", []string{"kubesleep", "suspend", "--all-namespaces", "--force"}, &cliConfig{allNamespaces: true, force: true}},
		{"suspend all namespaces namespace colision", []string{"kubesleep", "suspend", "--all-namespaces", "--namespace", "foo"}, &cliConfig{allNamespaces: true, namespaces: []string{"foo"}}},
		{"status no namespace", []string{"kubesleep", "status"}, &cliConfig{}},
		{"status empty namespace", []string{"kubesleep", "status", "-n", ""}, &cliConfig{namespaces: []string{""}}},
		{"status all namespaces namespace colision", []string{"kubesleep", "status", "--all-namespaces", "--namespace", "foo"}, &cliConfig{allNamespaces: true, namespaces: []string{"foo"}}},
		{"unknown command", []string{"kubesleep", "unknown"}, &cliConfig{}},
	}

	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			command, config := NewParser(testCase.args, nil, SetupLogging)

			err := command.Execute()

			s.Require().Error(err)
			testCase.config.outWriter = command.OutOrStdout()
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
		{"status no verbosity", []string{"kubesleep", "status", "-n", "foo"}, slog.LevelWarn},
		{"suspend some verbosity", []string{"kubesleep", "wake", "-n", "foo", "-v"}, slog.LevelInfo},
		{"wake some verbosity", []string{"kubesleep", "wake", "-n", "foo", "-v"}, slog.LevelInfo},
		{"status some verbosity", []string{"kubesleep", "status", "-n", "foo", "-v"}, slog.LevelInfo},
		{"suspend full verbosity", []string{"kubesleep", "wake", "-n", "foo", "-vv"}, slog.LevelDebug},
		{"suspend excessive verbosity", []string{"kubesleep", "wake", "-n", "foo", "-vvv"}, slog.LevelDebug},
		{"suspend different excessive verbosity", []string{"kubesleep", "wake", "-n", "foo", "-v", "-v", "-v", "-v"}, slog.LevelDebug},
	}

	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			k8s, factory := NewMockK8S()
			k8s.On("GetSuspendableNamespace", mock.Anything, mock.Anything).Return(&suspendableNamespaceImpl{}, errExpected)

			var logLevel slog.Level

			mockSetupLogging := func(l slog.Level) { logLevel = l }

			command, config := NewParser(testCase.args, factory, mockSetupLogging)
			err := command.Execute()

			s.Require().Equal(errExpected, err)
			k8s.AssertExpectations(s.T())
			expected := &cliConfig{namespaces: []string{"foo"}}
			expected.outWriter = command.OutOrStdout()
			s.Require().Equal(expected, config)
			s.Require().Equal(testCase.logLevel, logLevel)
		})
	}
}

func (s *Unittest) TestCliArgumentError_Error() {
	e := CliArgumentError("simple error")
	s.Require().Equal("simple error", e.Error())
}
