package jasper

import "syscall"

func makeMongodShutdownSignalTrigger() SignalTrigger {
	return func(_ ProcessInfo, _ syscall.Signal) bool {
		return false
	}
}
