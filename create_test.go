package jasper

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateConstructor(t *testing.T) {
	for _, test := range []struct {
		id         string
		shouldFail bool
		cmd        string
		args       []string
	}{
		{
			id:         "EmptyString",
			shouldFail: true,
		},
		{
			id:         "BasicCmd",
			args:       []string{"ls", "-lha"},
			cmd:        "ls -lha",
			shouldFail: false,
		},
		{
			id:         "SkipsCommentsAtBeginning",
			shouldFail: true,
			cmd:        "# wat",
		},
		{
			id:         "SkipsCommentsAtEnd",
			cmd:        "ls #what",
			args:       []string{"ls"},
			shouldFail: false,
		},
		{
			id:         "UnbalancedShellLex",
			cmd:        "' foo",
			shouldFail: true,
		},
	} {
		t.Run(test.id, func(t *testing.T) {
			opt, err := MakeCreationOptions(test.cmd)
			if test.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, opt)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, opt)
			assert.Equal(t, test.args, opt.Args)
		})
	}
}

func TestCreateOptions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for name, test := range map[string]func(t *testing.T, opts *CreateOptions){
		"DefaultConfigForTestsValidate": func(t *testing.T, opts *CreateOptions) {
			assert.NoError(t, opts.Validate())
		},
		"EmptyArgsShouldNotValidate": func(t *testing.T, opts *CreateOptions) {
			opts.Args = []string{}
			assert.Error(t, opts.Validate())
		},
		"ZeroTimeoutShouldNotError": func(t *testing.T, opts *CreateOptions) {
			opts.Timeout = 0
			assert.NoError(t, opts.Validate())
		},
		"SmallTimeoutShouldNotValidate": func(t *testing.T, opts *CreateOptions) {
			opts.Timeout = time.Millisecond
			assert.Error(t, opts.Validate())
		},
		"LargeTimeoutShouldValidate": func(t *testing.T, opts *CreateOptions) {
			opts.Timeout = time.Hour
			assert.NoError(t, opts.Validate())
		},
		"NonExistingWorkingDirectoryShouldNotValidate": func(t *testing.T, opts *CreateOptions) {
			opts.WorkingDirectory = "foo"
			assert.Error(t, opts.Validate())
		},
		"ExtantWorkingDirectoryShouldPass": func(t *testing.T, opts *CreateOptions) {
			wd, err := os.Getwd()
			assert.NoError(t, err)
			assert.NotZero(t, wd)

			opts.WorkingDirectory = wd
			assert.NoError(t, opts.Validate())
		},
		"WorkingDirectoryShouldErrorForFiles": func(t *testing.T, opts *CreateOptions) {
			gobin, err := exec.LookPath("go")
			assert.NoError(t, err)
			assert.NotZero(t, gobin)

			opts.WorkingDirectory = gobin
			assert.Error(t, opts.Validate())
		},
		"MustSpecifyValidOutputOptions": func(t *testing.T, opts *CreateOptions) {
			opts.Output.SendErrorToOutput = true
			opts.Output.SendOutputToError = true
			assert.Error(t, opts.Validate())
		},
		"WorkingDirectoryUnresolveableShouldNotError": func(t *testing.T, opts *CreateOptions) {
			cmd, err := opts.Resolve(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			assert.NotZero(t, cmd.Dir)
			assert.Equal(t, opts.WorkingDirectory, cmd.Dir)
		},
		"ResolveFailsIfOptionsAreFatal": func(t *testing.T, opts *CreateOptions) {
			opts.Args = []string{}
			cmd, err := opts.Resolve(ctx)
			assert.Error(t, err)
			assert.Nil(t, cmd)
		},
		"WithoutOverrideEnvironmentEnvIsPopulated": func(t *testing.T, opts *CreateOptions) {
			cmd, err := opts.Resolve(ctx)
			assert.NoError(t, err)
			assert.NotZero(t, cmd.Env)
		},
		"WithOverrideEnvironmentEnvIsEmpty": func(t *testing.T, opts *CreateOptions) {
			opts.OverrideEnviron = true
			cmd, err := opts.Resolve(ctx)
			assert.NoError(t, err)
			assert.Zero(t, cmd.Env)
		},
		"EnvironmentVariablesArePropogated": func(t *testing.T, opts *CreateOptions) {
			opts.Environment = map[string]string{
				"foo": "bar",
			}

			cmd, err := opts.Resolve(ctx)
			assert.NoError(t, err)
			assert.Contains(t, cmd.Env, "foo=bar")
			assert.NotContains(t, cmd.Env, "bar=foo")
		},
		"MultipleArgsArePropogated": func(t *testing.T, opts *CreateOptions) {
			opts.Args = append(opts.Args, "-lha")
			cmd, err := opts.Resolve(ctx)
			assert.NoError(t, err)
			assert.Contains(t, cmd.Path, "ls")
			assert.Len(t, cmd.Args, 2)
		},
		"WithOnlyCommandsArgsHasOneVal": func(t *testing.T, opts *CreateOptions) {
			cmd, err := opts.Resolve(ctx)
			assert.NoError(t, err)
			assert.Contains(t, cmd.Path, "ls")
			assert.Len(t, cmd.Args, 1)
			assert.Equal(t, "ls", cmd.Args[0])
		},
		"WithTimeout": func(t *testing.T, opts *CreateOptions) {
			opts.Timeout = time.Second
			opts.Args = []string{"sleep", "2"}

			cmd, err := opts.Resolve(ctx)
			assert.NoError(t, err)
			assert.Error(t, cmd.Run())
		},
		"ClosersAreAlwaysCalled": func(t *testing.T, opts *CreateOptions) {
			var counter int
			opts.closers = append(opts.closers,
				func() { counter++ },
				func() { counter += 2 })
			opts.Close()
			assert.Equal(t, counter, 3)

		},
	} {
		t.Run(name, func(t *testing.T) {
			opts := &CreateOptions{Args: []string{"ls"}}
			test(t, opts)
		})
	}
}
