package executor

import (
	"context"
	"io"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

// Local runs processes on a local machine via exec.
type Local struct {
	cmd  *exec.Cmd
	args []string
}

func NewLocal(ctx context.Context, args []string) Executor {
	executable := args[0]
	var execArgs []string
	if len(args) > 1 {
		execArgs = args[1:]
	}
	cmd := exec.CommandContext(ctx, executable, execArgs...)
	return &Local{cmd: cmd, args: args}
}

func MakeLocal(cmd *exec.Cmd) Executor {
	return &Local{
		cmd:  cmd,
		args: cmd.Args,
	}
}

func (e *Local) Args() []string {
	return e.args
}

func (e *Local) SetEnv(env []string) error {
	e.cmd.Env = env
	return nil
}

func (e *Local) Env() []string {
	return e.cmd.Env
}

func (e *Local) SetDir(dir string) {
	e.cmd.Dir = dir
}

func (e *Local) Dir() string {
	return e.cmd.Dir
}

func (e *Local) SetStdin(stdin io.Reader) {
	e.cmd.Stdin = stdin
}

func (e *Local) SetStdout(stdout io.Writer) {
	e.cmd.Stdout = stdout
}

func (e *Local) SetStderr(stderr io.Writer) {
	e.cmd.Stderr = stderr
}

func (e *Local) Start() error {
	return e.cmd.Start()
}

func (e *Local) Wait() error {
	return e.cmd.Wait()
}

func (e *Local) Signal(sig syscall.Signal) error {
	if e.cmd.Process == nil {
		return errors.New("cannot signal unstarted process")
	}
	return e.cmd.Process.Signal(sig)
}

func (e *Local) PID() int {
	if e.cmd.Process == nil {
		return -1
	}
	return e.cmd.Process.Pid
}

func (e *Local) ExitCode() int {
	if e.cmd.ProcessState == nil {
		return -1
	}
	status := e.cmd.ProcessState.Sys().(syscall.WaitStatus)
	return status.ExitStatus()
}

func (e *Local) Success() bool {
	if e.cmd.ProcessState == nil {
		return false
	}
	return e.cmd.ProcessState.Success()
}

func (e *Local) SignalInfo() (sig syscall.Signal, signaled bool) {
	if e.cmd.ProcessState == nil {
		return syscall.Signal(0), false
	}
	status := e.cmd.ProcessState.Sys().(syscall.WaitStatus)
	return status.Signal(), status.Signaled()
}
