package executor

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

// Local runs processes on a local machine via exec.
type Local struct {
	*exec.Cmd
}

func newLocal(ctx context.Context, args []string) Executor {
	executable := args[0]
	if len(args) > 1 {
		args = args[1:]
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	return &Local{cmd}
}

func (e *Local) SetEnv(env map[string]string) error {
	if e.Env == nil {
		e.Env = []string{}
	}
	for k, v := range env {
		e.Env = append(e.Env, fmt.Sprintf("%s=%s", k, v))
	}
	return nil
}

func (e *Local) SetDirectory(dir string) {
	e.Dir = dir
}

func (e *Local) SetStdin(stdin io.Reader) {
	e.Stdin = stdin
}

func (e *Local) SetStdout(stdout io.Writer) {
	e.Stdout = stdout
}

func (e *Local) SetStderr(stderr io.Writer) {
	e.Stderr = stderr
}

func (e *Local) Start() error {
	return e.Start()
}

func (e *Local) Wait() error {
	return e.Wait()
}

func (e *Local) Signal(sig syscall.Signal) error {
	if e.Process == nil {
		return errors.New("cannot signal unstarted process")
	}
	return e.Process.Signal(sig)
}

func (e *Local) PID() int {
	if e.Process == nil {
		return -1
	}
	return e.Process.Pid
}

func (e *Local) ExitCode() int {
	if e.ProcessState == nil {
		return -1
	}
	status := e.ProcessState.Sys().(syscall.WaitStatus)
	return status.ExitStatus()
}

func (e *Local) Success() bool {
	if e.ProcessState == nil {
		return false
	}
	return e.ProcessState.Success()
}

func (e *Local) SignalInfo() (sig syscall.Signal, signaled bool) {
	if e.ProcessState == nil {
		return syscall.Signal(0), false
	}
	status := e.ProcessState.Sys().(syscall.WaitStatus)
	return status.Signal(), status.Signaled()
}
