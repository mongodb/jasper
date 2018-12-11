package jasper

import (
	"context"
	"runtime"
	"syscall"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

func Terminate(ctx context.Context, p Process) error {
	// TODO: MAKE-XXX: Update signal.go functions with Windows-specific behaviors.
	if runtime.GOOS == "windows" {
		return nil
	}
	return errors.WithStack(p.Signal(ctx, syscall.SIGTERM))
}

func Kill(ctx context.Context, p Process) error {
	// TODO: MAKE-XXX: Update signal.go functions with Windows-specific behaviors.
	if runtime.GOOS == "windows" {
		return nil
	}
	return errors.WithStack(p.Signal(ctx, syscall.SIGKILL))
}

func TerminateAll(ctx context.Context, procs []Process) error {
	catcher := grip.NewBasicCatcher()

	for _, proc := range procs {
		if proc.Running(ctx) {
			catcher.Add(Terminate(ctx, proc))
		}
	}

	for _, proc := range procs {
		_ = proc.Wait(ctx)
	}

	return catcher.Resolve()
}

func KillAll(ctx context.Context, procs []Process) error {
	catcher := grip.NewBasicCatcher()

	for _, proc := range procs {
		if proc.Running(ctx) {
			catcher.Add(Kill(ctx, proc))
		}
	}

	for _, proc := range procs {
		_ = proc.Wait(ctx)
	}

	return catcher.Resolve()
}
