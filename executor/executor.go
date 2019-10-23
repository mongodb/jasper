package executor

import (
	"io"
	"syscall"
)

// Executor is an interface by which Jasper processes can manipulate and
// introspect on operating system process.
// kim: TODO: make this pass tests, for some reason causes stack overflow in
// process tests.
type Executor interface {
	Args() []string
	SetEnv(map[string]string) error
	Env() map[string]string
	SetDir(string)
	Dir() string
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
