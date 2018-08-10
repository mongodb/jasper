package jasper

import (
	"context"
	"syscall"
)

// TODO
//   - process tests
//   - manager tests
//   - helpers to configure output
//   - REST interface
//   - gRPC interface

type Process interface {
	ID() string
	Info(context.Context) ProcessInfo
	Running(context.Context) bool
	Complete(context.Context) bool
	Signal(context.Context, syscall.Signal) error
	Wait(context.Context) error
	RegisterTrigger(ProcessTrigger) error

	Tag(string)
	ResetTags()
	GetTags() []string
}

// Manager provides a basic, high level process management interface
type Manager interface {
	Create(context.Context, *CreateOptions) (Process, error)
	List(context.Context, Filter) ([]Process, error)
	Group(context.Context, string) ([]Process, error)
	Get(context.Context, string) (Process, error)
	Close(context.Context) error
}
