package jasper

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/shlex"
	"github.com/pkg/errors"
)

type CreateOptions struct {
	Args             []string
	Environment      map[string]string
	WorkingDirectory string
	Output           OutputOptions
	OverrideEnviron  bool
}

func MakeCreationOptions(cmdStr string) *CreateOptions {
	args, err := shlex.Split(cmdStr)
	if err != nil {
		return errors.Wrap(err, "problem parsing shell command")
	}

	if len(args) == 0 {
		return errors.Errorf("'%s' did not parse to valid args array", cmdStr)
	}

	return &CreateOptions{
		Args: args,
	}
}

func (opts *CreateOptions) Validate() error {
	if len(opts.Args) == 0 {
		return errors.New("invalid command, must specify at least one argument")
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

func (opts *CreateOptions) Resolve(ctx context.Context) (*exec.Command, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}

	if opts.WorkingDirectory == "" {
		opts.WorkingDirectory, err = os.Getwd()
		if err != nil {
			return nil, errors.WithStack(err)
		}
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

	cmd := exec.CommandContext(ctx, opts.Args[0], args...) // nolint

	cmd.Dir = opts.WorkingDirectory
	cmd.Stderr = opts.Output.GetError()
	cmd.Stdout = opts.Output.GetOutput()
	cmd.Env = env

	return cmd, nil
}
