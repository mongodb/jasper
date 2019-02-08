package jasper

import (
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

const (
	// The name for the mongod termination event has the format "Global\Mongo_<pid>".
	mongodShutdownEventNamePrefix = "Global\\Mongo_"
	mongodShutdownEventTimeout    = 60 * time.Second
)

// makeMongodEventTrigger is necessary to support clean termination of mongods on Windows by signaling the mongod
// shutdown event and waiting for the process to return.
func makeMongodShutdownSignalTrigger() SignalTrigger {
	return func(info ProcessInfo, sig syscall.Signal) (skipSignal bool) {
		skipSignal = false

		// Only run if the program is mongod.
		if len(info.Options.Args) == 0 || !strings.Contains(filepath.Base(info.Options.Args[0]), "mongod") {
			return
		}
		// Only care about termination signals.
		if sig != syscall.SIGKILL && sig != syscall.SIGTERM {
			return
		}

		pid := info.PID
		eventName := mongodShutdownEventNamePrefix + strconv.Itoa(pid)
		utf16EventName, err := syscall.UTF16PtrFromString(eventName)
		if err != nil {
			grip.Errorf(errors.Wrapf(err, "failed to convert mongod shutdown event name '%s'", eventName).Error())
			return
		}

		event, err := OpenEvent(utf16EventName)
		if err != nil {
			grip.Errorf(errors.Wrapf(err, "failed to open event '%s'", eventName).Error())
			return
		}
		defer CloseHandle(event)

		if err := SetEvent(event); err != nil {
			grip.Errorf(errors.Wrapf(err, "failed to signal event '%s'", eventName).Error())
			return
		}

		mongodProc, err := OpenProcess(PROCESS_ALL_ACCESS, false, uint32(pid))
		if err != nil {
			grip.Errorf(errors.Wrapf(err, "failed to get process handle for pid %d", pid).Error())
			return
		}
		defer CloseHandle(mongodProc)

		grip.Infof("mongod with pid %d received termination signal %d and was cleanly shut down", pid, sig)
		return true
	}
}
