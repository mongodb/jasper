package jasper

import (
	"context"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

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
	cmd, err := opts.Resolve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "problem building command from options")
	}

	// don't check the error here, becaues we need to call Start to
	// populate the process state, and we'll race to calls to
	// methods with the reactor starting up if we skip it here.
	_ = cmd.Start()

	opts.started = true

	p := &blockingProcess{
		id:   uuid.Must(uuid.NewV4()).String(),
		opts: *opts,
		tags: make(map[string]struct{}),
		ops:  make(chan func(*exec.Cmd)),
	}

	for _, t := range opts.Tags {
		p.Tag(t)
	}

	p.host, _ = os.Hostname()

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
				"id":           info.ID,
				"cmd":          strings.Join(p.opts.Args, " "),
				"success":      info.Successful,
				"num_triggers": len(p.triggers),
			}))

			p.info = &info
			p.triggers.Run(info)
			return
		case <-ctx.Done():
			// note, the process might take a moment to
			// die when it gets here.
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
			if op != nil {
				op(cmd)
			}
		}
	}

}

func (p *blockingProcess) ID() string { return p.id }
func (p *blockingProcess) Info(ctx context.Context) ProcessInfo {
	if p.info != nil {
		return *p.info
	}

	out := make(chan ProcessInfo)
	operation := func(cmd *exec.Cmd) {
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
	case p.ops <- operation:
		select {
		case res := <-out:
			return res
		case <-ctx.Done():
			return ProcessInfo{}
		}
	case <-ctx.Done():
		return ProcessInfo{}
	}
}

func (p *blockingProcess) Running(ctx context.Context) bool {
	if p.info != nil {
		return false
	}

	out := make(chan bool)
	operation := func(cmd *exec.Cmd) {
		defer close(out)

		if cmd == nil || cmd.Process == nil {
			out <- false
			return
		}

		if cmd.Process.Pid <= 0 {
			out <- false
			return
		}

		out <- true
	}

	select {
	case p.ops <- operation:
		return <-out
	case <-ctx.Done():
		return false
	}
}

func (p *blockingProcess) Complete(ctx context.Context) bool {
	return p.info != nil
}

func (p *blockingProcess) Signal(ctx context.Context, sig syscall.Signal) error {
	if p.info != nil {
		return errors.New("cannot signal a process that has terminated")
	}

	out := make(chan error)
	operation := func(cmd *exec.Cmd) {
		defer close(out)

		if cmd == nil {
			out <- errors.New("cannot signal nil process")
			return
		}

		out <- errors.Wrapf(cmd.Process.Signal(sig), "problem sending signal '%s' to '%s'",
			sig, p.id)
	}
	select {
	case p.ops <- operation:
		select {
		case res := <-out:
			return res
		case <-ctx.Done():
			return errors.New("context canceled")
		}
	case <-ctx.Done():
		return errors.New("context canceled")
	}
}

func (p *blockingProcess) RegisterTrigger(ctx context.Context, trigger ProcessTrigger) error {
	if p.Complete(ctx) {
		return errors.New("cannot register trigger after process exits")
	}

	if trigger == nil {
		return errors.New("cannot register nil trigger")
	}

	p.triggers = append(p.triggers, trigger)

	return nil
}

func (p *blockingProcess) Wait(ctx context.Context) error {
	if p.info != nil {
		return nil
	}

	out := make(chan error)
	waiter := func(cmd *exec.Cmd) {
		if p.info == nil {
			return
		}

		if p.info.Successful {
			out <- nil
			return
		}

		out <- errors.New("task exited with error")
	}

	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			timer.Reset(time.Duration(rand.Int63n(50)) * time.Millisecond)
		case p.ops <- waiter:
			continue
		case <-ctx.Done():
			return errors.New("wait operation canceled")
		case err := <-out:
			return errors.WithStack(err)
		default:
			if p.info != nil {
				if p.info.Successful {
					return nil
				}
				return errors.New("operation failed")
			}
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
