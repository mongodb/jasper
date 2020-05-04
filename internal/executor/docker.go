package executor

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/evergreen-ci/utility"
	"github.com/google/uuid"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
)

// kim: TODO: needs thorough testing

type docker struct {
	execOpts types.ExecConfig
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer

	image    string
	platform string

	client     *client.Client
	httpClient *http.Client
	ctx        context.Context

	containerID    string
	containerMutex sync.RWMutex

	status      Status
	statusMutex sync.RWMutex

	pid      int
	exitCode int
	exitErr  error
	signal   syscall.Signal
}

// NewDocker returns an Executor that creates a process within a Docker
// container. Callers are expected to clean up resources by calling Close.
func NewDocker(ctx context.Context, client *client.Client, httpClient *http.Client, platform, image string, args []string) Executor {
	return &docker{
		ctx: ctx,
		execOpts: types.ExecConfig{
			Cmd: args,
		},
		platform:   platform,
		image:      image,
		client:     client,
		httpClient: httpClient,
		status:     Unstarted,
		pid:        -1,
		exitCode:   -1,
		signal:     syscall.Signal(-1),
	}
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
	if e.getStatus().After(Unstarted) {
		return errors.New("cannot start a process that has already started, exited, or closed")
	}

	if err := e.setupContainer(); err != nil {
		return errors.Wrap(err, "could not set up container for process")
	}

	if err := e.startContainer(); err != nil {
		return errors.Wrap(err, "could not start process within container")
	}

	e.setStatus(Running)

	return nil
}

func (e *docker) getContainerState() (*types.ContainerState, error) {
	resp, err := e.client.ContainerInspect(e.ctx, e.getContainerID())
	if err != nil {
		return nil, errors.Wrap(err, "could not inspect container")
	}
	if resp.ContainerJSONBase == nil || resp.ContainerJSONBase.State == nil {
		return nil, errors.Wrap(err, "introspection of container did not contain state information")
	}
	return resp.ContainerJSONBase.State, nil
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
	// pp.Println("removing container:", containerID)

	// We must ensure the container is cleaned up, so do not reuse the
	// Executor's context, which may already be done. This is an expensive
	// operation.
	rmCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.client.ContainerRemove(rmCtx, e.containerID, types.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return errors.Wrap(err, "problem cleaning up container for process")
	}
	// pp.Println("container removed:", containerID)

	e.setContainerID("")

	return nil
}

func (e *docker) Wait() error {
	if e.getStatus().Before(Running) {
		return errors.New("cannot wait on unstarted process")
	}
	if e.getStatus().After(Running) {
		return e.exitErr
	}

	containerID := e.getContainerID()

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
	}

	e.setStatus(Exited)

	return e.exitErr
}

func (e *docker) Signal(sig syscall.Signal) error {
	if e.getStatus() != Running {
		return errors.New("cannot signal a non-running process")
	}

	dsig, err := syscallToDockerSignal(sig, e.platform)
	if err != nil {
		return errors.Wrapf(err, "could not get Docker equivalent of signal '%d'", sig)
	}
	if err := e.client.ContainerKill(e.ctx, e.getContainerID(), dsig); err != nil {
		return errors.Wrap(err, "could not signal process within container")
	}

	e.signal = sig

	return nil
}

// PID returns the PID of the process in the container, or -1 if the PID cannot
// be retrieved.
func (e *docker) PID() int {
	if e.pid > -1 || !e.getStatus().BetweenInclusive(Running, Exited) {
		return e.pid
	}

	state, err := e.getContainerState()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message":   "could not get container state",
			"op":        "pid",
			"container": e.getContainerID(),
			"executor":  "Docker",
		}))
	}

	e.pid = state.Pid

	return e.pid
}

// ExitCode returns the exit code of the process in the container, or -1 if the
// exit code cannot be retrieved.
func (e *docker) ExitCode() int {
	if e.exitCode > -1 || !e.getStatus().BetweenInclusive(Running, Exited) {
		return e.exitCode
	}

	state, err := e.getContainerState()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message":   "could not get container state",
			"container": e.getContainerID(),
			"executor":  "docker",
		}))
		return e.exitCode
	}

	e.exitCode = state.ExitCode

	return e.exitCode
}

func (e *docker) Success() bool {
	if e.getStatus().Before(Exited) {
		return false
	}
	return e.exitErr == nil
}

// SignalInfo returns information about signals that were sent to the process in
// the container. This will only return information about received signals if
// Signal is called.
func (e *docker) SignalInfo() (sig syscall.Signal, signaled bool) {
	return e.signal, e.signal != -1
}

// Close cleans up the container associated with this process executor and
// closes the connection to the Docker daemon.
func (e *docker) Close() error {
	catcher := grip.NewBasicCatcher()
	catcher.Wrap(e.removeContainer(), "error removing Docker container")
	catcher.Wrap(e.client.Close(), "error closing Docker client")
	if e.httpClient != nil {
		utility.PutHTTPClient(e.httpClient)
	}
	e.setStatus(Closed)
	return catcher.Resolve()
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

func (e *docker) getStatus() Status {
	e.statusMutex.RLock()
	defer e.statusMutex.RUnlock()
	return e.status
}

func (e *docker) setStatus(status Status) {
	e.statusMutex.Lock()
	defer e.statusMutex.Unlock()
	if status < e.status && status != Unknown {
		return
	}
	e.status = status
}
