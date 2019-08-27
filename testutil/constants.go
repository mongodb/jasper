package testutil

import "time"

const (
	TestTimeoput       = 5 * time.Second
	RPCTestTimeout     = 30 * time.Second
	ProcessTestTimeout = 15 * time.Second
	ManagerTestTimeout = 5 * TestTimeoput
	LongTestTimeout    = 100 * time.Second
)
