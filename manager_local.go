package jasper

import (
	"context"
	"sync"

	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

// MakeLocalManager wraps the given manager in a thread-safe Manager.
func MakeLocalManager(manager Manager) (Manager, error) {
	return &localProcessManager{manager: manager}, nil
}

// NewLocalManager is a constructor for a thread-safe Manager.
func NewLocalManager(trackProcs bool) (Manager, error) {
	basicManager, err := newBasicProcessManager(map[string]Process{}, false, trackProcs)
	if err != nil {
		return nil, err
	}

	return &localProcessManager{manager: basicManager}, nil
}

// NewLocalManagerBlockingProcesses is a constructor for a thread-safe Manager
// that uses blockingProcess instead of the default basicProcess.
func NewLocalManagerBlockingProcesses(trackProcs bool) (Manager, error) {
	basicBlockingManager, err := newBasicProcessManager(map[string]Process{}, true, trackProcs)
	if err != nil {
		return nil, err
	}

	return &localProcessManager{manager: basicBlockingManager}, nil
}

type localProcessManager struct {
	mu      sync.RWMutex
	manager Manager
}

func (m *localProcessManager) ID() string {
	return m.manager.ID()
}

func (m *localProcessManager) CreateProcess(ctx context.Context, opts *options.Create) (Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	proc, err := m.manager.CreateProcess(ctx, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &localProcess{proc: proc}, nil
}

func (m *localProcessManager) CreateCommand(ctx context.Context) *Command {
	return NewCommand().ProcConstructor(m.CreateProcess)
}

func (m *localProcessManager) Register(ctx context.Context, proc Process) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return errors.WithStack(m.manager.Register(ctx, proc))
}

func (m *localProcessManager) List(ctx context.Context, f options.Filter) ([]Process, error) {
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

func (m *localProcessManager) Clear(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.manager.Clear(ctx)
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
