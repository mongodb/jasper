package jasper

// processTracker provides a way to logically group processes that
// should be managed collectively. Implementation details are
// platform-specific since each one has its own means of managing
// groups of processes.
type processTracker interface {
	// add begins tracking a process identified by its PID.
	add(pid int) error
	// updateLimits modifies the resource limit on all tracked processes. The
	// expected limits parameter is platform-dependent.
	updateLimits(limits interface{}) error
	// cleanup terminates this group of processes.
	cleanup() error
}

// processTrackerBase provides convenience no-op implementations of the
// processTracker interface.
type processTrackerBase struct{}

func (processTrackerBase) add(_ int) error { return nil }

func (processTrackerBase) updateLimits(_ interface{}) error { return nil }

func (processTrackerBase) cleanup() error { return nil }
