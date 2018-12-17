package jasper

import (
	"context"

	"github.com/pkg/errors"
)

// NewSelfClearingProcessManager creates and returns a process manager that will
// clear itself of dead processes without the need for calling Clear() from the
// user. Clear() however can be called proactively.
func NewSelfClearingProcessManager(maxProcs int) Manager {
	return &selfClearingProcessManager{
		local:    NewLocalManager().(*localProcessManager),
		maxProcs: maxProcs,
	}
}

type selfClearingProcessManager struct {
	local    *localProcessManager
	maxProcs int
}

func (m *selfClearingProcessManager) checkProcCapacity(ctx context.Context) error {
	if len(m.local.manager.procs) == m.maxProcs {
		// We are at capacity, we can try to perform a clear.
		if err := m.Clear(ctx); err != nil &&
			len(m.local.manager.procs) == m.maxProcs {
			return errors.New("cannot create any more processes, reached maxProcs")
		}
	}

	return nil
}

func (m *selfClearingProcessManager) Create(ctx context.Context, opts *CreateOptions) (Process, error) {
	if err := m.checkProcCapacity(ctx); err != nil {
		return nil, err
	}

	proc, err := m.local.Create(ctx, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	proc.RegisterTrigger(ctx, makeDefaultTrigger(ctx, m, opts, proc.ID()))

	m.local.manager.procs[proc.ID()] = proc

	return proc, nil
}

func (m *selfClearingProcessManager) Register(ctx context.Context, proc Process) error {
	if err := m.checkProcCapacity(ctx); err != nil {
		return err
	}

	return errors.WithStack(m.local.Register(ctx, proc))
}

func (m *selfClearingProcessManager) List(ctx context.Context, f Filter) ([]Process, error) {
	procs, err := m.local.List(ctx, f)
	return procs, errors.WithStack(err)
}

func (m *selfClearingProcessManager) Get(ctx context.Context, id string) (Process, error) {
	proc, err := m.local.Get(ctx, id)
	return proc, errors.WithStack(err)
}

func (m *selfClearingProcessManager) Clear(ctx context.Context) error {
	return errors.WithStack(m.local.Clear(ctx))
}

func (m *selfClearingProcessManager) Close(ctx context.Context) error {
	return errors.WithStack(m.local.Close(ctx))
}

func (m *selfClearingProcessManager) Group(ctx context.Context, name string) ([]Process, error) {
	procs, err := m.local.Group(ctx, name)
	return procs, errors.WithStack(err)
}
