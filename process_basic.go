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

type basicProcess struct {
	info        ProcessInfo
	cmd         *exec.Cmd
	err         error
	id          string
	hostname    string
	opts        CreateOptions
	tags        map[string]struct{}
	triggers    ProcessTriggerSequence
	safety      chan struct{}
	initialized chan struct{}
	sync.RWMutex
}

func newBasicProcess(ctx context.Context, opts *CreateOptions) (Process, error) {
	id := uuid.Must(uuid.NewV4()).String()
	opts.AddEnvVar(EnvironID, id)

	cmd, err := opts.Resolve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "problem building command from options")
	}

	p := &basicProcess{
		id:          id,
		opts:        *opts,
		cmd:         cmd,
		tags:        make(map[string]struct{}),
		safety:      make(chan struct{}),
		initialized: make(chan struct{}),
		triggers:    ProcessTriggerSequence{},
	}
	p.hostname, _ = os.Hostname()

	for _, t := range opts.Tags {
		p.Tag(t)
	}

	p.RegisterTrigger(ctx, makeOptionsCloseTrigger())

	err = cmd.Start()
	if err != nil {
		return nil, errors.Wrap(err, "problem creating command")
	}

	p.info.ID = p.id
	p.info.Options = p.opts
	p.info.Host = p.hostname

	go func() {
		errs := make(chan error)
		sig := make(chan struct{})
		go func() {
			defer close(errs)
			close(sig)
			errs <- cmd.Wait()
		}()
		<-sig

		p.Lock()
		p.opts.started = true
		p.info.IsRunning = true
		p.info.PID = p.cmd.Process.Pid
		p.cmd = cmd
		close(p.initialized)
		p.Unlock()

		// cmd.Wait() has returned if we get past this line.
		err := <-errs
		p.Lock()
		p.err = err
		close(p.safety)
		p.info.IsRunning = false
		p.info.Complete = true
		p.info.ExitCode = p.cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		p.triggers.Run(p.info)
		p.Unlock()
	}()

	opts.started = true

	return p, nil
}

func (p *basicProcess) ID() string {
	p.RLock()
	defer p.RUnlock()

	return p.id
}
func (p *basicProcess) Info(_ context.Context) ProcessInfo {
	p.RLock()
	defer p.RUnlock()

	return p.info
}

func (p *basicProcess) Complete(_ context.Context) bool {
	<-p.initialized
	p.RLock()
	defer p.RUnlock()

	return p.info.Complete
}

func (p *basicProcess) Running(_ context.Context) bool {
	<-p.initialized
	p.RLock()
	defer p.RUnlock()

	return p.info.IsRunning
}

func (p *basicProcess) Signal(_ context.Context, sig syscall.Signal) error {
	p.RLock()
	defer p.RUnlock()

	if p.info.IsRunning {
		return errors.Wrapf(p.cmd.Process.Signal(sig), "problem sending signal '%s' to '%s'", sig, p.id)
	}
	return nil
}

func (p *basicProcess) Respawn(ctx context.Context) (Process, error) {
	p.RLock()
	defer p.RUnlock()

	opts := p.Info(ctx).Options
	opts.closers = []func(){}

	return newBasicProcess(ctx, &opts)
}

func (p *basicProcess) Wait(ctx context.Context) (int, error) {
	runningCheck := func() bool {
		<-p.initialized
		p.RLock()
		defer p.RUnlock()
		return p.info.IsRunning
	}

	if !runningCheck() {
		p.RLock()
		defer p.RUnlock()

		return p.info.ExitCode, p.err
	}

	select {
	case <-ctx.Done():
		return -1, errors.New("operation canceled")
	case <-p.safety:
	}

	return p.info.ExitCode, p.err
}

func (p *basicProcess) RegisterTrigger(ctx context.Context, trigger ProcessTrigger) error {
	if trigger == nil {
		return errors.New("cannot register nil trigger")
	}

	p.Lock()
	defer p.Unlock()

	if p.info.Complete {
		return errors.New("cannot register trigger after process exits")
	}

	p.triggers = append(p.triggers, trigger)

	return nil
}

func (p *basicProcess) Tag(t string) {
	_, ok := p.tags[t]
	if ok {
		return
	}

	p.tags[t] = struct{}{}
	p.opts.Tags = append(p.opts.Tags, t)
}

func (p *basicProcess) ResetTags() {
	p.tags = make(map[string]struct{})
	p.opts.Tags = []string{}
}

func (p *basicProcess) GetTags() []string {
	out := []string{}
	for t := range p.tags {
		out = append(out, t)
	}
	return out
}
