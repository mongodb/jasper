package jasper

import (
	"context"
	"fmt"
	"time"

	"github.com/mongodb/grip"
)

const (
	taskTimeout        = 5 * time.Second
	processTestTimeout = 15 * time.Second
	managerTestTimeout = 5 * taskTimeout
	longTaskTimeout    = 100 * time.Second
)

func makeLockingProcess(pmake ProcessConstructor) ProcessConstructor {
	return func(ctx context.Context, opts *CreateOptions) (Process, error) {
		proc, err := pmake(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &localProcess{proc: proc}, nil
	}
}

// this file contains tools and constants used throughout the test
// suite.

func trueCreateOpts() *CreateOptions {
	return &CreateOptions{
		Args: []string{"true"},
	}
}

func falseCreateOpts() *CreateOptions {
	return &CreateOptions{
		Args: []string{"false"},
	}
}

func sleepCreateOpts(num int) *CreateOptions {
	return &CreateOptions{
		Args: []string{"sleep", fmt.Sprint(num)},
	}
}

func createProcs(ctx context.Context, opts *CreateOptions, manager Manager, num int) ([]Process, error) {
	catcher := grip.NewBasicCatcher()
	out := []Process{}
	for i := 0; i < num; i++ {
		optsCopy := *opts

		proc, err := manager.CreateProcess(ctx, &optsCopy)
		catcher.Add(err)
		if proc != nil {
			out = append(out, proc)
		}
	}

	return out, catcher.Resolve()
}
