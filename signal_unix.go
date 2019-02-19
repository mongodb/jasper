// +build darwin linux

package jasper

import "syscall"

func modifySignal(sig syscall.Signal) syscall.Signal {
	return sig
}
