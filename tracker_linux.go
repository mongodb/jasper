package jasper

import (
	"encoding/json"
	"os"

	"github.com/containerd/cgroups"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
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

// linuxProcessTracker uses cgroups to track processes.
type linuxProcessTracker struct {
	processTrackerBase
	cgroup     cgroups.Cgroup
	cgroupName string
}

// NewProcessTracker creates a cgroup for all tracked processes if supported.
// Cgroups functionality requires admin privileges. If cgroups are not
// supported, all ProcessTracker functions are no-ops except for Cleanup, which
// terminates processes that have the environment variable EnvironID set to a
// value that matches the process tracker's name.
func NewProcessTracker(name string) (ProcessTracker, error) {
	tracker := &linuxProcessTracker{processTrackerBase: processTrackerBase{Name: name}}
	if err := tracker.setDefaultCgroupIfInvalid(); err != nil {
		grip.Debug(message.WrapErrorf(err, "could not initialize process tracker named '%s' with cgroup", name))
	}

	return tracker, nil
}

// validCgroup returns true if the cgroup is non-nil and not deleted.
func (t *linuxProcessTracker) validCgroup() bool {
	return t.cgroup != nil && t.cgroup.State() != cgroups.Deleted
}

// setDefaultCgroupIfInvalid attempts to set the tracker's cgroup if it is
// invalid. This can fail if cgroups is not a supported feature on this
// platform.
func (t *linuxProcessTracker) setDefaultCgroupIfInvalid() error {
	if t.validCgroup() {
		return nil
	}

	cgroup, err := cgroups.New(cgroups.V1, cgroups.StaticPath("/"+t.cgroupName), &specs.LinuxResources{})
	if err != nil {
		return errors.Wrap(err, "could not create default cgroup")
	}
	t.cgroup = cgroup

	return nil
}

// Add adds this PID to the cgroup if cgroups is available. Otherwise, it no-ops.
func (t *linuxProcessTracker) Add(pid int) error {
	if err := t.setDefaultCgroupIfInvalid(); err != nil {
		return nil
	}

	proc := cgroups.Process{Subsystem: defaultSubsystem, Pid: pid}
	if err := t.cgroup.Add(proc); err != nil {
		return errors.Wrap(err, "failed to add process with pid '%d' to cgroup")
	}
	return nil
}

// SetLimits requires a pointer to a LinuxResources struct. The new
// limits cannot be less than the current resource usage of the tracked
// processes. This may return EBUSY if processes are running when this is
// called. If cgroups is not available, it no-ops.
func (t *linuxProcessTracker) SetLimits(limits interface{}) error {
	if !t.validCgroup() {
		return nil
	}

	linuxResourceLimits, ok := limits.(*LinuxResources)
	if !ok {
		return errors.New("process tracker requires (*LinuxResources) in order to update resource limits")
	}
	return t.doCgroupUpdateLimits(linuxResourceLimits)
}

func (t *linuxProcessTracker) doCgroupUpdateLimits(resourceLimits *LinuxResources) error {
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

// listCgroupPIDs lists all PIDs in the cgroup. If no cgroup is available, this
// returns a nil slice.
func (t *linuxProcessTracker) listCgroupPIDs() ([]int, error) {
	if !t.validCgroup() {
		return nil, nil
	}

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

// doCleanupByEnvironmentVariable terminates running processes whose
// value for environment variable ManagerEnvironID equals this process
// tracker's name (except for the current process).
func (t *linuxProcessTracker) doCleanupByEnvironmentVariable() error {
	myPid := os.Getpid()
	pids, err := getActivePIDs()
	if err != nil {
		return errors.Wrap(err, "could not get active PIDs")
	}

	catcher := grip.NewBasicCatcher()
	for _, pid := range pids {
		if pid == myPid {
			continue
		}

		env, err := getEnvironmentVariables(pid)
		if err != nil {
			continue
		}

		if env[ManagerEnvironID] == t.Name {
			osProc, err := os.FindProcess(pid)
			if err != nil {
				catcher.Add(errors.Wrap(err, "failed to tracked process"))
				continue
			}
			catcher.Add(errors.Wrapf(cleanupProcess(osProc), "error while cleaning up process with pid '%d'", pid))
		}
	}

	return catcher.Resolve()
}

// doCleanupByCgroup terminates running processes in this process tracker's
// cgroup.
func (t *linuxProcessTracker) doCleanupByCgroup() error {
	if !t.validCgroup() {
		return errors.New("cgroup is invalid so cannot cleanup by cgroup")
	}

	pids, err := t.listCgroupPIDs()
	if err != nil {
		return errors.Wrap(err, "could not find tracked processes")
	}

	catcher := grip.NewBasicCatcher()
	for _, pid := range pids {
		osProc, err := os.FindProcess(pid)
		if err != nil {
			catcher.Add(errors.Wrap(err, "failed to find tracked process"))
			continue
		}
		catcher.Add(errors.Wrapf(cleanupProcess(osProc), "error while cleaning up process with pid '%d'", pid))
	}

	// Delete the cgroup. If the process tracker is still used, the cgroup must
	// be re-initialized.
	catcher.Add(t.cgroup.Delete())
	return catcher.Resolve()
}

// Cleanup kills all tracked processes. If cgroups is available, it kills all
// processes in the cgroup. Otherwise, it kills processes based on the expected
// environment variable that should be set in all managed processes. This means
// that there should be an environment variable named ManagerEnvironID that has
// a value equal to this process tracker's name.
func (t *linuxProcessTracker) Cleanup() error {
	catcher := grip.NewBasicCatcher()
	if t.validCgroup() {
		catcher.Add(errors.Wrap(t.doCleanupByCgroup(), "error occurred while cleaning up processes tracked by cgroup"))
	}
	catcher.Add(errors.Wrap(t.doCleanupByEnvironmentVariable(), "error occurred while cleaning up processes tracked by environment variable"))

	return catcher.Resolve()
}
