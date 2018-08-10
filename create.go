package jasper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/google/shlex"
	"github.com/pkg/errors"
)

type CreateOptions struct {
	Args             []string
	Environment      map[string]string
	WorkingDirectory string
	Output           OutputOptions
	OverrideEnviron  bool
	Timeout          time.Duration
	Tags             []string

	OnSuccess []*CreateOptions
	OnFailure []*CreateOptions
	OnTimeout []*CreateOptions

	//
	closers []func()
	started bool
}

func MakeCreationOptions(cmdStr string) (*CreateOptions, error) {
	args, err := shlex.Split(cmdStr)
	if err != nil {
		return nil, errors.Wrap(err, "problem parsing shell command")
	}

	if len(args) == 0 {
		return nil, errors.Errorf("'%s' did not parse to valid args array", cmdStr)
	}

	return &CreateOptions{
		Args: args,
	}, nil
}

func (opts *CreateOptions) Validate() error {
	if len(opts.Args) == 0 {
		return errors.New("invalid command, must specify at least one argument")
	}

	if opts.Timeout > 0 && opts.Timeout < time.Second {
		return errors.New("when specifying a timeout you must use out greater than one second")
	}

	if err := opts.Output.Validate(); err != nil {
		return errors.Wrap(err, "cannot create command with invalid output")
	}

	if opts.WorkingDirectory != "" {
		info, err := os.Stat(opts.WorkingDirectory)

		if os.IsNotExist(err) {
			return errors.Errorf("could not use non-extant %s as working directory", opts.WorkingDirectory)
		}

		if !info.IsDir() {
			return errors.Errorf("could not use file as working directory")
		}
	}

	return nil
}

func (opts *CreateOptions) Resolve(ctx context.Context) (*exec.Cmd, error) {
	var err error
	if ctx.Err() != nil {
		return nil, errors.New("cannot resolve command with canceled context")
	}

	if err = opts.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}

	if opts.WorkingDirectory == "" {
		opts.WorkingDirectory, _ = os.Getwd()
	}

	var env []string
	if !opts.OverrideEnviron {
		env = os.Environ()
	}

	for k, v := range opts.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	var args []string
	if len(opts.Args) > 1 {
		args = opts.Args[1:]
	}

	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		opts.closers = append(opts.closers, cancel)
	}

	cmd := exec.CommandContext(ctx, opts.Args[0], args...) // nolint
	cmd.Dir = opts.WorkingDirectory
	cmd.Stderr = opts.Output.GetError()
	cmd.Stdout = opts.Output.GetOutput()
	cmd.Env = env

	return cmd, nil
}

func (opts *CreateOptions) Close() {
	for _, c := range opts.closers {
		c()
	}
}
