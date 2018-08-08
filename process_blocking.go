package jasper

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type blockingProcess struct {
	id       string
	info     *ProcessInfo
	host     string
	opts     CreateOptions
	ops      chan func(*exec.Cmd)
	tags     map[string]struct{}
	triggers ProcessTriggerSequence
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
		tags: make(map[string]struct{}),
		ops:  make(chan func(*exec.Cmd)),
	}

	for _, t := range opts.Tags {
		p.Tag(t)
	}

	go p.reactor(ctx, cmd)
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
			info := ProcessInfo{
				ID:         p.id,
				Options:    p.opts,
				Host:       p.host,
				Complete:   true,
				IsRunning:  false,
				Successful: cmd.ProcessState.Success(),
			}
			grip.Debug(message.WrapError(err, message.Fields{
				"info":         info,
				"num_triggers": len(p.triggers),
			}))

			p.info = &info
			p.triggers.Run(info)
			return
		case <-ctx.Done():
			cmd.Process.Kill()

			info := ProcessInfo{
				ID:         p.id,
				Options:    p.opts,
				Host:       p.host,
				Complete:   true,
				IsRunning:  false,
				Successful: false,
			}
			p.info = &info
			p.triggers.Run(info)
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
	}

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

func (p *blockingProcess) RegisterTrigger(trigger ProcessTrigger) error {
	if p.info != nil {
		return errors.New("cannot register triggers after process completes")
	}

	if trigger == nil {
		return errors.New("cannot register nil trigger")
	}

	p.triggers = append(p.triggers, trigger)

	return nil
}

func (p *blockingProcess) Wait(ctx context.Context) error {
	out := make(chan error)

	if p.info != nil {
		return errors.New("cannot register triggers after process completes")
	}

	waiter := func(cmd *exec.Cmd) {
		select {
		case <-ctx.Done():
			out <- errors.New("operation canceled")
			close(out)
		default:

		}
	}

	for {
		select {
		case p.ops <- waiter:
			continue
		case <-ctx.Done():
			return errors.New("wait operation canceled")
		case err := <-out:
			return errors.WithStack(err)
		}
	}
}

func (p *blockingProcess) Tag(t string) {
	_, ok := p.tags[t]
	if ok {
		return
	}

	p.tags[t] = struct{}{}
	p.opts.Tags = append(p.opts.Tags, t)
}

func (p *blockingProcess) ResetTags() {
	p.tags = make(map[string]struct{})
	p.opts.Tags = []string{}
}

func (p *blockingProcess) GetTags() []string {
	out := []string{}
	for t := range p.tags {
		out = append(out, t)
	}
	return out
}
