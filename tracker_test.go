package jasper

import "github.com/pkg/errors"

type mockProcessTracker struct {
	FailAdd          bool
	FailUpdateLimits bool
	FailCleanup      bool
	trackedPIDs      []int
}

func newMockProcessTracker() mockProcessTracker {
	return mockProcessTracker{
		trackedPIDs: []int{},
	}
}

func (t *mockProcessTracker) add(pid int) error {
	if t.FailAdd {
		return errors.New("fail in add")
	}
	t.trackedPIDs = append(t.trackedPIDs, pid)
	return nil
}

func (t *mockProcessTracker) updateLimits(_ interface{}) error {
	if t.FailUpdateLimits {
		return errors.New("failed in updateLimits")
	}
	return nil
}

func (t *mockProcessTracker) cleanup() error {
	if t.FailCleanup {
		return errors.New("failed in cleanup")
	}
	t.trackedPIDs = []int{}
	return nil
}
