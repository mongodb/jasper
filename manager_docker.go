package jasper

import (
	"context"

	docker "github.com/docker/docker/client"
	"github.com/mongodb/grip"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

type dockerManager struct {
	id      string
	procs   map[string]Process
	loggers LoggingCache
	client  *docker.Client
}

// NewDockerManager returns a manager that runs each process within a Docker
// container.
// func NewDockerManager(ctx context.Context) (Manager, error) {
// // kim: TODO: set up docker client from input parameters
// // kim: NOTE: ctx controls when docker manager goroutine shuts down.
// }

func (m *dockerManager) ID() string {
	return m.id
}

func (m *dockerManager) CreateProcess(ctx context.Context, opts *options.Create) (Process, error) {
	// kim: TODO: initialize docker client if it doesn't exist using the
	// original inputs to NewDockerManager.
	// kim: TODO: make sure to RegisterTrigger to destroy the process
	// kim: TODO: start container - maybe ContainerCreate, and keep track of
	// container ID
	// kim: TODO: start process - maybe ContainerExecCreate + ContainerExecStart
	return nil, errors.New("TODO: implement")
}

func (m *dockerManager) CreateCommand(ctx context.Context) *Command {
	return NewCommand().ProcConstructor(m.CreateProcess)
}

func (m *dockerManager) Register(ctx context.Context, proc Process) error {
	return errors.New("cannot register local processes in a Docker manager")
}

func (m *dockerManager) Get(ctx context.Context, id string) (Process, error) {
	return nil, errors.New("TODO: implement")
}

func (m *dockerManager) List(ctx context.Context, f options.Filter) ([]Process, error) {
	return nil, errors.New("TODO: implement")
}

func (m *dockerManager) Clear(ctx context.Context) {
	grip.Info("TODO: implement")
}

func (m *dockerManager) Close(ctx context.Context) error {
	// kim: TODO: kill the docker containers and remove them
	// kim: TODO: close the docker client, maybe
	return nil
}

// kim: TODO: do not hook up dockerProcess until we have some confidence that
// the docker executor works.
// type dockerProcess struct {
//     info ProcessInfo
//     client docker.Client
//     containerID string
// }

// func newDockerProcess(ctx context.Context, opts *options.Create) {
// kim: TODO: have to translate createoptions to equivalent
// kim: TODO: opts.Resolve() won't work since it's only meant for local, need to
// write custom Resolve method
// Jasper processes.
// kim: TODO: use ContainerCreate + ContainerStart
// }

// func (p *dockerProcess) Info(ctx context.Context) ProcessInfo {
//
// }

// func (p *dockerProcess) Running(ctx context.Context) {
//
// }

// func(p *dockerProcess) Complete(ctx context.Context) {
//
// }

// func (p *dockerProcess) Wait(ctx context.Context) (int, error) {
// // kim: TODO: find way to wait for process completion, maybe ContainerWait or
// ContainerInspect
// kim: NOTE: we can use ContainerInspect to sanity check that the container it
// out of the "created" state
// // kim: NOTE: make sure that we wait on a proper condition, i.e. using
// WaitNotRunning requires that the container is definitely out of the "created"
// phase. If we register a trigger at the end of the process to remove the
// container, it will be ensured to return on WaitConditionRemoved. However, we
// still must ensure that the executor's Wait() is not started in the "created"
// phase.
// }

// func (p *dockerProcess) Signal(ctx context.Context) (int, error) {
// // kim: TODO: find way to signal process - maybe ContainerKill
// }

// func (p *dockerProcess) RegisterSignalTrigger(ctx context.Context, SignalTrigger) error  {

// }

// func (p *dockerProcess) Respawn(ctx context.Context) (Process, error) {
//		kim: TODO: this may not be doable, unless we pass in a client or make a
//		client that can start a new container
// }

// func (p *dockerProcess) RegisterTrigger(ctx context.Context, t ProcessTrigger) error {

// }

// func (p *dockerProcess) Tag(t string) {

// }

// func (p *dockerProcess) ResetTags() {

// }

// func (p *dockerProcess) GetTags() []string {
//	return []string{"TODO"}
// }
