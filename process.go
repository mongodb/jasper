package jasper

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mongodb/grip"
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

	err = defaultOOMScoreAdj(ctx, proc)
	if err != nil {
		return nil, errors.Wrap(err, "setting default oom_score_adj")
	}

	return &synchronizedProcess{proc: proc}, nil
}

// adjustOOMScoreAdj sets the oom_score_adj for the given process to the given score.
func adjustOOMScoreAdj(ctx context.Context, proc Process, score int) error {
	if proc == nil {
		return errors.New("cannot adjust oom_score_adj for nil process")
	}
	if score < -1000 || score > 1000 {
		return errors.Errorf("oom_score_adj must be between -1000 and 1000, but got %d", score)
	}

	procInfo := proc.Info(ctx)
	if procInfo.PID == 0 {
		return errors.New("cannot adjust oom_score_adj for process with empty info")
	}
	PID := strconv.Itoa(procInfo.PID)

	oomScoreAdjPath := filepath.Join("/proc", PID, "oom_score_adj")
	// WriteFile will create a new file with read only permissions if the file doesn't exist,
	// however, the oom_score_adj file should always exist so it is just a failsafe.
	if err := os.WriteFile(oomScoreAdjPath, []byte(strconv.Itoa(score)), 0444); err != nil {
		return errors.Wrapf(err, "writing to '%s'", oomScoreAdjPath)
	}
	grip.Infof("bynnbynn adjusted oom_score_adj for process '%s' to %d", PID, score)
	return nil
}

// defaultOOMScoreAdj sets the oom_score_adj for the given process to 0, which
// is the default value.
func defaultOOMScoreAdj(ctx context.Context, proc Process) error {
	return adjustOOMScoreAdj(ctx, proc, 0)
}
