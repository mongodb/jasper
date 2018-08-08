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

func KillBlocking(ctx context.Context, p Process) error {
	if err := errors.WithStack(Kill(ctx, p)); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(p.Wait(ctx))
}

func TerminateBlocking(ctx context.Context, p Process) error {
	if err := errors.WithStack(Terminate(ctx, p)); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(p.Wait(ctx))
}

func TerminateAll(ctx context.Context, procs []Process) error {
	catcher := grip.NewBasicCatcher()

	for _, proc := range procs {
		catcher.Add(Terminate(ctx, proc))
	}

	for _, proc := range procs {
		catcher.Add(proc.Wait(ctx))
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
		catcher.Add(proc.Wait(ctx))
		opts := proc.Info(ctx).Options
		opts.Close()
	}

	return catcher.Resolve()
}
