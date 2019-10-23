package executor

import (
	"context"
	"fmt"
	"io"
	"strings"
	"syscall"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// SSH runs processes on a remote machine via SSH.
type SSH struct {
	session *ssh.Session
	client  *ssh.Client
	args    []string
	dir     string
	env     []string
	exited  bool
	exitErr error
	// kim: TOD: requires cancel func to manually stop goroutine
	closeConn context.CancelFunc
}

func MakeSSH(ctx context.Context, client *ssh.Client, session *ssh.Session, args []string) Executor {
	ctx, cancel := context.WithCancel(ctx)
	e := &SSH{session: session, client: client, args: args}
	e.closeConn = cancel
	go func() {
		<-ctx.Done()
		if err := e.session.Close(); err != nil && err != io.EOF {
			grip.Warning(errors.Wrap(err, "error closing SSH session"))
		}
		if err := e.client.Close(); err != nil && err != io.EOF {
			grip.Warning(errors.Wrap(err, "error closing SSH client"))
		}
	}()
	return e
}

func (e *SSH) Args() []string {
	return e.args
}

func (e *SSH) SetEnv(env []string) error {
	e.env = env
	return nil
}

func (e *SSH) Env() []string {
	return e.env
}

func (e *SSH) SetDir(dir string) {
	e.dir = dir
}

func (e *SSH) Dir() string {
	return e.dir
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
	args := []string{}
	for _, entry := range e.env {
		args = append(args, fmt.Sprintf("export %s", entry))
	}
	if e.dir != "" {
		args = append(args, fmt.Sprintf("cd %s", e.dir))
	}
	args = append(args, strings.Join(e.args, " "))
	return e.session.Start(strings.Join(args, "\n"))
}

func (e *SSH) Wait() error {
	defer e.closeConn()
	catcher := grip.NewBasicCatcher()
	e.exitErr = e.session.Wait()
	catcher.Add(e.exitErr)
	e.exited = true
	return catcher.Resolve()
}

func (e *SSH) Signal(sig syscall.Signal) error {
	return e.session.Signal(syscallToSSHSignal(sig))
}

func (e *SSH) PID() int {
	// There is no simple way of retrieving the PID of the remote process.
	return -1
}

func (e *SSH) ExitCode() int {
	if !e.exited {
		return -1
	}
	if e.exitErr == nil {
		return 0
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
		return syscall.Signal(-1), false
	}
	sshExitErr, ok := e.exitErr.(*ssh.ExitError)
	if !ok {
		return syscall.Signal(-1), false
	}
	sshSig := ssh.Signal(sshExitErr.Waitmsg.Signal())
	return sshToSyscallSignal(sshSig), sshSig != ""
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
