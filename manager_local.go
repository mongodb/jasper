package jasper

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type localProcessManager struct {
	mu      sync.RWMutex
	manager *basicProcessManager
}

func (m *localProcessManager) Create(ctx context.Context, opts *CreateOptions) (Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.manager.skipDefaultTrigger = true
	proc, err := m.manager.Create(ctx, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	proc.RegisterTrigger(ctx, makeDefaultTrigger(ctx, m, opts, proc.ID()))

	proc = &localProcess{proc: proc}
	m.manager.procs[proc.ID()] = proc

	return proc, nil
}

func (m *localProcessManager) List(ctx context.Context, f Filter) ([]Process, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	procs, err := m.manager.List(ctx, f)
	return procs, errors.WithStack(err)
}

func (m *localProcessManager) Get(ctx context.Context, id string) (Process, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	proc, err := m.manager.Get(ctx, id)
	return proc, errors.WithStack(err)
}

func (m *localProcessManager) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return errors.WithStack(m.manager.Close(ctx))
}

func (m *localProcessManager) Group(ctx context.Context, name string) ([]Process, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	procs, err := m.manager.Group(ctx, name)
	return procs, errors.WithStack(err)
}

type basicProcessManager struct {
	procs              map[string]Process
	blocking           bool
	skipDefaultTrigger bool
}

func (m *basicProcessManager) Create(ctx context.Context, opts *CreateOptions) (Process, error) {
	var (
		proc Process
		err  error
	)

	if m.blocking {
		proc, err = newBlockingProcess(ctx, opts)
	} else {
		proc, err = newBasicProcess(ctx, opts)
	}

	if err != nil {
		return nil, errors.Wrap(err, "problem constructing local process")
	}

	// TODO this will race because it runs later
	if !m.skipDefaultTrigger {
		proc.RegisterTrigger(ctx, makeDefaultTrigger(ctx, m, opts, proc.ID()))
	}

	m.procs[proc.ID()] = proc

	return proc, nil
}

func (m *basicProcessManager) List(ctx context.Context, f Filter) ([]Process, error) {
	out := []Process{}

	for _, proc := range m.procs {
		if ctx.Err() != nil {
			return nil, errors.New("operation canceled")
		}

		info := proc.Info(ctx)

		switch {
		case f == Terminated && !info.Complete:
			continue
		case f == Running && !info.IsRunning:
			continue
		case f == Successful && !info.Successful:
			continue
		case f == Failed && info.Successful:
			continue
		case f == All:
		}

		out = append(out, proc)
	}

	if len(out) == 0 {
		return nil, errors.New("no processes")
	}

	return out, nil
}

func (m *basicProcessManager) Get(ctx context.Context, id string) (Process, error) {
	proc, ok := m.procs[id]
	if !ok {
		return nil, errors.Errorf("process '%s' does not exist", id)
	}

	return proc, nil
}

func (m *basicProcessManager) Close(ctx context.Context) error {
	if len(m.procs) == 0 {
		return nil
	}
	procs, err := m.List(ctx, All)
	if err != nil {
		return errors.WithStack(err)
	}

	termCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := TerminateAll(termCtx, procs); err != nil {
		killCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		return errors.WithStack(KillAll(killCtx, procs))
	}

	for _, p := range m.procs {
		info := p.Info(ctx)
		for _, c := range info.Options.closers {
			c()
		}
	}

	return nil
}

func (m *basicProcessManager) Group(ctx context.Context, name string) ([]Process, error) {
	out := []Process{}
	for _, proc := range m.procs {
		if ctx.Err() != nil {
			return nil, errors.New("request canceled")
		}

		if sliceContains(proc.GetTags(), name) {
			out = append(out, proc)
		}
	}

	if len(out) == 0 {
		return nil, errors.Errorf("no jobs tagged '%s'", name)
	}

	return out, nil
}
