package jasper

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mongodb/grip/level"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	echo, ls         = "echo", "ls"
	arg1, arg2, arg3 = "123", "456", "789"
)

type BufferCloser struct {
	*bytes.Buffer
}

func (b *BufferCloser) Close() error { return nil }

func verifyOutput(cmd *Command, t *testing.T, expectSuccess bool, expectedOutputs ...string) {
	var buf bytes.Buffer
	bufCloser := &BufferCloser{&buf}

	cmd.SetOutputWriter(bufCloser)
	if expectSuccess {
		assert.NoError(t, cmd.Run(context.Background()))
	} else {
		assert.Error(t, cmd.Run(context.Background()))
	}

	output := bufCloser.String()

	for _, expected := range expectedOutputs {
		assert.True(t, strings.Contains(output, expected))
	}
}

func TestCommandImplementation(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	for name, testCase := range map[string]func(context.Context, *testing.T){
		"ValidRunCommandDoesNotError": func(ctx context.Context, t *testing.T) {
			assert.NoError(
				t,
				RunCommand(
					ctx,
					t.Name(),
					level.Info,
					[]string{echo, arg1},
					cwd,
					map[string]string{},
				),
			)
		},
		"InvalidRunCommandErrors": func(ctx context.Context, t *testing.T) {
			assert.Error(
				t,
				RunCommand(
					ctx,
					t.Name(),
					level.Info,
					[]string{ls, arg2},
					cwd,
					map[string]string{},
				),
			)
		},
		"SingleInvalidSubCommandCausesTotalError": func(ctx context.Context, t *testing.T) {
			assert.Error(
				t,
				RunCommandGroup(
					ctx,
					t.Name(),
					level.Info,
					[][]string{
						[]string{echo, arg1},
						[]string{ls, arg2},
						[]string{echo, arg3},
					},
					cwd,
					map[string]string{},
				),
			)
		},
		"IgnoreErrorCausesSuccessfulReturn": func(ctx context.Context, t *testing.T) {
			cmd := NewCommand()
			cmd.Extend([][]string{
				[]string{echo, arg1},
				[]string{ls, arg3},
				[]string{echo, arg2},
			})
			cmd.SetIgnoreError(true)
			verifyOutput(cmd, t, true, arg1, fmt.Sprintf("%s: %s: No such file or directory", ls, arg3), arg2)
		},
		"CommandOutput": func(ctx context.Context, t *testing.T) {

			for subName, subTestCase := range map[string]func(context.Context, *testing.T){
				"StdOutOnly": func(ctx context.Context, t *testing.T) {
					cmd := NewCommand()
					cmd.Add([]string{echo, arg1})
					cmd.Add([]string{echo, arg2})
					verifyOutput(cmd, t, true, arg1, arg2)
				},
				"StdErrOnly": func(ctx context.Context, t *testing.T) {
					cmd := NewCommand()
					cmd.Add([]string{ls, arg3})
					lsOutput := fmt.Sprintf("%s: %s: No such file or directory", ls, arg3)
					verifyOutput(cmd, t, false, lsOutput)
				},
				"StdOutAndStdErr": func(ctx context.Context, t *testing.T) {
					cmd := NewCommand()
					cmd.Add([]string{echo, arg1})
					cmd.Add([]string{echo, arg2})
					cmd.Add([]string{ls, arg3})
					lsOutput := fmt.Sprintf("%s: %s: No such file or directory", ls, arg3)
					verifyOutput(cmd, t, false, arg1, arg2, lsOutput)
				},
			} {
				t.Run(subName, func(t *testing.T) {
					subTestCase(ctx, t)
				})
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
			defer cancel()

			testCase(ctx, t)
		})
	}
}
