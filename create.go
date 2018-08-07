package jasper

import (
	"github.com/google/shlex"
	"github.com/pkg/errors"
)

type CreateOptions struct {
	Args             []string
	Environment      map[string]string
	WorkingDirectory string
	Output           OutputOptions
}

func MakeCreationOptions(cmdStr string) CreateOptions {
	args, err := shlex.Split(cmdStr)
	if err != nil {
		return errors.Wrap(err, "problem parsing shell command")
	}

	if len(args) == 0 {
		return errors.Errorf("'%s' did not parse to valid args array", cmdStr)
	}

	return CreateOptions{
		Args: args,
	}
}
