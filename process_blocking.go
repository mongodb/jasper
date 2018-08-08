package jasper

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type blockingProcess struct {
	id   string
	info *ProcessInfo
	host string
	opts CreateOptions
	ops  chan func(*exec.Cmd)
}

func newBlockingProcess(ctx context.Context, opts *CreateOptions) (Process, error) {
	hn, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "problem finding hostname when creating process")
	}

	cmd, err := opts.Resolve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "problem building command from options")
	}

	p := &blockingProcess{
		id:   uuid.Must(uuid.NewV4()).String(),
		host: hn,
		opts: *opts,
		ops:  make(chan func(*exec.Cmd)),
	}
	go p.reactor(start, cmd)
	return p, nil
}

func (p *blockingProcess) reactor(ctx context.Context, cmd *exec.Cmd) {
	signal := make(chan error)
	go func() {
		defer close(signal)
		signal <- cmd.Wait()
	}()

	for {
		select {
		case err := <-signal:
			p.info = &ProcessInfo{
				ID:         p.id,
				Options:    p.opts,
				Host:       p.host,
				Complete:   true,
				IsRunning:  false,
				Successful: cmd.ProcessState.Success(),
			}

			return
		case <-ctx.Done():
			cmd.Process.Kill()

			p.info = &ProcessInfo{
				ID:         p.id,
				Options:    p.opts,
				Host:       p.host,
				Complete:   true,
				IsRunning:  false,
				Successful: false,
			}

			return
		case op := <-p.ops:
			op(cmd)
		}
	}

}

func (p *blockingProcess) ID() string { return p.id }
func (p *blockingProcess) Info(ctx context.Context) ProcessInfo {
	if p.info != nil {
		return *p.info
	}

	out := make(chan ProcessInfo)
	p.ops <- func(cmd *exec.Cmd) {
		out <- ProcessInfo{
			ID:        p.id,
			Options:   p.opts,
			Host:      p.host,
			Complete:  cmd.Process.Pid == -1,
			IsRunning: cmd.Process.Pid > 0,
		}
		close(out)
	}()

	select {
	case <-ctx.Done():
		return ProcessInfo{}
	case res := <-out:
		return res
	}
}

func (p *blockingProcess) Running(ctx context.Context) bool {
	if p.info != nil {
		return false
	}

	out := make(chan bool)
	p.ops <- func(cmd *exec.Cmd) {
		defer close(out)

		if cmd.ProcessState.Pid() > 0 {
			out <- true
			return
		}

		out <- false
	}
	return <-out
}

func (p *blockingProcess) Complete(ctx context.Context) bool {
	return p.info != nil
}

func (p *blockingProcess) Signal(ctx context.Context, sig syscall.Signal) error {
	if p.info != nil {
		return errors.New("cannot signal a process that has terminated")
	}

	out := make(chan error)
	p.ops <- func(cmd *exec.Cmd) {
		defer close(out)
		out <- errors.Wrapf(cmd.Process.Signal(sig), "problem sending signal '%s' to '%s'",
			sig, p.id)

	}

	return errors.WithStack(<-out)
}
