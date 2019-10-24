package executor

import (
	"context"
	"io"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

// execLocal runs processes on a local machine via exec.
type execLocal struct {
	cmd  *exec.Cmd
	args []string
}

// NewexecLocal returns an Executor that creates processes locally.
func NewLocal(ctx context.Context, args []string) Executor {
	executable := args[0]
	var execArgs []string
	if len(args) > 1 {
		execArgs = args[1:]
	}
	cmd := exec.CommandContext(ctx, executable, execArgs...)
	return &execLocal{cmd: cmd, args: args}
}

// MakeLocal wraps an existing local process.
func MakeLocal(cmd *exec.Cmd) Executor {
	return &execLocal{
		cmd:  cmd,
		args: cmd.Args,
	}
}

// Args returns the arguments to the command.
func (e *execLocal) Args() []string {
	return e.args
}

func (e *execLocal) SetEnv(env []string) {
	e.cmd.Env = env
	return nil
}

func (e *execLocal) Env() []string {
	return e.cmd.Env
}

func (e *execLocal) SetDir(dir string) {
	e.cmd.Dir = dir
}

func (e *execLocal) Dir() string {
	return e.cmd.Dir
}

func (e *execLocal) SetStdin(stdin io.Reader) {
	e.cmd.Stdin = stdin
}

func (e *execLocal) SetStdout(stdout io.Writer) {
	e.cmd.Stdout = stdout
}

func (e *execLocal) SetStderr(stderr io.Writer) {
	e.cmd.Stderr = stderr
}

func (e *execLocal) Start() error {
	return e.cmd.Start()
}

func (e *execLocal) Wait() error {
	return e.cmd.Wait()
}

func (e *execLocal) Signal(sig syscall.Signal) error {
	if e.cmd.Process == nil {
		return errors.New("cannot signal unstarted process")
	}
	return e.cmd.Process.Signal(sig)
}

func (e *execLocal) PID() int {
	if e.cmd.Process == nil {
		return -1
	}
	return e.cmd.Process.Pid
}

func (e *execLocal) ExitCode() int {
	if e.cmd.ProcessState == nil {
		return -1
	}
	status := e.cmd.ProcessState.Sys().(syscall.WaitStatus)
	return status.ExitStatus()
}

func (e *execLocal) Success() bool {
	if e.cmd.ProcessState == nil {
		return false
	}
	return e.cmd.ProcessState.Success()
}

func (e *execLocal) SignalInfo() (sig syscall.Signal, signaled bool) {
	if e.cmd.ProcessState == nil {
		return syscall.Signal(0), false
	}
	status := e.cmd.ProcessState.Sys().(syscall.WaitStatus)
	return status.Signal(), status.Signaled()
}

// CLose is a no-op.
func (e *execLocal) Close() {
	return nil
}
