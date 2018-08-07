package jasper

import (
	"context"
	"syscall"

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
	if err := errors.WithStack(Teriminate(ctx, p)); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(p.Wait(ctx))
}
