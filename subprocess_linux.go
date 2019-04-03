package jasper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// Documentation for each subsystem can be found in the kernel docs:
// https://www.kernel.org/doc/Documentation/cgroup-v1

// LinuxResources represents runtime resource constraints.
type LinuxResources struct {
	// Devices configures the device whitelist.
	Devices []LinuxDeviceCgroup `json:"devices,omitempty"`
	// Memory restriction configuration
	Memory *LinuxMemory `json:"memory,omitempty"`
	// CPU resource restriction configuration
	CPU *LinuxCPU `json:"cpu,omitempty"`
	// Task resource restriction configuration.
	Pids *LinuxPids `json:"pids,omitempty"`
	// BlockIO restriction configuration
	BlockIO *LinuxBlockIO `json:"blockIO,omitempty"`
	// Hugetlb limit (in bytes)
	HugepageLimits []LinuxHugepageLimit `json:"hugepageLimits,omitempty"`
	// Network restriction configuration
	Network *LinuxNetwork `json:"network,omitempty"`
	// Rdma resource restriction configuration.
	// Limits are a set of key value pairs that define RDMA resource limits,
	// where the key is device name and value is resource limits.
	Rdma map[string]LinuxRdma `json:"rdma,omitempty"`
}

// LinuxDeviceCgroup represents a device rule for the whitelist controller.
type LinuxDeviceCgroup struct {
	// Allow or deny
	Allow bool `json:"allow"`
	// Device type, block, char, etc.
	Type string `json:"type,omitempty"`
	// Major is the device's major number.
	Major *int64 `json:"major,omitempty"`
	// Minor is the device's minor number.
	Minor *int64 `json:"minor,omitempty"`
	// Cgroup access permissions format, rwm.
	Access string `json:"access,omitempty"`
}

// LinuxMemory represents limits for the memory subsystem.
type LinuxMemory struct {
	// Memory limit (in bytes).
	Limit *int64 `json:"limit,omitempty"`
	// Memory reservation or soft_limit (in bytes).
	Reservation *int64 `json:"reservation,omitempty"`
	// Total memory limit (memory + swap).
	Swap *int64 `json:"swap,omitempty"`
	// Kernel memory limit (in bytes).
	Kernel *int64 `json:"kernel,omitempty"`
	// Kernel memory limit for tcp (in bytes)
	KernelTCP *int64 `json:"kernelTCP,omitempty"`
	// How aggressive the kernel will swap memory pages.
	Swappiness *uint64 `json:"swappiness,omitempty"`
	// DisableOOMKiller disables the OOM killer for out of memory conditions
	DisableOOMKiller *bool `json:"disableOOMKiller,omitempty"`
}

// LinuxCPU represents limits for the cpuacct,cpuset subsystems.
type LinuxCPU struct {
	// CPU shares (relative weight (ratio) vs. other cgroups with cpu shares).
	Shares *uint64 `json:"shares,omitempty"`
	// CPU hardcap limit (in usecs). Allowed cpu time in a given period.
	Quota *int64 `json:"quota,omitempty"`
	// CPU period to be used for hardcapping (in usecs).
	Period *uint64 `json:"period,omitempty"`
	// How much time realtime scheduling may use (in usecs).
	RealtimeRuntime *int64 `json:"realtimeRuntime,omitempty"`
	// CPU period to be used for realtime scheduling (in usecs).
	RealtimePeriod *uint64 `json:"realtimePeriod,omitempty"`
	// CPUs to use within the cpuset. Default is to use any CPU available.
	Cpus string `json:"cpus,omitempty"`
	// List of memory nodes in the cpuset. Default is to use any available memory node.
	Mems string `json:"mems,omitempty"`
}

// LinuxPids represents limits for the pid subsystem.
type LinuxPids struct {
	// Maximum number of PIDs. Default is "no limit".
	Limit int64 `json:"limit"`
}

// LinuxBlockIO represents limits for the blkio subsystem.
type LinuxBlockIO struct {
	// Specifies per cgroup weight
	Weight *uint16 `json:"weight,omitempty"`
	// Specifies tasks' weight in the given cgroup while competing with the cgroup's child cgroups, CFQ scheduler only
	LeafWeight *uint16 `json:"leafWeight,omitempty"`
	// Weight per cgroup per device, can override BlkioWeight
	WeightDevice []LinuxWeightDevice `json:"weightDevice,omitempty"`
	// IO read rate limit per cgroup per device, bytes per second
	ThrottleReadBpsDevice []LinuxThrottleDevice `json:"throttleReadBpsDevice,omitempty"`
	// IO write rate limit per cgroup per device, bytes per second
	ThrottleWriteBpsDevice []LinuxThrottleDevice `json:"throttleWriteBpsDevice,omitempty"`
	// IO read rate limit per cgroup per device, IO per second
	ThrottleReadIOPSDevice []LinuxThrottleDevice `json:"throttleReadIOPSDevice,omitempty"`
	// IO write rate limit per cgroup per device, IO per second
	ThrottleWriteIOPSDevice []LinuxThrottleDevice `json:"throttleWriteIOPSDevice,omitempty"`
}

// LinuxWeightDevice represents the weighting for time-based division of disk policy.
type LinuxWeightDevice struct {
	// Weight is the bandwidth rate for the device.
	Weight *uint16 `json:"weight,omitempty"`
	// LeafWeight is the bandwidth rate for the device while competing with the cgroup's child cgroups, CFQ scheduler only
	LeafWeight *uint16 `json:"leafWeight,omitempty"`
}

// LinuxThrottleDevice represents the upper limit on I/O rate to a device.
type LinuxThrottleDevice struct {
	// Rate is the IO rate limit per cgroup per device
	Rate uint64 `json:"rate"`
}

// LinuxHugepageLimit represents limits for the hugetlb subsystem.
type LinuxHugepageLimit struct {
	// Pagesize is the hugepage size
	Pagesize string `json:"pageSize"`
	// Limit is the limit of "hugepagesize" hugetlb usage
	Limit uint64 `json:"limit"`
}

// LinuxNetwork represents priorities for the net_cls,net_prio subsystems.
type LinuxNetwork struct {
	// Set class identifier for container's network packets
	ClassID *uint32 `json:"classID,omitempty"`
	// Set priority of network traffic for container
	Priorities []LinuxInterfacePriority `json:"priorities,omitempty"`
}

// LinuxInterfacePriority represents a priority on a network interface.
type LinuxInterfacePriority struct {
	// Name is the name of the network interface
	Name string `json:"name"`
	// Priority for the interface
	Priority uint32 `json:"priority"`
}

// LinuxRdma represents limits on the rdma subsystem.
type LinuxRdma struct {
	// Maximum number of HCA handles that can be opened. Default is "no limit".
	HcaHandles *uint32 `json:"hcaHandles,omitempty"`
	// Maximum number of HCA objects that can be created. Default is "no limit".
	HcaObjects *uint32 `json:"hcaObjects,omitempty"`
}

// getActivePIDs returns a list of active PIDs on the system by listing the
// contents of /proc and looking for entries that appear to be valid PIDs.
func getActivePIDs() ([]int, error) {
	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	results := make([]int, 0, 50)
	for {
		fileInfos, err := d.Readdir(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, fileInfo := range fileInfos {
			// PID must be a directory with a numeric name.
			if !fileInfo.IsDir() {
				continue
			}

			// Using Atoi here will also filter out "." and "..".
			pid, err := strconv.Atoi(fileInfo.Name())
			if err != nil {
				continue
			}
			results = append(results, pid)
		}
	}
	return results, nil
}

// getEnvironmentVariables returns a map of environment variables for the
// given pid. This function works by reading from /proc/${pid}/environ, so the
// entries returned only reflect the values of the environment variables at the
// time that the process was started.
func getEnvironmentVariables(pid int) (map[string]string, error) {
	rawEnv, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/environ", pid))
	if err != nil {
		// This is probably either "permission denied" because we do not own the process,
		// or the process simply doesn't exist anymore.
		return nil, err
	}
	parts := bytes.Split(rawEnv, []byte{0})

	env := map[string]string{}
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		nameAndValue := strings.Split(string(part), "=")
		if len(nameAndValue) == 1 {
			env[nameAndValue[0]] = ""
		} else if len(nameAndValue) == 2 {
			env[nameAndValue[0]] = nameAndValue[1]
		}
	}

	return env, nil
}

// cleanupProcess kills the given process.
func cleanupProcess(proc *os.Process) error {
	catcher := grip.NewBasicCatcher()
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		catcher.Add(errors.Wrap(err, "could not signal process to terminate"))
		catcher.Add(errors.Wrap(proc.Kill(), "could not kill process"))
	}
	return catcher.Resolve()
}
