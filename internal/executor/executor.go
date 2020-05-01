package executor

import (
	"io"
	"syscall"
)

// Executor is an interface by which Jasper processes can manipulate and
// introspect on processes.
type Executor interface {
	// Args returns the command and the arguments used to create the process.
	Args() []string
	// SetEnv sets the process environment.
	SetEnv([]string)
	// Env returns the process environment.
	Env() []string
	// SetDir sets the process working directory.
	SetDir(string)
	// Dir returns the process working directory.
	Dir() string
	// SetStdin sets the process standard input.
	SetStdin(io.Reader)
	// SetStdout sets the process standard output.
	SetStdout(io.Writer)
	// Stdout returns the process standard output.
	Stdout() io.Writer
	// SetStderr sets the process standard error.
	SetStderr(io.Writer)
	// Stderr returns the process standard error.
	Stderr() io.Writer
	// Start begins execution of the process.
	Start() error
	// Wait waits for the process to complete.
	Wait() error
	// Signal sends a signal to a running process.
	Signal(syscall.Signal) error
	// PID returns the local process ID of the process if it is running or
	// complete. This is not guaranteed to return a valid value for remote
	// executors and will return -1 if it could not be retrieved.
	PID() int
	// ExitCode returns the exit code of a completed process. This will return a
	// non-negative value if it successfully retrieved the exit code.
	ExitCode() int
	// Success returns whether or not the completed process ran successfully.
	Success() bool
	// SignalInfo returns information about signals the process has received.
	SignalInfo() (sig syscall.Signal, signaled bool)
	// Close cleans up the executor's resources.
	Close()
}
