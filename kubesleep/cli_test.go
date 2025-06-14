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
			"suspend no ns",
			[]string{"kubesleep", "suspend"},
			"suspend",
			&cliConfig{"", false},
		},
		{
			"suspend with ns",
			[]string{"kubesleep", "suspend", "-n", "test-ns"},
			"suspend",
			&cliConfig{"test-ns", false},
		},
		{
			"suspend with force",
			[]string{"kubesleep", "suspend", "-n", "test-ns", "-f"},
			"suspend",
			&cliConfig{"test-ns", true},
		},
		{
			"wake with ns",
			[]string{"kubesleep", "wake", "-n", "test-ns"},
			"wake",
			&cliConfig{"test-ns", false},
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
		name string
		args []string
	}{
		{"wake no namespace", []string{"kubesleep", "wake"}},
		{"unknown command", []string{"kubesleep", "unknown"}},
	}

	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			command, config := newParser(testCase.args, nil)
			err := command.Execute()
			s.Require().Error(err)
			s.Require().Equal(&cliConfig{}, config)
		})
	}
}
