package jasper

import (
	"context"
	"runtime"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// MockManager implements the Manager interface with exported fields to
// configure and introspect the mock's behavior.
type MockManager struct {
	FailCreate          bool
	CreateProcessConfig MockProcess
	FailRegister        bool
	FailList            bool
	FailGroup           bool
	FailGet             bool
	FailClose           bool
	Procs               []Process
}

func mockFail() error {
	progCounter := make([]uintptr, 2)
	n := runtime.Callers(2, progCounter)
	frames := runtime.CallersFrames(progCounter[:n])
	frame, _ := frames.Next()
	return errors.Errorf("function failed: %s", frame.Function)
}

func (m *MockManager) CreateProcess(ctx context.Context, opts *CreateOptions) (Process, error) {
	if m.FailCreate {
		return nil, mockFail()
	}

	proc := MockProcess(m.CreateProcessConfig)
	proc.ProcInfo.Options = *opts
	grip.Infof("mock proc = %+v", proc)

	m.Procs = append(m.Procs, &proc)

	return &proc, nil
}

func (m *MockManager) CreateCommand(ctx context.Context) *Command {
	return NewCommand().ProcConstructor(m.CreateProcess)
}

func (m *MockManager) Register(ctx context.Context, proc Process) error {
	if m.FailRegister {
		return mockFail()
	}

	m.Procs = append(m.Procs, proc)

	return nil
}

func (m *MockManager) List(ctx context.Context, f Filter) ([]Process, error) {
	if m.FailList {
		return nil, mockFail()
	}

	filteredProcs := []Process{}

	for _, proc := range m.Procs {
		info := proc.Info(ctx)
		switch f {
		case All:
			filteredProcs = append(filteredProcs, proc)
		case Running:
			if info.IsRunning {
				filteredProcs = append(filteredProcs, proc)
			}
		case Terminated:
			if !info.IsRunning {
				filteredProcs = append(filteredProcs, proc)
			}
		case Failed:
			if info.Complete && !info.Successful {
				filteredProcs = append(filteredProcs, proc)
			}
		case Successful:
			if info.Successful {
				filteredProcs = append(filteredProcs, proc)
			}
		default:
			return nil, errors.Errorf("invalid filter '%s'", f)
		}
	}

	return filteredProcs, nil
}

func (m *MockManager) Group(ctx context.Context, tag string) ([]Process, error) {
	if m.FailGroup {
		return nil, mockFail()
	}

	matchingProcs := []Process{}
	for _, proc := range m.Procs {
		for _, procTag := range proc.GetTags() {
			if procTag == tag {
				matchingProcs = append(matchingProcs, proc)
			}
		}
	}

	return matchingProcs, nil
}

func (m *MockManager) Get(ctx context.Context, id string) (Process, error) {
	if m.FailGet {
		return nil, mockFail()
	}

	for _, proc := range m.Procs {
		if proc.ID() == id {
			return proc, nil
		}
	}

	return nil, errors.Errorf("proc with id '%s' not found", id)
}

func (m *MockManager) Clear(ctx context.Context) {
	m.Procs = []Process{}
}

func (m *MockManager) Close(ctx context.Context) error {
	if m.FailClose {
		return mockFail()
	}
	m.Clear(ctx)
	return nil
}
