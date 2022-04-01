package jasper

import (
	"context"

	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

// NewProcess is a factory function which constructs a thread-safe standalone
// process outside of the context of a manager.
func NewProcess(ctx context.Context, opts *options.Create) (Process, error) {
	var (
		proc Process
		err  error
	)

	if err = opts.Validate(); err != nil {
		return nil, errors.WithStack((err))
	}

	switch opts.Implementation {
	case options.ProcessImplementationBlocking:
		proc, err = newBlockingProcess(ctx, opts)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case options.ProcessImplementationBasic:
		proc, err = newBasicProcess(ctx, opts)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	default:
		return nil, errors.Errorf("cannot create unrecognized process type '%s'", opts.Implementation)
	}

	if !opts.Synchronized {
		return proc, nil
	}
	return &synchronizedProcess{proc: proc}, nil
}
