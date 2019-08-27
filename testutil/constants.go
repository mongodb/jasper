package testutil

import "time"

const (
	TaskTimeout        = 5 * time.Second
	RPCTaskTimeout     = 30 * time.Second
	ProcessTestTimeout = 15 * time.Second
	ManagerTestTimeout = 5 * TaskTimeout
	LongTaskTimeout    = 100 * time.Second
)
