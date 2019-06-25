package jasper

import (
	"context"
	"syscall"

	"github.com/mongodb/grip"
)

// MockProcess implements the Process interface with exported fields to
// configure and introspect the mock's behavior.
type MockProcess struct {
	ProcInfo ProcessInfo

	FailRespawn bool

	FailRegisterTrigger bool
	Triggers            ProcessTriggerSequence

	FailRegisterSignalTrigger bool
	SignalTriggers            SignalTriggerSequence
	SignalTriggerIDs          []SignalTriggerID

	FailSignal bool
	Signals    []syscall.Signal

	Tags []string

	FailWait     bool
	WaitExitCode int
}

func (p *MockProcess) ID() string {
	return p.ProcInfo.ID
}

func (p *MockProcess) Info(ctx context.Context) ProcessInfo {
	return p.ProcInfo
}

func (p *MockProcess) Running(ctx context.Context) bool {
	return p.ProcInfo.IsRunning
}

func (p *MockProcess) Complete(ctx context.Context) bool {
	return p.ProcInfo.Complete
}

func (p *MockProcess) GetTags() []string {
	return p.Tags
}

func (p *MockProcess) Tag(tag string) {
	p.Tags = append(p.Tags, tag)
}

func (p *MockProcess) ResetTags() {
	p.Tags = []string{}
}

func (p *MockProcess) Signal(ctx context.Context, sig syscall.Signal) error {
	if p.FailSignal {
		return mockFail()
	}

	p.Signals = append(p.Signals, sig)

	return nil
}

func (p *MockProcess) Wait(_ context.Context) (int, error) {
	if p.FailWait {
		return -1, mockFail()
	}

	return p.WaitExitCode, nil
}

func (p *MockProcess) Respawn(ctx context.Context) (Process, error) {
	if p.FailRespawn {
		return nil, mockFail()
	}

	newProc := *p

	return &newProc, nil
}

func (p *MockProcess) RegisterTrigger(ctx context.Context, t ProcessTrigger) error {
	grip.Infof("kim: FailRegisterTrigger = %t", p.FailRegisterTrigger)
	if p.FailRegisterTrigger {
		grip.Infof("FailRegisterTrigger")
		return mockFail()
	}

	p.Triggers = append(p.Triggers, t)

	return nil
}

func (p *MockProcess) RegisterSignalTrigger(ctx context.Context, t SignalTrigger) error {
	if p.FailRegisterSignalTrigger {
		return mockFail()
	}

	p.SignalTriggers = append(p.SignalTriggers, t)

	return nil
}

func (p *MockProcess) RegisterSignalTriggerID(ctx context.Context, sigID SignalTriggerID) error {
	if p.FailRegisterSignalTrigger {
		return mockFail()
	}

	p.SignalTriggerIDs = append(p.SignalTriggerIDs, sigID)

	return nil
}
