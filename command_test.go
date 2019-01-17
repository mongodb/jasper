package jasper

import (
	"bytes"
	"context"
	"fmt"
	"math"
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

func lsErrorMsg(badDir string) string {
	return fmt.Sprintf("%s: %s: No such file or directory", ls, badDir)
}

func verifyCommandAndGetOutput(cmd *Command, t *testing.T, expectSuccess bool) string {
	var buf bytes.Buffer
	bufCloser := &BufferCloser{&buf}

	cmd.SetCombinedWriter(bufCloser)

	if expectSuccess {
		assert.NoError(t, cmd.Run(context.Background()))
	} else {
		assert.Error(t, cmd.Run(context.Background()))
	}

	return bufCloser.String()
}

func checkOutput(cmd *Command, t *testing.T, output string, exists bool, expectedOutputs ...string) {
	for _, expected := range expectedOutputs {
		assert.True(t, exists == strings.Contains(output, expected))
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
		"UnsuccessfulRunCommandErrors": func(ctx context.Context, t *testing.T) {
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
		"InvalidArgsCommandErrors": func(ctx context.Context, t *testing.T) {
			cmd := NewCommand()
			cmd.Add([]string{})
			assert.EqualError(t, cmd.Run(ctx), "args invalid")
		},
		"PreconditionDeterminesExecution": func(ctx context.Context, t *testing.T) {
			for _, precondition := range []func() bool{
				func() bool {
					return true
				},
				func() bool {
					return false
				},
			} {
				t.Run(fmt.Sprintf("%tPrecondition", precondition()), func(t *testing.T) {
					cmd := NewCommand()
					cmd.SetPrecondition(precondition)
					cmd.Add([]string{echo, arg1})
					output := verifyCommandAndGetOutput(cmd, t, true)
					checkOutput(cmd, t, output, precondition(), arg1)
				})
			}
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
		"ExecutionFlags": func(ctx context.Context, t *testing.T) {
			numCombinations := int(math.Pow(2, 3))
			for i := 0; i < numCombinations; i++ {
				continueOnError, ignoreError, includeBadCmd := i&1 == 1, i&2 == 2, i&4 == 4

				cmd := NewCommand()
				cmd.Add([]string{echo, arg1})
				if includeBadCmd {
					cmd.Add([]string{ls, arg3})
				}
				cmd.Add([]string{echo, arg2})

				subTestName := fmt.Sprintf(
					"ContinueOnErrorIs%tAndIgnoreErrorIs%tAndIncludeBadCmdIs%t",
					continueOnError,
					ignoreError,
					includeBadCmd,
				)
				t.Run(subTestName, func(t *testing.T) {
					cmd.SetContinueOnError(continueOnError)
					cmd.SetIgnoreError(ignoreError)
					successful := ignoreError || !includeBadCmd
					outputAfterLsExists := !includeBadCmd || continueOnError
					output := verifyCommandAndGetOutput(cmd, t, successful)
					checkOutput(cmd, t, output, true, arg1)
					checkOutput(cmd, t, output, includeBadCmd, lsErrorMsg(arg3))
					checkOutput(cmd, t, output, outputAfterLsExists, arg2)
				})
			}
		},
		"Output": func(ctx context.Context, t *testing.T) {
			for subName, subTestCase := range map[string]func(context.Context, *testing.T){
				"StdOutOnly": func(ctx context.Context, t *testing.T) {
					cmd := NewCommand()
					cmd.Add([]string{echo, arg1})
					cmd.Add([]string{echo, arg2})
					output := verifyCommandAndGetOutput(cmd, t, true)
					checkOutput(cmd, t, output, true, arg1, arg2)
				},
				"StdErrOnly": func(ctx context.Context, t *testing.T) {
					cmd := NewCommand()
					cmd.Add([]string{ls, arg3})
					output := verifyCommandAndGetOutput(cmd, t, false)
					checkOutput(cmd, t, output, true, lsErrorMsg(arg3))
				},
				"StdOutAndStdErr": func(ctx context.Context, t *testing.T) {
					cmd := NewCommand()
					cmd.Add([]string{echo, arg1})
					cmd.Add([]string{echo, arg2})
					cmd.Add([]string{ls, arg3})
					output := verifyCommandAndGetOutput(cmd, t, false)
					checkOutput(cmd, t, output, true, arg1, arg2, lsErrorMsg(arg3))
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
