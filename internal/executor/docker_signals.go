package executor

import (
	"syscall"

	"github.com/pkg/errors"
)

// This file contains OS-dependent Docker signals taken from
// github.com/docker/docker/pkg/signal.

// syscallToDockerSignal converts the syscall.Signal to the equivalent Docker
// signal.
func syscallToDockerSignal(sig syscall.Signal, os string) (string, error) {
	switch os {
	case "darwin":
		if dsig, ok := syscallToDockerDarwin()[sig]; ok {
			return dsig, nil
		}
	case "linux":
		if dsig, ok := syscallToDockerLinux()[sig]; ok {
			return dsig, nil
		}
	case "windows":
		if dsig, ok := syscallToDockerWindows()[sig]; ok {
			return dsig, nil
		}
	default:
		return "", errors.Errorf("unrecognized OS '%s'", os)
	}
	return "", errors.Errorf("unrecognized Docker signal '%d' for OS '%s'", sig, os)
}

// These are constants taken from the signals in the syscall package for
// GOOS="linux".
const (
	linuxSIGABRT   = syscall.Signal(0x6)
	linuxSIGALRM   = syscall.Signal(0xe)
	linuxSIGBUS    = syscall.Signal(0x7)
	linuxSIGCHLD   = syscall.Signal(0x11)
	linuxSIGCLD    = syscall.Signal(0x11) // Synonym for SIGCHLD
	linuxSIGCONT   = syscall.Signal(0x12)
	linuxSIGFPE    = syscall.Signal(0x8)
	linuxSIGHUP    = syscall.Signal(0x1)
	linuxSIGILL    = syscall.Signal(0x4)
	linuxSIGINT    = syscall.Signal(0x2)
	linuxSIGIO     = syscall.Signal(0x1d)
	linuxSIGIOT    = syscall.Signal(0x6) // Synonym for SIGABRT
	linuxSIGKILL   = syscall.Signal(0x9)
	linuxSIGPIPE   = syscall.Signal(0xd)
	linuxSIGPOLL   = syscall.Signal(0x1d) // Synonym for SIGIO
	linuxSIGPROF   = syscall.Signal(0x1b)
	linuxSIGPWR    = syscall.Signal(0x1e)
	linuxSIGQUIT   = syscall.Signal(0x3)
	linuxSIGSEGV   = syscall.Signal(0xb)
	linuxSIGSTKFLT = syscall.Signal(0x10)
	linuxSIGSTOP   = syscall.Signal(0x13)
	linuxSIGSYS    = syscall.Signal(0x1f)
	linuxSIGTERM   = syscall.Signal(0xf)
	linuxSIGTRAP   = syscall.Signal(0x5)
	linuxSIGTSTP   = syscall.Signal(0x14)
	linuxSIGTTIN   = syscall.Signal(0x15)
	linuxSIGTTOU   = syscall.Signal(0x16)
	linuxSIGURG    = syscall.Signal(0x17)
	linuxSIGUSR1   = syscall.Signal(0xa)
	linuxSIGUSR2   = syscall.Signal(0xc)
	linuxSIGVTALRM = syscall.Signal(0x1a)
	linuxSIGWINCH  = syscall.Signal(0x1c)
	linuxSIGXCPU   = syscall.Signal(0x18)
	linuxSIGXFSZ   = syscall.Signal(0x19)
)

const (
	sigrtmin = 34
	sigrtmax = 64
)

func syscallToDockerLinux() map[syscall.Signal]string {
	return map[syscall.Signal]string{
		linuxSIGABRT: "ABRT",
		linuxSIGALRM: "ALRM",
		linuxSIGBUS:  "BUS",
		linuxSIGCHLD: "CHLD",
		// linuxSIGCLD:    "CLD",
		linuxSIGCONT: "CONT",
		linuxSIGFPE:  "FPE",
		linuxSIGHUP:  "HUP",
		linuxSIGILL:  "ILL",
		linuxSIGINT:  "INT",
		linuxSIGIO:   "IO",
		// linuxSIGIOT:    "IOT",
		linuxSIGKILL: "KILL",
		linuxSIGPIPE: "PIPE",
		// linuxSIGPOLL:   "POLL",
		linuxSIGPROF:   "PROF",
		linuxSIGPWR:    "PWR",
		linuxSIGQUIT:   "QUIT",
		linuxSIGSEGV:   "SEGV",
		linuxSIGSTKFLT: "STKFLT",
		linuxSIGSTOP:   "STOP",
		linuxSIGSYS:    "SYS",
		linuxSIGTERM:   "TERM",
		linuxSIGTRAP:   "TRAP",
		linuxSIGTSTP:   "TSTP",
		linuxSIGTTIN:   "TTIN",
		linuxSIGTTOU:   "TTOU",
		linuxSIGURG:    "URG",
		linuxSIGUSR1:   "USR1",
		linuxSIGUSR2:   "USR2",
		linuxSIGVTALRM: "VTALRM",
		linuxSIGWINCH:  "WINCH",
		linuxSIGXCPU:   "XCPU",
		linuxSIGXFSZ:   "XFSZ",
		sigrtmin:       "RTMIN",
		sigrtmin + 1:   "RTMIN+1",
		sigrtmin + 2:   "RTMIN+2",
		sigrtmin + 3:   "RTMIN+3",
		sigrtmin + 4:   "RTMIN+4",
		sigrtmin + 5:   "RTMIN+5",
		sigrtmin + 6:   "RTMIN+6",
		sigrtmin + 7:   "RTMIN+7",
		sigrtmin + 8:   "RTMIN+8",
		sigrtmin + 9:   "RTMIN+9",
		sigrtmin + 10:  "RTMIN+10",
		sigrtmin + 11:  "RTMIN+11",
		sigrtmin + 12:  "RTMIN+12",
		sigrtmin + 13:  "RTMIN+13",
		sigrtmin + 14:  "RTMIN+14",
		sigrtmin + 15:  "RTMIN+15",
		sigrtmax - 14:  "RTMAX-14",
		sigrtmax - 13:  "RTMAX-13",
		sigrtmax - 12:  "RTMAX-12",
		sigrtmax - 11:  "RTMAX-11",
		sigrtmax - 10:  "RTMAX-10",
		sigrtmax - 9:   "RTMAX-9",
		sigrtmax - 8:   "RTMAX-8",
		sigrtmax - 7:   "RTMAX-7",
		sigrtmax - 6:   "RTMAX-6",
		sigrtmax - 5:   "RTMAX-5",
		sigrtmax - 4:   "RTMAX-4",
		sigrtmax - 3:   "RTMAX-3",
		sigrtmax - 2:   "RTMAX-2",
		sigrtmax - 1:   "RTMAX-1",
		sigrtmax:       "RTMAX",
	}
}

// These are constants taken from the signals in the syscall package for
// GOOS="darwin".
const (
	darwinSIGABRT   = syscall.Signal(0x6)
	darwinSIGALRM   = syscall.Signal(0xe)
	darwinSIGBUS    = syscall.Signal(0xa)
	darwinSIGCHLD   = syscall.Signal(0x14)
	darwinSIGCONT   = syscall.Signal(0x13)
	darwinSIGEMT    = syscall.Signal(0x7)
	darwinSIGFPE    = syscall.Signal(0x8)
	darwinSIGHUP    = syscall.Signal(0x1)
	darwinSIGILL    = syscall.Signal(0x4)
	darwinSIGINFO   = syscall.Signal(0x1d)
	darwinSIGINT    = syscall.Signal(0x2)
	darwinSIGIO     = syscall.Signal(0x17)
	darwinSIGIOT    = syscall.Signal(0x6)
	darwinSIGKILL   = syscall.Signal(0x9)
	darwinSIGPIPE   = syscall.Signal(0xd)
	darwinSIGPROF   = syscall.Signal(0x1b)
	darwinSIGQUIT   = syscall.Signal(0x3)
	darwinSIGSEGV   = syscall.Signal(0xb)
	darwinSIGSTOP   = syscall.Signal(0x11)
	darwinSIGSYS    = syscall.Signal(0xc)
	darwinSIGTERM   = syscall.Signal(0xf)
	darwinSIGTRAP   = syscall.Signal(0x5)
	darwinSIGTSTP   = syscall.Signal(0x12)
	darwinSIGTTIN   = syscall.Signal(0x15)
	darwinSIGTTOU   = syscall.Signal(0x16)
	darwinSIGURG    = syscall.Signal(0x10)
	darwinSIGUSR1   = syscall.Signal(0x1e)
	darwinSIGUSR2   = syscall.Signal(0x1f)
	darwinSIGVTALRM = syscall.Signal(0x1a)
	darwinSIGWINCH  = syscall.Signal(0x1c)
	darwinSIGXCPU   = syscall.Signal(0x18)
	darwinSIGXFSZ   = syscall.Signal(0x19)
)

func syscallToDockerDarwin() map[syscall.Signal]string {
	return map[syscall.Signal]string{
		darwinSIGABRT: "ABRT",
		darwinSIGALRM: "ALRM",
		darwinSIGBUS:  "BUG",
		darwinSIGCHLD: "CHLD",
		darwinSIGCONT: "CONT",
		darwinSIGEMT:  "EMT",
		darwinSIGFPE:  "FPE",
		darwinSIGHUP:  "HUP",
		darwinSIGILL:  "ILL",
		darwinSIGINFO: "INFO",
		darwinSIGINT:  "INT",
		darwinSIGIO:   "IO",
		// darwinSIGIOT:    "IOT",
		darwinSIGKILL:   "KILL",
		darwinSIGPIPE:   "PIPE",
		darwinSIGPROF:   "PROF",
		darwinSIGQUIT:   "QUIT",
		darwinSIGSEGV:   "SEGV",
		darwinSIGSTOP:   "STOP",
		darwinSIGSYS:    "SYS",
		darwinSIGTERM:   "TERM",
		darwinSIGTRAP:   "TRAP",
		darwinSIGTSTP:   "TSTP",
		darwinSIGTTIN:   "TTIN",
		darwinSIGTTOU:   "TTOU",
		darwinSIGURG:    "URG",
		darwinSIGUSR1:   "USR1",
		darwinSIGUSR2:   "USR2",
		darwinSIGVTALRM: "VTALRM",
		darwinSIGWINCH:  "WINCH",
		darwinSIGXCPU:   "XCPU",
		darwinSIGXFSZ:   "XFSZ",
	}
}

// These are constants taken from the signals in the syscall package for
// GOOS="windows".
const (
	windowsSIGTERM = syscall.Signal(0x9)
	windowsSIGKILL = syscall.Signal(0xf)
)

func syscallToDockerWindows() map[syscall.Signal]string {
	return map[syscall.Signal]string{
		windowsSIGKILL: "KILL",
		windowsSIGTERM: "TERM",
	}
}
