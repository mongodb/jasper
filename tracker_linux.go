package jasper

import (
	"os"
	"syscall"

	"github.com/containerd/cgroups"
	"github.com/mongodb/grip"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

const (
	defaultSubsystem = cgroups.Memory
)

func newProcessTracker(name string) (processTracker, error) {
	cgroup, err := cgroups.New(cgroups.V1, cgroups.StaticPath(name), &specs.LinuxResources{})
	if err != nil {
		return nil, errors.Wrap(err, "could not create default cgroup")
	}

	tracker := &linuxProcessTracker{cgroup: cgroup}

	return tracker, nil
}

func (t *linuxProcessTracker) add(pid int) error {
	proc := cgroups.Process{Subsystem: defaultSubsystem, Pid: pid}

	// kim: TODO: check if pid is already in the cgroup when forked. Can do this in testing.
	t.cgroup.Add(proc)
	return nil
}

// kim: updateLimit updates the resource limit on all tracked processes.
func (t *linuxProcessTracker) updateLimit(limit Subsystem) error {

	return nil
}

func (t *linuxProcessTracker) cleanup() error {
	// Get all procs.
	cgroupProcs, err := t.cgroup.Processes(defaultSubsystem, true)
	if err != nil {
		return errors.Wrap(err, "could not find tracked processes")
	}

	catcher := grip.NewBasicCatcher()
	for _, cgroupProc := range cgroupProcs {
		osProc, err := os.FindProcess(cgroupProc.Pid)
		catcher.Add(errors.Wrap(err, "failed to find tracked process"))
		err = osProc.Signal(syscall.SIGTERM)
		if err != nil {
			catcher.Add(errors.Wrap(err, "could not signal process to terminate"))
			catcher.Add(errors.Wrap(osProc.Kill(), "could not kill process"))
		}
	}

	return catcher.Resolve()
}
