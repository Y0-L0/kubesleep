package kubesleep

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (s *Unittest) TestMainHelp() {
	err := Main([]string{"kubesleep", "--help"}, nil)
	s.Require().NoError(err)
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
			command, config := newParser(testCase.args, nil)

			actualSubcommand := map[string]int{"suspend": 0, "wake": 0}
			for _, subCmd := range command.Commands() {
				switch subCmd.Name() {
				case "suspend":
					subCmd.RunE = func(cmd *cobra.Command, args []string) error {
						actualSubcommand["suspend"] = 1
						return nil
					}
				case "wake":
					subCmd.RunE = func(cmd *cobra.Command, args []string) error {
						actualSubcommand["wake"] = 1
						return nil
					}
				default:
					panic(fmt.Errorf("unknown subcommand: %s", subCmd.Name()))
				}
			}

			err := command.Execute()
			s.Require().NoError(err)

			s.Equal("kubesleep", command.Name())
			s.Equal(1, actualSubcommand[testCase.command])
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
