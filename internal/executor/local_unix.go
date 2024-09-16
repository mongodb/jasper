//go:build unix

package executor

import "syscall"

// SetGroupLeader sets this process as a group leader.
func (e *local) SetGroupLeader() {
	e.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
