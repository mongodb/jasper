package jasper

import (
	"context"

	"github.com/pkg/errors"
)

// NewSelfClearingProcessManager creates and returns a process manager that will
// clear itself of dead processes without the need for calling Clear() from the
// user. Clear() however can be called proactively.
func NewSelfClearingProcessManager(numProcs int) Manager {
	return &selfClearingProcessManager{
		manager: &basicProcessManager{
			procs:              map[string]Process{},
			skipDefaultTrigger: true,
		},
		maxProcs: numProcs,
	}
}

type selfClearingProcessManager struct {
	manager  *basicProcessManager
	maxProcs int
}

func (m *selfClearingProcessManager) checkProcCapacity(ctx context.Context) error {
	// TODO: Do a num procs check here and attempt a clear?
	if len(m.manager.procs) == m.maxProcs {
		// We are at capacity, we can try to perform a clear
		if err := m.Clear(ctx); err != nil && len(m.manager.procs) == m.maxProcs {
			return errors.New("cannot create any more processes")
		}
	}

	return nil
}

func (m *selfClearingProcessManager) Create(ctx context.Context, opts *CreateOptions) (Process, error) {
	if err := m.checkProcCapacity(ctx); err != nil {
		return nil, err
	}

	proc, err := m.manager.Create(ctx, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	proc.RegisterTrigger(ctx, makeDefaultTrigger(ctx, m, opts, proc.ID()))

	m.manager.procs[proc.ID()] = proc

	return proc, nil
}

func (m *selfClearingProcessManager) Register(ctx context.Context, proc Process) error {
	if err := m.checkProcCapacity(ctx); err != nil {
		return err
	}

	return errors.WithStack(m.manager.Register(ctx, proc))
}

func (m *selfClearingProcessManager) List(ctx context.Context, f Filter) ([]Process, error) {
	procs, err := m.manager.List(ctx, f)
	return procs, errors.WithStack(err)
}

func (m *selfClearingProcessManager) Get(ctx context.Context, id string) (Process, error) {
	proc, err := m.manager.Get(ctx, id)
	return proc, errors.WithStack(err)
}

func (m *selfClearingProcessManager) Clear(ctx context.Context) error {
	return errors.WithStack(m.Clear(ctx))
}

func (m *selfClearingProcessManager) Close(ctx context.Context) error {
	return errors.WithStack(m.Close(ctx))
}

func (m *selfClearingProcessManager) Group(ctx context.Context, name string) ([]Process, error) {
	procs, err := m.manager.Group(ctx, name)
	return procs, errors.WithStack(err)
}
