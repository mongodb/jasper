package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"syscall"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
)

// ClientOptions represents the options to connect the CLI client to the Jasper
// service.
type ClientOptions struct {
	CLIPath string
	Type    string
	Host    string
	Port    int
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
		opts.Client.CLIPath,
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

// sshManager uses SSH to access a remote machine's Jasper CLI, which has access to
// methods in the Manager interface.
type sshManager struct {
	opts sshClientOptions
}

// NewSSHManager creates a new Jasper manager that connects to a remote
// machine's Jasper service over SSH using the remote machine's Jasper CLI.
func NewSSHManager(ctx context.Context, remoteOpts jasper.RemoteOptions, clientOpts ClientOptions) (jasper.Manager, error) {
	m := sshManager{
		opts: sshClientOptions{
			Machine: remoteOpts,
			Client:  clientOpts,
		},
	}
	return &m, nil
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
	return errors.New("cannot register existing processes on remote process manager")
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

// sshProcess uses SSH to access a remote machine's Jasper CLI, which has access
// to methods in the Process interface.
type sshProcess struct {
	opts sshClientOptions
	info jasper.ProcessInfo
}

func (p *sshProcess) runCommand(ctx context.Context, processSubcommand string, subcommandInput interface{}) ([]byte, error) {
	var input io.Reader
	if subcommandInput != nil {
		inputBytes, err := json.MarshalIndent(subcommandInput, "", "    ")
		if err != nil {
			return nil, errors.Wrap(err, "could not encode input")
		}
		input = bytes.NewBuffer(inputBytes)
	}
	output := sshOutput()
	cmdArgs := []string{ProcessCommand, processSubcommand}

	cmd := p.opts.newCommand(cmdArgs, input, io.WriteCloser(output))
	if err := cmd.Run(ctx); err != nil {
		return nil, errors.Wrap(err, "problem running command over SSH")
	}

	return output.Bytes(), nil
}

func (p *sshProcess) ID() string {
	return p.info.ID
}

func (p *sshProcess) Info(ctx context.Context) jasper.ProcessInfo {
	if p.info.Complete {
		return p.info
	}

	output, err := p.runCommand(ctx, InfoCommand, nil)
	if err != nil {
		return jasper.ProcessInfo{}
	}
	resp, err := ExtractInfoResponse(output)
	if err != nil {
		return jasper.ProcessInfo{}
	}
	p.info = resp.Info

	return p.info
}

func (p *sshProcess) Running(ctx context.Context) bool {
	if p.info.Complete {
		return false
	}

	output, err := p.runCommand(ctx, RunningCommand, &IDInput{ID: p.info.ID})
	if err != nil {
		return false
	}
	resp, err := ExtractRunningResponse(output)
	if err != nil {
		return false
	}
	p.info.IsRunning = resp.Running

	return p.info.IsRunning
}

func (p *sshProcess) Complete(ctx context.Context) bool {
	if p.info.Complete {
		return true
	}

	output, err := p.runCommand(ctx, CompleteCommand, &IDInput{ID: p.info.ID})
	if err != nil {
		return false
	}
	resp, err := ExtractCompleteResponse(output)
	if err != nil {
		return false
	}
	p.info.Complete = resp.Complete

	return p.info.Complete
}

func (p *sshProcess) Signal(ctx context.Context, sig syscall.Signal) error {
	output, err := p.runCommand(ctx, SignalCommand, &SignalInput{ID: p.info.ID, Signal: int(sig)})
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err = ExtractOutcomeResponse(output); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *sshProcess) Wait(ctx context.Context) (int, error) {
	output, err := p.runCommand(ctx, WaitCommand, &IDInput{ID: p.info.ID})
	if err != nil {
		return -1, errors.WithStack(err)
	}
	resp, err := ExtractWaitResponse(output)
	if err != nil {
		return -1, errors.WithStack(err)
	}
	return resp.ExitCode, nil
}

func (p *sshProcess) Respawn(ctx context.Context) (jasper.Process, error) {
	output, err := p.runCommand(ctx, RespawnCommand, &IDInput{ID: p.info.ID})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := ExtractInfoResponse(output)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &sshProcess{
		opts: p.opts,
		info: resp.Info,
	}, nil
}

func (p *sshProcess) RegisterTrigger(ctx context.Context, t jasper.ProcessTrigger) error {
	return errors.New("cannot register triggers on remote processes")
}

func (p *sshProcess) RegisterSignalTrigger(ctx context.Context, t jasper.SignalTrigger) error {
	return errors.New("cannot register signal triggers on remote processes")
}

func (p *sshProcess) RegisterSignalTriggerID(ctx context.Context, sigID jasper.SignalTriggerID) error {
	output, err := p.runCommand(ctx, RegisterSignalTriggerIDCommand, &SignalTriggerIDInput{
		ID:              p.info.ID,
		SignalTriggerID: sigID,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err = ExtractOutcomeResponse(output); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *sshProcess) Tag(tag string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _ = p.runCommand(ctx, TagCommand, &TagIDInput{
		ID:  p.info.ID,
		Tag: tag,
	})
}

func (p *sshProcess) GetTags() []string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	output, err := p.runCommand(ctx, GetTagsCommand, &IDInput{
		ID: p.info.ID,
	})
	if err != nil {
		return nil
	}
	resp, err := ExtractTagsResponse(output)
	if err != nil {
		return nil
	}
	return resp.Tags
}

func (p *sshProcess) ResetTags() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _ = p.runCommand(ctx, ResetTagsCommand, &IDInput{
		ID: p.info.ID,
	})
}

func sshOutput() *jasper.CappedWriter {
	return &jasper.CappedWriter{
		Buffer:   &bytes.Buffer{},
		MaxBytes: 1024 * 1024, // 1 MB
	}
}
