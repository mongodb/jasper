package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"syscall"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
)

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
