package jasper

import (
	"context"

	"github.com/mongodb/jasper/options"
)

func makeLockingProcess(pmake ProcessConstructor) ProcessConstructor {
	return func(ctx context.Context, opts *options.Create) (Process, error) {
		proc, err := pmake(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &synchronizedProcess{proc: proc}, nil
	}
}
