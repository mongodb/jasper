package jasper

import (
	"syscall"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
)

const cleanTerminationSignalTriggerSource = "clean termination trigger"

// makeCleanTerminationSignalTrigger terminates a process so that it will return exit code 0.
func makeCleanTerminationSignalTrigger() SignalTrigger {
	return func(info ProcessInfo, sig syscall.Signal) bool {
		if sig != syscall.SIGTERM {
			return false
		}

		proc, err := OpenProcess(procRightTerminate|procRightQueryInformation, false, uint32(info.PID))
		if err != nil {
			// OpenProcess returns errInvalidParameter if the process has already exited.
			if err == errInvalidParameter {
				grip.Debug(message.WrapError(err, message.Fields{
					"id":      info.ID,
					"pid":     info.PID,
					"source":  cleanTerminationSignalTriggerSource,
					"message": "did not open process because it has already exited",
				}))
			} else {
				grip.Error(message.WrapError(err, message.Fields{
					"id":      info.ID,
					"pid":     info.PID,
					"source":  cleanTerminationSignalTriggerSource,
					"message": "failed to open process",
				}))
			}
			return false
		}
		defer func() {
			if err := NewWindowsError("CloseHandle", CloseHandle(proc)); err != nil {
				grip.Warning(message.WrapError(err, message.Fields{
					"message": "failed to close job object handle",
					"id":      info.ID,
					"pid":     info.PID,
					"source":  cleanTerminationSignalTriggerSource,
				}))
			}
		}()

		if err := TerminateProcess(proc, 0); err != nil {
			// TerminateProcess returns errAccessDenied if the process has
			// already died.
			if err != errAccessDenied {
				grip.Error(message.WrapError(err, message.Fields{
					"id":      info.ID,
					"pid":     info.PID,
					"source":  cleanTerminationSignalTriggerSource,
					"message": "failed to terminate process",
				}))
				return false
			}

			var exitCode uint32
			err := GetExitCodeProcess(proc, &exitCode)
			if err != nil {
				grip.Error(message.WrapError(err, message.Fields{
					"id":      info.ID,
					"pid":     info.PID,
					"source":  cleanTerminationSignalTriggerSource,
					"message": "terminate process was sent but failed to get exit code",
				}))
				return false
			}
			if exitCode == procStillActive {
				grip.Error(message.WrapError(err, message.Fields{
					"id":      info.ID,
					"pid":     info.PID,
					"source":  cleanTerminationSignalTriggerSource,
					"message": "terminate process was sent but process is still active",
				}))
				return false
			}
		}

		return true
	}
}
