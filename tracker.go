package jasper

import "github.com/pkg/errors"

// processTracker provides a way to logically group processes that
// should be managed collectively. Implementation details are
// platform-specific since each one has its own means of managing
// groups of processes.
type processTracker interface {
	// add begins tracking a process identified by its PID.
	add(pid int) error
	// setLimits modifies the resource limit on all tracked processes. The
	// expected limits parameter is platform-dependent.
	setLimits(limits interface{}) error
	// cleanup terminates this group of processes.
	cleanup() error
}

// processTrackerBase provides convenience no-op implementations of the
// processTracker interface.
type processTrackerBase struct{}

func (processTrackerBase) add(_ int) error { return nil }

func (processTrackerBase) setLimits(_ interface{}) error { return nil }

func (processTrackerBase) cleanup() error { return nil }

type mockProcessTracker struct {
	failAdd          bool
	failUpdateLimits bool
	failCleanup      bool
	pids             []int
}

func newMockProcessTracker() mockProcessTracker {
	return mockProcessTracker{
		pids: []int{},
	}
}

func (t *mockProcessTracker) add(pid int) error {
	if t.failAdd {
		return errors.New("fail in add")
	}
	t.pids = append(t.pids, pid)
	return nil
}

func (t *mockProcessTracker) setLimits(_ interface{}) error {
	if t.failUpdateLimits {
		return errors.New("failed in setLimits")
	}
	return nil
}

func (t *mockProcessTracker) cleanup() error {
	if t.failCleanup {
		return errors.New("failed in cleanup")
	}
	t.pids = []int{}
	return nil
}
