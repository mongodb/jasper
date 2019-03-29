package jasper

import "github.com/pkg/errors"

// ProcessTracker provides a way to logically group processes that
// should be managed collectively. Implementation details are
// platform-specific since each one has its own means of managing
// groups of processes.
type ProcessTracker interface {
	// Add begins tracking a process identified by its PID.
	Add(pid int) error
	// SetLimits modifies the resource limit on all tracked processes. The
	// expected limits parameter is platform-dependent.
	SetLimits(limits interface{}) error
	// Cleanup terminates this group of processes.
	Cleanup() error
}

// processTrackerBase provides convenience no-op implementations of the
// processTracker interface.
type processTrackerBase struct {
	Name string
}

func (processTrackerBase) Add(_ int) error { return nil }

func (processTrackerBase) SetLimits(_ interface{}) error { return nil }

func (processTrackerBase) Cleanup() error { return nil }

type mockProcessTracker struct {
	FailAdd          bool
	FailUpdateLimits bool
	FailCleanup      bool
	PIDs             []int
}

func newMockProcessTracker() ProcessTracker {
	return &mockProcessTracker{
		PIDs: []int{},
	}
}

func (t *mockProcessTracker) Add(pid int) error {
	if t.FailAdd {
		return errors.New("fail in add")
	}
	t.PIDs = append(t.PIDs, pid)
	return nil
}

func (t *mockProcessTracker) SetLimits(_ interface{}) error {
	if t.FailUpdateLimits {
		return errors.New("failed in setLimits")
	}
	return nil
}

func (t *mockProcessTracker) Cleanup() error {
	if t.FailCleanup {
		return errors.New("failed in cleanup")
	}
	t.PIDs = []int{}
	return nil
}
