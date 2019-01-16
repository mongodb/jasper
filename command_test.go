package jasper

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/level"
	"github.com/stretchr/testify/assert"
)

type BufferCloser struct {
	*bytes.Buffer
}

func (b *BufferCloser) Close() error { return nil }

func TestCommandImplementation(t *testing.T) {
	for name, testCase := range map[string]func(context.Context, *testing.T){
		"ValidRunCommandDoesNotError": func(ctx context.Context, t *testing.T) {
			assert.NoError(
				t,
				RunCommand(
					context.Background(),
					"test",
					level.Info,
					[]string{"echo", "hello world"},
					"/Users/may/quick/",
					map[string]string{},
				),
			)
		},
		"CommandOutput": func(ctx context.Context, t *testing.T) {
			echo, ls := "echo", "ls"
			arg1, arg2, arg3 := "lalala", "second", "DNE"

			verifyOutput := func(cmd *Command, t *testing.T, expectedOutputs ...string) {
				var buf bytes.Buffer
				bufCloser := &BufferCloser{&buf}

				cmd.SetOutputWriter(bufCloser)
				assert.NoError(t, cmd.Run(context.Background()))
				output := bufCloser.String()
				grip.Debugf("Got output: %v", output)

				for _, expected := range expectedOutputs {
					assert.True(t, strings.Contains(output, expected))
				}
			}

			for subName, subTestCase := range map[string]func(context.Context, *testing.T){
				"StdOutOnly": func(ctx context.Context, t *testing.T) {
					cmd := NewCommand()
					cmd.Add([]string{echo, arg1})
					cmd.Add([]string{echo, arg2})
					verifyOutput(cmd, t, arg1, arg2)
				},
				"StdErrOnly": func(ctx context.Context, t *testing.T) {
					cmd := NewCommand()
					cmd.Add([]string{ls, arg3})
					lsOutput := fmt.Sprintf("%s: %s: No such file or directory", ls, arg3)
					verifyOutput(cmd, t, lsOutput)
				},
				"StdOutAndStdErr": func(ctx context.Context, t *testing.T) {
					cmd := NewCommand()
					cmd.Add([]string{echo, arg1})
					cmd.Add([]string{echo, arg2})
					cmd.Add([]string{ls, arg3})
					lsOutput := fmt.Sprintf("%s: %s: No such file or directory", ls, arg3)
					verifyOutput(cmd, t, arg1, arg2, lsOutput)
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
