package jasper

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

// LinuxCPU represents limits for the CPU subsystem.
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

// LinuxWeightDevice struct holds a `major:minor weight` pair for weightDevice.
type LinuxWeightDevice struct {
	// Weight is the bandwidth rate for the device.
	Weight *uint16 `json:"weight,omitempty"`
	// LeafWeight is the bandwidth rate for the device while competing with the cgroup's child cgroups, CFQ scheduler only
	LeafWeight *uint16 `json:"leafWeight,omitempty"`
}

// LinuxThrottleDevice struct holds a `major:minor rate_per_second` pair.
type LinuxThrottleDevice struct {
	// Rate is the IO rate limit per cgroup per device
	Rate uint64 `json:"rate"`
}

// LinuxHugepageLimit
type LinuxHugepageLimit struct {
	// Pagesize is the hugepage size
	Pagesize string `json:"pageSize"`
	// Limit is the limit of "hugepagesize" hugetlb usage
	Limit uint64 `json:"limit"`
}

type LinuxNetwork struct {
	// Set class identifier for container's network packets
	ClassID *uint32 `json:"classID,omitempty"`
	// Set priority of network traffic for container
	Priorities []LinuxInterfacePriority `json:"priorities,omitempty"`
}

type LinuxInterfacePriority struct {
	// Name is the name of the network interface
	Name string `json:"name"`
	// Priority for the interface
	Priority uint32 `json:"priority"`
}

type LinuxRdma struct {
	// Maximum number of HCA handles that can be opened. Default is "no limit".
	HcaHandles *uint32 `json:"hcaHandles,omitempty"`
	// Maximum number of HCA objects that can be created. Default is "no limit".
	HcaObjects *uint32 `json:"hcaObjects,omitempty"`
}
