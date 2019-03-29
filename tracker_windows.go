package jasper

import (
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

type windowsProcessTracker struct {
	processTrackerBase
	job *Job
}

func NewProcessTracker(name string) (ProcessTracker, error) {
	job, err := NewJob(name)
	if err != nil {
		return nil, err
	}
	return &windowsProcessTracker{processTrackerBase: processTrackerBase{Name: name}, job: job}, nil
}

func (t *windowsProcessTracker) Add(pid int) error {
	if t.job == nil {
		return errors.New("cannot add process because job is invalid")
	}
	return t.job.AssignProcess(uint(pid))
}

func (t *windowsProcessTracker) Cleanup() error {
	if t.job == nil {
		return errors.New("cannot close because job is invalid")
	}
	catcher := grip.NewBasicCatcher()
	catcher.Add(t.job.Terminate(0))
	catcher.Add(t.job.Close())
	t.job = nil
	return catcher.Resolve()
}
