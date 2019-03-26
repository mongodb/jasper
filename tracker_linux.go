package jasper

import (
	"fmt"

	"github.com/containerd/cgroups"
	"github.com/pkg/errors"
)

// TODO: implement
// kim: TODO: allow cgroups to:
// 1. add procs to the list (for later killing).
// 2. impose limits (via a separate functionality).

type linuxProcessTracker struct {
	cgroups map[string]cgroups.Cgroup
}

func newProcessTracker(name string) (processTracker, error) {
	shares := uint64(100)
	cgroups := make(map[string]cgroups.Cgroup)

	cgroup, err := cgroups.New(cgroups.V1, cgroups.StaticPath("/jasper"), &specs.LinuxResources{
		CPU: &specs.CPU{
			Shares: &shares,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create cgroup")
	}
	fmt.Println(cgroup)

	return &linuxProcessTracker{cgroups: cgroups}, nil
}

func (t *linuxProcessTracker) add(pid uint) error {
	proc := cgroups.Process{}
	t.cgroup.Add(proc)
	return nil
}

// kim: TODO: addLimit for other OSes
// func (t *linuxProcessTracker) addLimit(pid uint, limit string) error {
//     return nil
// }

func (t *linuxProcessTracker) cleanup() error {
	// Get all procs.
	// t.cgroup.Processes("something", false)
	// Send SIGTERM to all procs.
	// for _, proc := range procs
}
