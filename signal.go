package jasper

import (
	"context"
	"syscall"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

func Terminate(ctx context.Context, p Process) error {
	return errors.WithStack(p.Signal(ctx, syscall.SIGTERM))
}

func Kill(ctx context.Context, p Process) error {
	return errors.WithStack(p.Signal(ctx, syscall.SIGKILL))
}

func TerminateAll(ctx context.Context, procs []Process) error {
	catcher := grip.NewBasicCatcher()

	for _, proc := range procs {

		catcher.Add(Terminate(ctx, proc))
	}

	for _, proc := range procs {
		_ = proc.Wait(ctx)
		opts := proc.Info(ctx).Options
		opts.Close()
	}

	return catcher.Resolve()
}

func KillAll(ctx context.Context, procs []Process) error {
	catcher := grip.NewBasicCatcher()

	for _, proc := range procs {
		catcher.Add(Kill(ctx, proc))
	}

	for _, proc := range procs {
		_ = proc.Wait(ctx)
		opts := proc.Info(ctx).Options
		opts.Close()
	}

	return catcher.Resolve()
}
