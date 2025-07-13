package kubesleep

import "github.com/stretchr/testify/mock"

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
			)
			err := command.Execute()
			s.Require().Equal(errExpected, err)
			k8s.AssertExpectations(s.T())

			s.Equal("kubesleep", command.Name())
			s.Equal(testCase.config, config)
		})
	}
}

// Consolidated error state tests
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
			command, config := newParser(testCase.args, nil)
			err := command.Execute()
			s.Require().Error(err)
			s.Require().Equal(testCase.config, config)
		})
	}
}
