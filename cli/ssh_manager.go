package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
)

// sshManager uses SSH to access a remote machine's Jasper CLI, which has access to
// methods in the Manager interface.
type sshManager struct {
	opts sshClientOptions
}

// NewSSHManager creates a new Jasper manager that connects to a remote
// machine's Jasper service over SSH using the remote machine's Jasper CLI.
func NewSSHManager(remoteOpts jasper.RemoteOptions, clientOpts ClientOptions) (jasper.Manager, error) {
	if err := remoteOpts.Validate(); err != nil {
		return nil, errors.Wrap(err, "problem validating remote options")
	}
	if err := clientOpts.Validate(); err != nil {
		return nil, errors.Wrap(err, "problem validating client options")
	}
	m := sshManager{
		opts: sshClientOptions{
			Machine: remoteOpts,
			Client:  clientOpts,
		},
	}
	return &m, nil
}

func (m *sshManager) CreateProcess(ctx context.Context, opts *jasper.CreateOptions) (jasper.Process, error) {
	output, err := m.runCommand(ctx, CreateProcessCommand, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := ExtractInfoResponse(output)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &sshProcess{
		opts: m.opts,
		info: resp.Info,
	}, nil
}

func (m *sshManager) CreateCommand(ctx context.Context) *jasper.Command {
	return jasper.NewCommand().ProcConstructor(m.CreateProcess)
}

func (m *sshManager) Register(ctx context.Context, proc jasper.Process) error {
	return errors.New("cannot register existing processes on remote manager")
}

func (m *sshManager) List(ctx context.Context, f jasper.Filter) ([]jasper.Process, error) {
	output, err := m.runCommand(ctx, ListCommand, &FilterInput{Filter: f})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := ExtractInfosResponse(output)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	procs := make([]jasper.Process, 0, len(resp.Infos))
	for i := range resp.Infos {
		procs[i] = &sshProcess{
			opts: m.opts,
			info: resp.Infos[i],
		}
	}
	return procs, nil
}

func (m *sshManager) Group(ctx context.Context, tag string) ([]jasper.Process, error) {
	output, err := m.runCommand(ctx, GroupCommand, &TagInput{Tag: tag})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := ExtractInfosResponse(output)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	procs := make([]jasper.Process, 0, len(resp.Infos))
	for i := range resp.Infos {
		procs[i] = &sshProcess{
			opts: m.opts,
			info: resp.Infos[i],
		}
	}
	return procs, nil
}

func (m *sshManager) Get(ctx context.Context, id string) (jasper.Process, error) {
	output, err := m.runCommand(ctx, GetCommand, &IDInput{ID: id})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := ExtractInfoResponse(output)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &sshProcess{
		opts: m.opts,
		info: resp.Info,
	}, nil
}

func (m *sshManager) Clear(ctx context.Context) {
	_, _ = m.runCommand(ctx, ClearCommand, nil)
}

func (m *sshManager) Close(ctx context.Context) error {
	output, err := m.runCommand(ctx, CloseCommand, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err = ExtractOutcomeResponse(output); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (m *sshManager) runCommand(ctx context.Context, managerSubcommand string, subcommandInput interface{}) ([]byte, error) {
	var input io.Reader
	if subcommandInput != nil {
		inputBytes, err := json.MarshalIndent(subcommandInput, "", "    ")
		if err != nil {
			return nil, errors.Wrap(err, "could not encode input")
		}
		input = bytes.NewBuffer(inputBytes)
	}
	output := sshOutput()
	subcommand := []string{ManagerCommand, managerSubcommand}

	cmd := m.opts.newCommand(subcommand, input, io.WriteCloser(output))
	if err := cmd.Run(ctx); err != nil {
		return nil, errors.Wrapf(err, "problem running command '%s' over SSH", m.opts.args(subcommand))
	}

	return output.Bytes(), nil
}

func sshOutput() *CappedWriter {
	return &CappedWriter{
		Buffer:   &bytes.Buffer{},
		MaxBytes: 1024 * 1024, // 1 MB
	}
}
