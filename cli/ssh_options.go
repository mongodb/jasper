package cli

import (
	"fmt"
	"io"

	"github.com/mongodb/grip"
	"github.com/mongodb/jasper"
)

// ClientOptions represents the options to connect the CLI client to the Jasper
// service.
type ClientOptions struct {
	BinaryPath string
	Type       string
	Host       string
	Port       int
}

// Validate checks that the binary path is set and it is a recognized Jasper
// client type.
func (opts *ClientOptions) Validate() error {
	catcher := grip.NewBasicCatcher()
	if opts.BinaryPath == "" {
		catcher.New("client binary path cannot be empty")
	}
	if opts.Type != RPCService && opts.Type != RESTService {
		catcher.New("client type must be RPC or REST")
	}
	return catcher.Resolve()
}

// sshClientOptions represents the options necessary to run a Jasper CLI
// command over SSH.
type sshClientOptions struct {
	Machine jasper.RemoteOptions
	Client  ClientOptions
}

// newCommand creates a command that invokes the Jasper client CLI over SSH with
// the given arguments.
func (opts *sshClientOptions) newCommand(clientSubcommand []string, input io.Reader, output io.WriteCloser) *jasper.Command {
	cmd := jasper.NewCommand().Host(opts.Machine.Host).User(opts.Machine.User).ExtendRemoteArgs(opts.Machine.Args...).
		Add(opts.args(clientSubcommand))
	if input != nil {
		cmd.SetInput(input)
	}
	if output != nil {
		cmd.SetCombinedWriter(output)
	}

	return cmd
}

// args returns the Jasper CLI command that will be run over SSH.
func (opts *sshClientOptions) args(clientSubcommand []string) []string {
	args := append([]string{
		opts.Client.BinaryPath,
		JasperCommand,
		ClientCommand},
		clientSubcommand...,
	)
	args = append(args, fmt.Sprintf("--%s=%s", serviceFlagName, opts.Client.Type))

	if opts.Client.Host != "" {
		args = append(args, fmt.Sprintf("--%s=%s", hostFlagName, opts.Client.Host))
	}

	if opts.Client.Port != 0 {
		args = append(args, fmt.Sprintf("--%s=%d", portFlagName, opts.Client.Port))
	}

	return args
}
