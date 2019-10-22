package jasper

import (
	"context"
	"sync"

	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

// NewSynchronizedManager is a constructor for a thread-safe Manager.
func NewSynchronizedManager(trackProcs bool) (Manager, error) {
	basicManager, err := newBasicProcessManager(map[string]Process{}, false, false, trackProcs)
	if err != nil {
		return nil, err
	}
	return &synchronizedProcessManager{
		manager: basicManager.(*basicProcessManager),
	}, nil
}

// NewSynchronizedManagerBlockingProcesses is a constructor for synchronizedProcessManager,
// that uses blockingProcess instead of the default basicProcess.
func NewSynchronizedManagerBlockingProcesses(trackProcs bool) (Manager, error) {
	basicBlockingManager, err := newBasicProcessManager(map[string]Process{}, false, true, trackProcs)
	if err != nil {
		return nil, err
	}
	return &synchronizedProcessManager{
		manager: basicBlockingManager.(*basicProcessManager),
	}, nil
}

type synchronizedProcessManager struct {
	mu      sync.RWMutex
	manager *basicProcessManager
}

func (m *synchronizedProcessManager) ID() string {
	return m.manager.ID()
}

func (m *synchronizedProcessManager) CreateProcess(ctx context.Context, opts *options.Create) (Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.manager.skipDefaultTrigger = true
	proc, err := m.manager.CreateProcess(ctx, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_ = proc.RegisterTrigger(ctx, makeDefaultTrigger(ctx, m, opts, proc.ID()))

	proc = &synchronizedProcess{proc: proc}
	m.manager.procs[proc.ID()] = proc

	return proc, nil
}

func (m *synchronizedProcessManager) CreateCommand(ctx context.Context) *Command {
	return NewCommand().ProcConstructor(m.CreateProcess)
}

func (m *synchronizedProcessManager) Register(ctx context.Context, proc Process) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return errors.WithStack(m.manager.Register(ctx, proc))
}

func (m *synchronizedProcessManager) List(ctx context.Context, f options.Filter) ([]Process, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	procs, err := m.manager.List(ctx, f)
	return procs, errors.WithStack(err)
}

func (m *synchronizedProcessManager) Get(ctx context.Context, id string) (Process, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	proc, err := m.manager.Get(ctx, id)
	return proc, errors.WithStack(err)
}

func (m *synchronizedProcessManager) Clear(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.manager.Clear(ctx)
}

func (m *synchronizedProcessManager) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return errors.WithStack(m.manager.Close(ctx))
}

func (m *synchronizedProcessManager) Group(ctx context.Context, name string) ([]Process, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	procs, err := m.manager.Group(ctx, name)
	return procs, errors.WithStack(err)
}
