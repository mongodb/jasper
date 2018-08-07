package jasper

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

type localProcess struct {
	id   string
	opts CreateOptions
	cmd  *exec.Cmd
}

func (p *localProcess) ID() string { return p.id }
func (p *localProcess) Info(ctx context.Context) ProcessInfo {
	hn, _ := os.Hostname()

	if p.Running(ctx) {
		return ProcessInfo{
			Host:      hn,
			PID:       p.cmd.ProcessState.Pid(),
			IsRunning: true,
			ID:        p.id,
			Options:   CreateOptions,
		}
	}

	return ProcessInfo{
		ID:        p.id,
		Options:   CreateOptions,
		PID:       -1,
		Host:      hn,
		IsRunning: false,
	}
}

func (p *localProcess) Running(ctx context.Context) bool {
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}

	return !p.cmd.ProcessState.Exited()
}

func (p *localProcess) Signal(ctx context.Context, sig syscall.Signal) error {
	return p.cmd.Process.Signal(sig)
}

func (p *localProcess) Wait(ctx context.Context) error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	if p.cmd.ProcessState.Exited() {
		return nil
	}

	sig := make(chan error)
	go func() {
		sig <- p.cmd.Wait()
		close(sig)
	}()

	select {
	case <-ctx.Done():
		return errors.New("context canceled while waiting for process to exit")
	case err := <-sig:
		return errors.WithStack(err)
	}
}
