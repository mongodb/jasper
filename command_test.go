package jasper

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/send"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	echo, ls         = "echo", "ls"
	arg1, arg2, arg3 = "ZXZlcmdyZWVu", "aXM=", "c28gY29vbCE="
)

type BufferCloser struct {
	*bytes.Buffer
}

func (b *BufferCloser) Close() error { return nil }

func lsErrorMsg(badDir string) string {
	return fmt.Sprintf("%s: %s: No such file or directory", ls, badDir)
}

func verifyCommandAndGetOutput(t *testing.T, cmd *Command, expectSuccess bool) string {
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

func checkOutput(t *testing.T, cmd *Command, exists bool, output string, expectedOutputs ...string) {
	for _, expected := range expectedOutputs {
		// TODO: Maybe don't try to be so cheeky with an XOR...
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
			cmd := NewCommand().Add([]string{})
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
					cmd := NewCommand().SetPrecondition(precondition).Add([]string{echo, arg1})
					output := verifyCommandAndGetOutput(t, cmd, true)
					checkOutput(t, cmd, precondition(), output, arg1)
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

				cmd := NewCommand().Add([]string{echo, arg1})
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
					output := verifyCommandAndGetOutput(t, cmd, successful)
					checkOutput(t, cmd, true, output, arg1)
					checkOutput(t, cmd, includeBadCmd, output, lsErrorMsg(arg3))
					checkOutput(t, cmd, outputAfterLsExists, output, arg2)
				})
			}
		},
		"CommandOutputAndErrorIsReadable": func(ctx context.Context, t *testing.T) {
			for subName, subTestCase := range map[string]func(context.Context, *testing.T, *Command){
				"StdOutOnly": func(ctx context.Context, t *testing.T, cmd *Command) {
					cmd.Add([]string{echo, arg1})
					cmd.Add([]string{echo, arg2})
					output := verifyCommandAndGetOutput(t, cmd, true)
					checkOutput(t, cmd, true, output, arg1, arg2)
				},
				"StdErrOnly": func(ctx context.Context, t *testing.T, cmd *Command) {
					cmd.Add([]string{ls, arg3})
					output := verifyCommandAndGetOutput(t, cmd, false)
					checkOutput(t, cmd, true, output, lsErrorMsg(arg3))
				},
				"StdOutAndStdErr": func(ctx context.Context, t *testing.T, cmd *Command) {
					cmd.Add([]string{echo, arg1})
					cmd.Add([]string{echo, arg2})
					cmd.Add([]string{ls, arg3})
					output := verifyCommandAndGetOutput(t, cmd, false)
					checkOutput(t, cmd, true, output, arg1, arg2, lsErrorMsg(arg3))
				},
			} {
				t.Run(subName, func(t *testing.T) {
					cmd := NewCommand()
					subTestCase(ctx, t, cmd)
				})
			}
		},
		"WriterOutputAndErrorIsSettable": func(ctx context.Context, t *testing.T) {
			for subName, subTestCase := range map[string]func(context.Context, *testing.T, *Command, *BufferCloser){
				"StdOutOnly": func(ctx context.Context, t *testing.T, cmd *Command, buf *BufferCloser) {
					cmd.SetOutputWriter(buf)
					require.NoError(t, cmd.Run(context.Background()))
					grip.Debugf("Buf string: %v", buf.String())
					grip.Debugf("Buf string: %v", buf.String())
					checkOutput(t, cmd, true, buf.String(), arg1, arg2)
					checkOutput(t, cmd, false, buf.String(), lsErrorMsg(arg3))
				},
				"StdErrOnly": func(ctx context.Context, t *testing.T, cmd *Command, buf *BufferCloser) {
					cmd.SetErrorWriter(buf)
					require.NoError(t, cmd.Run(context.Background()))
					checkOutput(t, cmd, true, buf.String(), lsErrorMsg(arg3))
					checkOutput(t, cmd, false, buf.String(), arg1, arg2)
				},
				"StdOutAndStdErr": func(ctx context.Context, t *testing.T, cmd *Command, buf *BufferCloser) {
					cmd.SetCombinedWriter(buf)
					require.NoError(t, cmd.Run(context.Background()))
					checkOutput(t, cmd, true, buf.String(), arg1, arg2, lsErrorMsg(arg3))
				},
			} {
				t.Run(subName, func(t *testing.T) {
					cmd := NewCommand().Extend([][]string{
						[]string{echo, arg1},
						[]string{echo, arg2},
						[]string{ls, arg3},
					}).SetContinueOnError(true).SetIgnoreError(true)

					var buf bytes.Buffer
					bufCloser := &BufferCloser{&buf}

					subTestCase(ctx, t, cmd, bufCloser)
				})
			}
		},
		"SenderOutputAndErrorIsSettable": func(ctx context.Context, t *testing.T) {
			for subName, subTestCase := range map[string]func(context.Context, *testing.T, *Command, *send.InMemorySender){
				"StdOutOnly": func(ctx context.Context, t *testing.T, cmd *Command, sender *send.InMemorySender) {
					cmd.SetOutputSender(cmd.priority, sender)
					require.NoError(t, cmd.Run(context.Background()))
					out, err := sender.GetString()
					require.NoError(t, err)
					checkOutput(t, cmd, true, strings.Join(out, "\n"), "[p=info]:", arg1, arg2)
					checkOutput(t, cmd, false, strings.Join(out, "\n"), lsErrorMsg(arg3))
				},
				"StdErrOnly": func(ctx context.Context, t *testing.T, cmd *Command, sender *send.InMemorySender) {
					cmd.SetErrorSender(cmd.priority, sender)
					require.NoError(t, cmd.Run(context.Background()))
					out, err := sender.GetString()
					require.NoError(t, err)
					checkOutput(t, cmd, true, strings.Join(out, "\n"), "[p=info]:", lsErrorMsg(arg3))
					checkOutput(t, cmd, false, strings.Join(out, "\n"), arg1, arg2)
				},
				"StdOutAndStdErr": func(ctx context.Context, t *testing.T, cmd *Command, sender *send.InMemorySender) {
					cmd.SetCombinedSender(cmd.priority, sender)
					require.NoError(t, cmd.Run(context.Background()))
					out, err := sender.GetString()
					require.NoError(t, err)
					checkOutput(t, cmd, true, strings.Join(out, "\n"), "[p=info]:", arg1, arg2, lsErrorMsg(arg3))
				},
			} {
				t.Run(subName, func(t *testing.T) {
					cmd := NewCommand().Extend([][]string{
						[]string{echo, arg1},
						[]string{echo, arg2},
						[]string{ls, arg3},
					}).SetContinueOnError(true).SetIgnoreError(true).Priority(level.Info)

					levelInfo := send.LevelInfo{Default: cmd.priority, Threshold: cmd.priority}
					sender, err := send.NewInMemorySender(t.Name(), levelInfo, 100)
					require.NoError(t, err)

					subTestCase(ctx, t, cmd, sender.(*send.InMemorySender))
				})
			}
		},
		"RunParallelRunsInParallel": func(ctx context.Context, t *testing.T) {
			cmd := NewCommand().Extend([][]string{
				[]string{"sleep", "3"},
				[]string{"sleep", "3"},
				[]string{"sleep", "3"},
			})
			threePointOneSeconds := time.Second*3 + time.Millisecond*100
			maxRunTimeAllowed := threePointOneSeconds
			cctx, cancel := context.WithTimeout(ctx, maxRunTimeAllowed)
			defer cancel()
			// If this does not run in parallel, the context will timeout and we will
			// get an error.
			assert.NoError(t, cmd.RunParallel(cctx))
		},
		// "": func(ctx context.Context, t *testing.T) {},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
			defer cancel()

			testCase(ctx, t)
		})
	}
}
