package jasper

import "syscall"

func modifySignal(sig syscall.Signal) syscall.Signal {
	if sig == syscall.SIGTERM {
		return syscall.SIGKILL
	}
	return sig
}
