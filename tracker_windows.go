package jasper

import (
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

type windowsProcessTracker struct {
	*processTrackerBase
	job *JobObject
}

func (t *windowsProcessTracker) setJobIfInvalid() error {
	if t.job != nil {
		return nil
	}
	job, err := NewWindowsJobObject(t.Name)
	if err != nil {
		return errors.Wrap(err, "creating new job object")
	}
	t.job = job
	return nil
}

// NewProcessTracker creates a job object for all tracked processes.
func NewProcessTracker(name string) (ProcessTracker, error) {
	t := &windowsProcessTracker{processTrackerBase: &processTrackerBase{Name: name}}
	if err := t.setJobIfInvalid(); err != nil {
		return nil, errors.Wrap(err, "creating job object for new process tracker")
	}
	return t, nil
}

func (t *windowsProcessTracker) Add(info ProcessInfo) error {
	if err := t.setJobIfInvalid(); err != nil {
		return errors.Wrap(err, "setting up job object")
	}
	return t.job.AssignProcess(uint(info.PID))
}

func (t *windowsProcessTracker) Cleanup() error {
	if t.job == nil {
		return nil
	}
	catcher := grip.NewBasicCatcher()
	catcher.Add(t.job.Terminate(0))
	catcher.Add(t.job.Close())
	t.job = nil
	return catcher.Resolve()
}
