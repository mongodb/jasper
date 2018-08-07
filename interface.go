package jasper

import (
	"context"
	"syscall"

	"github.com/pkg/errors"
)

type Process interface {
	ID() string
	Info(context.Context) ProcessInfo
	Running(context.Context) bool
	Signal(context.Context, syscall.Siginal) error
	Wait(context.Context) error
}

type ProcessInfo struct {
	Host      string
	PID       int
	IsRunning bool
	ID        string
	Options   CreateOptions
}

type CreateOptions struct {
	Args             []string
	Environment      map[string]string
	WorkingDirectory string
	Output           OutputOptions
}

type Manager interface {
	Create(context.Context, CreateOptions) (Process, error)

	List(context.Context) ([]Process, error)
	Get(context.Context, string) (Process, error)

	Terminate(context.Context, string) error
	TerminateAll(context.Context) error
	Close(context.Context) error
}

type Filter string

const (
	Running    Filter = "running"
	Terminated        = "terminated"
)

func (f Filter) Validate() error {
	switch f {
	case Running, Terminated:
		return nil
	default:
		return errors.Errorf("%s is not a valid filter", f)
	}

}

////////////////////////////////////////////////////////////////////////

type MockProcessManager struct{}

type remoteProcessManager struct{}

type localProcessManager struct{}

type basicProcessManager struct {
	procs map[string]localProcess
}
