package jasper

import (
	"github.com/mongodb/grip"
)

type windowsProcessTracker struct {
	job *Job
}

func newProcessTracker(name string) (processTracker, error) {
	job, err := NewJob(name)
	if err != nil {
		return nil, err
	}
	return &windowsProcessTracker{job: job}, nil
}

func (t *windowsProcessTracker) add(pid uint) error {
	return t.job.AssignProcess(pid)
}

func (t *windowsProcessTracker) cleanup() error {
	catcher := grip.NewBasicCatcher()
	catcher.Add(t.job.Terminate(0))
	catcher.Add(t.job.Close())
	return catcher.Resolve()
}
