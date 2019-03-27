package jasper

import (
	"encoding/json"
	"os"
	"syscall"

	"github.com/containerd/cgroups"
	"github.com/mongodb/grip"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

const (
	// defaultSubsystem is the default subsystem where all tracked processes are
	// added. There is no significance behind using the freezer subsystem
	// over any other subsystem for this purpose; its purpose is to ensure all
	// processes can be tracked in a single subsystem for cleanup.
	defaultSubsystem = cgroups.Freezer
)

type linuxProcessTracker struct {
	processTrackerBase
	cgroup     cgroups.Cgroup
	cgroupName string
}

// newProcessTracker creates a cgroup for all tracked processes. Cgroups
// functionality requires admin privileges.
func newProcessTracker(name string) (processTracker, error) {
	tracker := &linuxProcessTracker{cgroupName: name}
	if err := tracker.setDefaultCgroupIfInvalid(); err != nil {
		return nil, errors.Wrap(err, "could not initialize process tracker")
	}

	return tracker, nil
}

func (t *linuxProcessTracker) valid() bool {
	return t.cgroup != nil && t.cgroup.State() != cgroups.Deleted
}

func (t *linuxProcessTracker) setDefaultCgroupIfInvalid() error {
	if t.valid() {
		return nil
	}

	cgroup, err := cgroups.New(cgroups.V1, cgroups.StaticPath("/"+t.cgroupName), &specs.LinuxResources{})
	if err != nil {
		return errors.Wrap(err, "could not create default cgroup")
	}
	t.cgroup = cgroup

	return nil
}

func (t *linuxProcessTracker) add(pid int) error {
	if err := t.setDefaultCgroupIfInvalid(); err != nil {
		return errors.Wrapf(err, "failed to initialize cgroup while adding process with pid '%d'", pid)
	}

	proc := cgroups.Process{Subsystem: defaultSubsystem, Pid: pid}
	if err := t.cgroup.Add(proc); err != nil {
		return errors.Wrap(err, "failed to add process with pid '%d' to cgroup")
	}
	return nil
}

// updateLimits requires a pointer to a LinuxResources struct. The new
// limits cannot be less than the current resource usage of the tracked
// processes.
func (t *linuxProcessTracker) updateLimits(limits interface{}) error {
	linuxResourceLimits, ok := limits.(*LinuxResources)
	if !ok {
		return errors.New("process tracker requires (*LinuxResources) in order to update resource limits")
	}
	return t.doUpdateLimits(linuxResourceLimits)
}

func (t *linuxProcessTracker) doUpdateLimits(resourceLimits *LinuxResources) error {
	bytes, err := json.Marshal(resourceLimits)
	if err != nil {
		return errors.Wrap(err, "error marshalling resource limits")
	}
	specsResourceLimits := &specs.LinuxResources{}
	if err = json.Unmarshal(bytes, specsResourceLimits); err != nil {
		return errors.Wrap(err, "error unmarshalling resource limits")
	}

	if err := t.setDefaultCgroupIfInvalid(); err != nil {
		return errors.Wrapf(err, "failed to initialize cgroup while updating limits")
	}

	return t.cgroup.Update(specsResourceLimits)
}

func (t *linuxProcessTracker) listPIDs() ([]int, error) {
	procs, err := t.cgroup.Processes(defaultSubsystem, false)
	if err != nil {
		return nil, errors.Wrap(err, "could not list tracked PIDs")
	}

	pids := make([]int, 0, len(procs))
	for _, proc := range procs {
		pids = append(pids, proc.Pid)
	}
	return pids, nil
}

func (t *linuxProcessTracker) cleanup() error {
	if !t.valid() {
		return nil
	}

	// Get all tracked processes from the default subsystem.
	pids, err := t.listPIDs()
	if err != nil {
		return errors.Wrap(err, "could not find tracked processes")
	}

	// Attempt to kill all tracked and running processes.
	catcher := grip.NewBasicCatcher()
	for _, pid := range pids {
		osProc, err := os.FindProcess(pid)
		catcher.Add(errors.Wrap(err, "failed to find tracked process"))
		err = osProc.Signal(syscall.SIGTERM)
		if err != nil {
			catcher.Add(errors.Wrap(err, "could not signal process to terminate"))
			catcher.Add(errors.Wrap(osProc.Kill(), "could not kill process"))
		}
	}

	// Delete the cgroup. If the process tracker is still used, the cgroup must
	// be re-initialized.
	catcher.Add(t.cgroup.Delete())

	return catcher.Resolve()
}
