package executor

import (
	"context"
	"io"
	"io/ioutil"
	"sync"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
)

type docker struct {
	execOpts types.ExecConfig
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer

	image    string
	platform string

	client    *client.Client
	ctx       context.Context
	closeConn context.CancelFunc

	containerID    string
	containerMutex sync.RWMutex
	execID         string
	execMutex      sync.RWMutex

	started  bool
	exited   bool
	exitErr  error
	exitCode int
}

// NewDocker returns an Executor that creates a process within a Docker
// container. Callers are expected to clean up resources by either cancelling
// the context or explicitly calling Close.
func NewDocker(ctx context.Context, opts []client.Opt, platform, image string, args []string) (Executor, error) {
	e := &docker{
		execOpts: types.ExecConfig{
			Cmd: args,
		},
		platform: platform,
		image:    image,
	}

	// kim: TODO: figure out why HTTP client doesn't work - get unable to connect
	// httpClient := utility.GetHTTPClient()
	// opts = append(opts, client.WithHTTPClient(httpClient))

	client, err := client.NewClientWithOpts(opts...)
	if err != nil {
		// utility.PutHTTPClient(httpClient)
		return nil, errors.Wrap(err, "could not create Docker client")
	}
	e.client = client

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-ctx.Done()
		grip.Error(errors.Wrap(e.client.Close(), "error closing Docker client"))
		grip.Error(e.removeContainer())
		// utility.PutHTTPClient(httpClient)
	}()
	e.ctx = ctx
	e.closeConn = cancel

	return e, nil
}

func (e *docker) Args() []string {
	return e.execOpts.Cmd
}

func (e *docker) SetEnv(env []string) {
	e.execOpts.Env = env
}

func (e *docker) Env() []string {
	return e.execOpts.Env
}

func (e *docker) SetDir(dir string) {
	e.execOpts.WorkingDir = dir
}

func (e *docker) Dir() string {
	return e.execOpts.WorkingDir
}

func (e *docker) SetStdin(stdin io.Reader) {
	e.stdin = stdin
	e.execOpts.AttachStdin = stdin != nil
}

func (e *docker) SetStdout(stdout io.Writer) {
	e.stdout = stdout
	e.execOpts.AttachStdout = stdout != nil
}

func (e *docker) Stdout() io.Writer {
	return e.stdout
}

func (e *docker) SetStderr(stderr io.Writer) {
	e.stderr = stderr
	e.execOpts.AttachStderr = stderr != nil
}

func (e *docker) Stderr() io.Writer {
	return e.stderr
}

func (e *docker) Start() error {
	if err := e.setupContainer(); err != nil {
		return errors.Wrap(err, "could not set up container for process")
	}

	if err := e.startContainer(); err != nil {
		return errors.Wrap(err, "could not start process within container")
	}

	e.started = true

	return nil
}

// setupContainer creates a container for the process without starting it.
func (e *docker) setupContainer() error {
	containerName := uuid.New().String()
	createResp, err := e.client.ContainerCreate(e.ctx, &container.Config{
		Image:        e.image,
		Cmd:          e.execOpts.Cmd,
		Env:          e.execOpts.Env,
		WorkingDir:   e.execOpts.WorkingDir,
		AttachStdin:  e.execOpts.AttachStdin,
		StdinOnce:    e.execOpts.AttachStdin,
		OpenStdin:    e.execOpts.AttachStdin,
		AttachStdout: e.execOpts.AttachStdout,
		AttachStderr: e.execOpts.AttachStderr,
	}, &container.HostConfig{}, &network.NetworkingConfig{}, containerName)
	if err != nil {
		return errors.Wrap(err, "problem creating container for process")
	}
	grip.WarningWhen(len(createResp.Warnings) != 0, message.Fields{
		"message":  "warnings during container creation for process",
		"warnings": createResp.Warnings,
	})

	e.setContainerID(createResp.ID)

	return nil
}

// startContainer attaches any I/O streams and starts the container.
func (e *docker) startContainer() error {
	if e.stdin != nil || e.stdout != nil || e.stderr != nil {
		stream, err := e.client.ContainerAttach(e.ctx, e.getContainerID(), types.ContainerAttachOptions{
			Stream: true,
			Stdin:  e.execOpts.AttachStdin,
			Stdout: e.execOpts.AttachStdout,
			Stderr: e.execOpts.AttachStderr,
		})
		if err != nil {
			return e.withRemoveContainer(errors.Wrap(err, "problem creating process within container"))
		}

		go func() {
			defer stream.Close()
			var wg sync.WaitGroup

			if e.stdin != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := io.Copy(stream.Conn, e.stdin)
					grip.Error(errors.Wrap(err, "problem streaming input to process"))
					grip.Error(errors.Wrap(stream.CloseWrite(), "problem closing process input stream"))
				}()
			}

			if e.stdout != nil || e.stderr != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					stdout := e.stdout
					stderr := e.stderr
					if stdout == nil {
						stdout = ioutil.Discard
					}
					if stderr == nil {
						stderr = ioutil.Discard
					}
					if _, err := stdcopy.StdCopy(stdout, stderr, stream.Reader); err != nil {
						grip.Error(errors.Wrap(err, "problem streaming output from process"))
					}
				}()
			}

			wg.Wait()
		}()
	}

	if err := e.client.ContainerStart(e.ctx, e.getContainerID(), types.ContainerStartOptions{}); err != nil {
		return e.withRemoveContainer(errors.Wrap(err, "problem starting container for process"))
	}

	return nil
}

// withRemoveContainer returns the error as well as any error from cleaning up
// the container.
func (e *docker) withRemoveContainer(err error) error {
	catcher := grip.NewBasicCatcher()
	catcher.Add(err)
	catcher.Add(e.removeContainer())
	return catcher.Resolve()
}

// removeContainer cleans up the container running this process.
func (e *docker) removeContainer() error {
	containerID := e.getContainerID()
	if containerID == "" {
		return nil
	}

	if err := e.client.ContainerRemove(e.ctx, e.containerID, types.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return errors.Wrap(err, "problem cleaning up container for process")
	}

	e.setContainerID("")

	return nil
}

func (e *docker) getContainerID() string {
	e.containerMutex.RLock()
	defer e.containerMutex.RUnlock()
	return e.containerID
}

func (e *docker) setContainerID(id string) {
	e.containerMutex.Lock()
	defer e.containerMutex.Unlock()
	e.containerID = id
}

func (e *docker) getExecID() string {
	e.execMutex.RLock()
	defer e.execMutex.RUnlock()
	return e.execID
}

func (e *docker) setExecID(id string) {
	e.execMutex.Lock()
	defer e.execMutex.Unlock()
	e.execID = id
}

func (e *docker) Wait() error {
	if !e.started {
		return errors.New("cannot wait on unstarted process")
	}
	if e.exited {
		return e.exitErr
	}

	containerID := e.getContainerID()

	// kim: TODO: verify that this is the correct condition to wait on. There
	// could be a race between container completion and this if we don't lock
	// the exited state.
	wait, errs := e.client.ContainerWait(e.ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errs:
		return errors.Wrap(err, "error waiting for container to finish running")
	case <-e.ctx.Done():
		return e.ctx.Err()
	case res := <-wait:
		if res.Error != nil {
			e.exitErr = errors.New(res.Error.Message)
		}
		e.exitCode = int(res.StatusCode)
	}

	e.exited = true

	return e.exitErr
}

func (e *docker) Signal(sig syscall.Signal) error {
	if !e.started {
		return errors.New("cannot signal an unstarted process")
	}
	if e.exited {
		return errors.New("cannot signal an exited process")
	}

	dsig, err := syscallToDockerSignal(sig, e.platform)
	if err != nil {
		return errors.Wrapf(err, "could not get Docker equivalent of signal '%d'", sig)
	}
	if err := e.client.ContainerKill(e.ctx, e.getExecID(), dsig); err != nil {
		return errors.Wrap(err, "could not signal process within container")
	}
	return nil
}

func (e *docker) PID() int {
	resp, err := e.client.ContainerExecInspect(e.ctx, e.getExecID())
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message":  "could not get exit code",
			"executor": "docker",
		}))
		return -1
	}
	// kim: TODO: test if PID is still valid even when process has exited. If
	// not, it may be necessarily a real pid.
	return resp.Pid
}

// ExitCode returns, or -1 if it cannot be retrieved.
func (e *docker) ExitCode() int {
	if !e.exited {
		return -1
	}
	return e.exitCode

	// resp, err := e.client.ContainerExecInspect(e.ctx, e.getExecID())
	// if err != nil {
	//     grip.Error(message.WrapError(err, message.Fields{
	//         "message":  "could not get exit code",
	//         "executor": "docker",
	//     }))
	//     return -1
	// }
	// if resp.Running {
	//     return -1
	// }
	//
	// return resp.ExitCode
}

func (e *docker) Success() bool {
	if !e.started || !e.exited {
		return false
	}
	return e.exitErr == nil
}

func (e *docker) SignalInfo() (sig syscall.Signal, signaled bool) {
	return -1, false
}

func (e *docker) Close() {
	e.closeConn()
}
