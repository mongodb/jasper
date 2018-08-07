package jasper

import (
	"context"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

type Process interface {
	ID() string
	Info(context.Context) ProcessInfo
	Running(context.Context) bool
	Complete(context.Context) bool
	Signal(context.Context, syscall.Signal) error
	Wait(context.Context) error
}

type ProcessInfo struct {
	ID         string
	Host       string
	PID        int
	IsRunning  bool
	Successful bool
	Complete   bool
	Options    CreateOptions
}

// Manager provides a basic, high level process management interface
type Manager interface {
	Create(context.Context, *CreateOptions) (Process, error)
	List(context.Context, Filter) ([]Process, error)
	Get(context.Context, string) (Process, error)
	Close(context.Context) error
}

type Filter string

const (
	Running    Filter = "running"
	Terminated        = "terminated"
	All               = "all"
	Failed            = "failed"
	Successful        = "successful"
)

func (f Filter) Validate() error {
	switch f {
	case Running, Terminated, All, Failed, Successful:
		return nil
	default:
		return errors.Errorf("%s is not a valid filter", f)
	}

}

////////////////////////////////////////////////////////////////////////

type localProcessManager struct {
	mu      sync.RWMutex
	manager *basicProcessManager
}

func (m *localProcessManager) Create(ctx context.Context, opts *CreateOptions) (Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	proc, err := m.manager.Create(ctx, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

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

type basicProcessManager struct {
	procs map[string]Process
}

func (m *basicProcessManager) Create(ctx context.Context, opts *CreateOptions) (Process, error) {
	proc, err := newBasicProcess(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "problem constructing local process")
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
		case f == Terminated && info.Complete:
		case f == Running && info.IsRunning:
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

	return nil
}
