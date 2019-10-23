package executor

import (
	"context"
	"io"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
)

// SSH runs processes on a remote machine via SSH.
type SSH struct {
	session *ssh.Session
	args    []string
	ctx     context.Context
	exitErr error
}

// kim: TODO: have to establish connection and somehow set up closing the
// connection on ctx cancel.
func NewSSH(ctx context.Context, args []string) Executor {
	return &SSH{}
}

func (e *SSH) Args() []string {
	return e.args
}

func (e *SSH) SetEnv(env map[string]string) error {
	for k, v := range env {
		if err := e.session.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (e *SSH) Env() map[string]string {
	// kim: TODO or remove
	return nil
}

func (e *SSH) SetDir(dir string) {
	// kim: TODO
}

func (e *SSH) Dir() string {
	// kim: TODO
	return ""
}

func (e *SSH) SetStdin(stdin io.Reader) {
	e.session.Stdin = stdin
}

func (e *SSH) SetStdout(stdout io.Writer) {
	e.session.Stdout = stdout
}

func (e *SSH) SetStderr(stderr io.Writer) {
	e.session.Stderr = stderr
}

func (e *SSH) Start() error {
	return e.session.Run(strings.Join(e.args, " "))
}

func (e *SSH) Wait() error {
	return e.session.Wait()
}

func syscallToSSHSignal(sig syscall.Signal) ssh.Signal {
	switch sig {
	case syscall.SIGABRT:
		return ssh.SIGABRT
	case syscall.SIGALRM:
		return ssh.SIGALRM
	case syscall.SIGFPE:
		return ssh.SIGFPE
	case syscall.SIGHUP:
		return ssh.SIGHUP
	case syscall.SIGILL:
		return ssh.SIGILL
	case syscall.SIGINT:
		return ssh.SIGINT
	case syscall.SIGKILL:
		return ssh.SIGKILL
	case syscall.SIGPIPE:
		return ssh.SIGPIPE
	case syscall.SIGQUIT:
		return ssh.SIGQUIT
	case syscall.SIGSEGV:
		return ssh.SIGSEGV
	case syscall.SIGTERM:
		return ssh.SIGTERM
	case syscall.SIGUSR1:
		return ssh.SIGUSR1
	case syscall.SIGUSR2:
		return ssh.SIGUSR2
	}
	return ssh.Signal("")
}

func sshToSyscallSignal(sig ssh.Signal) syscall.Signal {
	switch sig {
	case ssh.SIGABRT:
		return syscall.SIGABRT
	case ssh.SIGALRM:
		return syscall.SIGALRM
	case ssh.SIGFPE:
		return syscall.SIGFPE
	case ssh.SIGHUP:
		return syscall.SIGHUP
	case ssh.SIGILL:
		return syscall.SIGILL
	case ssh.SIGINT:
		return syscall.SIGINT
	case ssh.SIGKILL:
		return syscall.SIGKILL
	case ssh.SIGPIPE:
		return syscall.SIGPIPE
	case ssh.SIGQUIT:
		return syscall.SIGQUIT
	case ssh.SIGSEGV:
		return syscall.SIGSEGV
	case ssh.SIGTERM:
		return syscall.SIGTERM
	case ssh.SIGUSR1:
		return syscall.SIGUSR1
	case ssh.SIGUSR2:
		return syscall.SIGUSR2
	}
	return syscall.Signal(0)
}

func (e *SSH) Signal(sig syscall.Signal) error {
	return e.session.Signal(syscallToSSHSignal(sig))
}

func (e *SSH) PID() int {
	return -1
}

func (e *SSH) ExitCode() int {
	if e.exitErr == nil {
		return -1
	}
	sshExitErr, ok := e.exitErr.(*ssh.ExitError)
	if !ok {
		return -1
	}
	return sshExitErr.Waitmsg.ExitStatus()
}

func (e *SSH) Success() bool {
	return e.exitErr == nil
}

func (e *SSH) SignalInfo() (sig syscall.Signal, signaled bool) {
	if e.exitErr == nil {
		return syscall.Signal(0), false
	}
	sshExitErr, ok := e.exitErr.(*ssh.ExitError)
	if !ok {
		return syscall.Signal(0), false
	}
	sshSig := ssh.Signal(sshExitErr.Waitmsg.Signal())
	return sshToSyscallSignal(sshSig), sshSig != ""
}
