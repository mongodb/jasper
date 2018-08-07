package jasper

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type localProcess struct {
	proc  Process
	mutex sync.RWMutex
}

func (p *localProcess) ID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.proc.ID()
}

func (p *localProcess) Info(ctx context.Context) ProcessInfo {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.proc.Info(ctx)
}

func (p *localProcess) Running(ctx context.Context) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.proc.Running(ctx)
}

func (p *localProcess) Complete(ctx context.Context) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.proc.Complete(ctx)
}

func (p *localProcess) Signal(ctx context.Context, sig syscall.Signal) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return errors.WithStack(p.proc.Signal(ctx, sig))
}

func (p *localProcess) Wait(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return errors.WithStack(p.proc.Wait(ctx))
}

type basicProcess struct {
	id       string
	hostname string
	opts     CreateOptions
	cmd      *exec.Cmd
}

func newBasicProcess(ctx context.Context, opts *CreateOptions) (Process, error) {
	hn, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "problem finding hostname when creating process")
	}

	cmd, err := opts.Resolve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "problem building command from options")
	}

	if err = cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "problem starting command")
	}

	return &basicProcess{
		hostname: hn,
		id:       uuid.Must(uuid.NewV4()).String(),
		opts:     *opts,
		cmd:      cmd,
	}, nil
}

func (p *basicProcess) ID() string { return p.id }
func (p *basicProcess) Info(ctx context.Context) ProcessInfo {
	info := ProcessInfo{
		ID:        p.id,
		Options:   p.opts,
		Host:      p.hostname,
		Complete:  p.Complete(ctx),
		IsRunning: p.Running(ctx),
	}

	if info.Complete {
		info.Successful = p.cmd.ProcessState.Success()
		info.PID = -1
	}

	if info.IsRunning {
		info.PID = p.cmd.Process.Pid
	}

	return info
}

func (p *basicProcess) Complete(ctx context.Context) bool {
	if p.cmd == nil {
		return false
	}

	return p.cmd.Process.Pid == -1
}

func (p *basicProcess) Running(ctx context.Context) bool {
	// if we haven't created the command or it hasn't started than
	// it isn't running
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}

	// ProcessState is populated once you start waiting for the
	// process, but not until then. Exited can be false if the
	// process was stopped/canceled.
	if p.cmd.ProcessState != nil && !p.cmd.ProcessState.Exited() {
		return true
	}

	// if we have a viable pid then it's running
	return p.cmd.Process.Pid > 0
}

func (p *basicProcess) Signal(ctx context.Context, sig syscall.Signal) error {
	return errors.Wrapf(p.cmd.Process.Signal(sig), "problem sending signal '%s' to '%s'",
		sig, p.id)
}

func (p *basicProcess) Wait(ctx context.Context) error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	if p.cmd.ProcessState.Exited() {
		return nil
	}

	sig := make(chan error)
	go func() {
		defer close(sig)

		select {
		case sig <- p.cmd.Wait():
		case <-ctx.Done():
			select {
			case sig <- ctx.Err():
			default:
			}
		}

		return
	}()

	select {
	case <-ctx.Done():
		return errors.New("context canceled while waiting for process to exit")
	case err := <-sig:
		return errors.WithStack(err)
	}
}
