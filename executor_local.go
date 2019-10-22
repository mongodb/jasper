package jasper

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

// executorLocal is an Executor running processes on a local machine.
type executorLocal struct {
	*exec.Cmd
}

func newLocalExecutor(ctx context.Context, args []string) Executor {
	executable := args[0]
	if len(args) > 1 {
		args = args[1:]
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	return &executorLocal(cmd)
}

func (e *executorLocal) SetEnv(env map[string]string) error {
	if e.Env == nil {
		e.Env = []string{}
	}
	for k, v := range env {
		e.Env = append(e.Env, fmt.Sprintf("%s=%s", k, v))
	}
	return nil
}

func (e *executorLocal) SetWorkingDirectory(dir string) {
	e.WorkingDirectory = dir
}

func (e *executorLocal) SetStdin(stdin io.Reader) {
	e.Stdin = stdin
}

func (e *executorLocal) SetStdout(stdout io.Writer) {
	e.Stdout = stdout
}

func (e *executorLocal) SetStderr(stderr io.Writer) {
	e.Stderr = stderr
}

func (e *executorLocal) Start() error {
	return e.Start()
}

func (e *executorLocal) Wait() error {
	return e.Wait()
}

func (e *executorLocal) Signal(sig syscall.Signal) error {
	if e.Process == nil {
		return errors.New("cannot signal unstarted process")
	}
	return e.Process.Signal(sig)
}

func (e *executorLocal) PID() int {
	if e.Process == nil {
		return -1
	}
	return e.Process.Pid
}

func (e *executorLocal) ExitCode() int {
	if cmd.ProcessState == nil {
		return -1
	}
	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	return status.ExitStatus()
}

func (e *executorLocal) Success() bool {
	if e.ProcessState == nil {
		return false
	}
	return e.ProcessState.Success()
}

func (e *executorLocal) SignalInfo() (sig syscall.Signal, signaled bool) {
	if cmd.ProcessState == nil {
		return syscall.Signal(0), false
	}
	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	return status.Signal(), status.Signaled()
}
