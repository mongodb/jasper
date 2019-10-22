package executor

import (
	"io"
	"syscall"
)

// Executor is an interface by which Jasper processes can manipulate and
// introspect on operating system process.
type Executor interface {
	SetEnv(map[string]string) error
	SetDirectory(string)
	SetStdin(io.Reader)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	Start() error
	Wait() error
	Signal(syscall.Signal) error
	PID() int
	ExitCode() int
	Success() bool
	SignalInfo() (sig syscall.Signal, signaled bool)
}
