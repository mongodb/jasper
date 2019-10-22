package jasper

// executorSSH is an Executor running processes on a remote machine via SSH.
type executorSSH {
	*ssh.Session
	args []string
	ctx context.Context
	exitErr error
}

// kim: TODO: have to establish connection and somehow set up closing the
// connection on ctx cancel.
func newSSHExecutor(ctx context.Context, args []string) Executor {

}

func (e *executorSSH) SetEnv(env map[string]string) error {
	for k, v := range env {
		if err := e.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (e *executorSSH) SetWorkingDirectory(dir string) {
	// kim: TODO
}

func (e *executorSSH) SetStdin(stdin io.Reader) {
	e.Stdin = stdin
}

func (e *executorLocal) SetStdout(stdout io.Writer) {
	e.Stdout = stdout
}

func (e *executorLocal) SetStderr(stderr io.Writer) {
	e.Stderr = stderr
}

func (e *executorLocal) Start() error {
	return e.Run(e.args)
}

func (e *executorLocal) Wait() error {
	return e.Wait()
}

func (e *executorLocal) Signal(sig syscall.Signal) error {
	e.Signal(ssh.Signal(sig))
}

func (e *executorLocal) PID() int {
	return -1
}

func (e *executorLocal) ExitCode() int {
	if e.exitErr == nil {
		return -1
	}
	sshExitErr, ok := e.exitErr.(ssh.ExitError)
	if !ok {
		return -1
	}
	return sshExitErr.Waitmsg.ExitStatus()
}

func (e *executorLocal) Success() bool {
	return sshExitErr == nil
}

func (e *executorLocal) SignalInfo() (sig syscall.Signal, signaled bool) {
	if e.exitErr == nil {
		return syscall.Signal(0), false
	}
	sshExitErr, ok := e.exitErr.(ssh.ExitError)
	if !ok {
		return syscall.Signal(0), false
	}
	return syscall.Signal(sshExitErr.Waitmsg.Signal()), sshExitErr.Waitmsg.Signaled()
}
