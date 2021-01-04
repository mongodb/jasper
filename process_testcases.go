package jasper

import (
	"context"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ProcessTestCase func(context.Context, *testing.T, *options.Create, ProcessConstructor)

func ProcessTests() map[string]ProcessTestCase {
	return map[string]ProcessTestCase{
		"WithPopulatedArgsCommandCreationPasses": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			assert.NotZero(t, opts.Args)
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			assert.NotNil(t, proc)
		},
		"ErrorToCreateWithInvalidArgs": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Args = []string{}
			proc, err := makeProc(ctx, opts)
			assert.Error(t, err)
			assert.Nil(t, proc)
		},
		"WithCanceledContextProcessCreationFails": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			pctx, pcancel := context.WithCancel(ctx)
			pcancel()
			proc, err := makeProc(pctx, opts)
			assert.Error(t, err)
			assert.Nil(t, proc)
		},
		"ProcessLacksTagsByDefault": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			tags := proc.GetTags()
			assert.Empty(t, tags)
		},
		"ProcessTagsPersist": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Tags = []string{"foo"}
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			tags := proc.GetTags()
			assert.Contains(t, tags, "foo")
		},
		"InfoTagsMatchGetTags": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Tags = []string{"foo"}
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			tags := proc.GetTags()
			assert.Contains(t, tags, "foo")
			assert.Equal(t, tags, proc.Info(ctx).Options.Tags)

			proc.ResetTags()
			tags = proc.GetTags()
			assert.Empty(t, tags)
			assert.Empty(t, proc.Info(ctx).Options.Tags)
		},
		"InfoHasMatchingID": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			_, err = proc.Wait(ctx)
			require.NoError(t, err)
			assert.Equal(t, proc.ID(), proc.Info(ctx).ID)
		},
		"ResetTags": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			proc.Tag("foo")
			assert.Contains(t, proc.GetTags(), "foo")
			proc.ResetTags()
			assert.Len(t, proc.GetTags(), 0)
		},
		"TagsHaveSetSemantics": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)

			for i := 0; i < 10; i++ {
				proc.Tag("foo")
			}

			assert.Len(t, proc.GetTags(), 1)
			proc.Tag("bar")
			assert.Len(t, proc.GetTags(), 2)
		},
		"CompleteIsTrueAfterWait": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			_, err = proc.Wait(ctx)
			assert.NoError(t, err)
			assert.True(t, proc.Complete(ctx))
		},
		"WaitReturnsWithCanceledContext": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Args = testutil.SleepCreateOpts(20).Args
			pctx, pcancel := context.WithCancel(ctx)
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			assert.True(t, proc.Running(ctx))
			pcancel()
			_, err = proc.Wait(pctx)
			assert.Error(t, err)
		},
		"RegisterTriggerErrorsWithNilTrigger": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			assert.Error(t, proc.RegisterTrigger(ctx, nil))
		},
		"RegisterSignalTriggerErrorsWithNilTrigger": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			assert.Error(t, proc.RegisterSignalTrigger(ctx, nil))
		},
		"RegisterSignalTriggerErrorsWithExitedProcess": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			_, err = proc.Wait(ctx)
			assert.NoError(t, err)
			assert.Error(t, proc.RegisterSignalTrigger(ctx, func(ProcessInfo, syscall.Signal) bool { return false }))
		},
		"RegisterSignalTriggerIDErrorsWithExitedProcess": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			_, err = proc.Wait(ctx)
			assert.NoError(t, err)
			assert.Error(t, proc.RegisterSignalTriggerID(ctx, CleanTerminationSignalTrigger))
		},
		"RegisterSignalTriggerIDFailsWithInvalidTriggerID": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Args = testutil.SleepCreateOpts(3).Args
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			assert.Error(t, proc.RegisterSignalTriggerID(ctx, SignalTriggerID("foo")))
		},
		"RegisterSignalTriggerIDPassesWithValidTriggerID": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Args = testutil.SleepCreateOpts(3).Args
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			assert.NoError(t, proc.RegisterSignalTriggerID(ctx, CleanTerminationSignalTrigger))
		},
		"WaitOnRespawnedProcessDoesNotError": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			require.NotNil(t, proc)
			_, err = proc.Wait(ctx)
			require.NoError(t, err)

			newProc, err := proc.Respawn(ctx)
			require.NoError(t, err)
			_, err = newProc.Wait(ctx)
			assert.NoError(t, err)
		},
		"RespawnedProcessGivesSameResult": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			require.NotNil(t, proc)

			_, err = proc.Wait(ctx)
			require.NoError(t, err)
			exitCode := proc.Info(ctx).ExitCode

			newProc, err := proc.Respawn(ctx)
			require.NoError(t, err)
			_, err = newProc.Wait(ctx)
			require.NoError(t, err)
			assert.Equal(t, exitCode, proc.Info(ctx).ExitCode)
		},
		"RespawningCompletedProcessIsOK": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			require.NotNil(t, proc)
			_, err = proc.Wait(ctx)
			require.NoError(t, err)

			newProc, err := proc.Respawn(ctx)
			require.NoError(t, err)
			require.NotNil(t, newProc)
			_, err = newProc.Wait(ctx)
			require.NoError(t, err)
			assert.True(t, newProc.Info(ctx).Successful)
		},
		"RespawningRunningProcessIsOK": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Args = testutil.SleepCreateOpts(2).Args
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			require.NotNil(t, proc)

			newProc, err := proc.Respawn(ctx)
			require.NoError(t, err)
			require.NotNil(t, newProc)
			_, err = newProc.Wait(ctx)
			require.NoError(t, err)
			assert.True(t, newProc.Info(ctx).Successful)
		},
		"RespawnShowsConsistentStateValues": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Args = testutil.SleepCreateOpts(2).Args
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			require.NotNil(t, proc)
			_, err = proc.Wait(ctx)
			require.NoError(t, err)

			newProc, err := proc.Respawn(ctx)
			require.NoError(t, err)
			assert.True(t, newProc.Running(ctx))
			_, err = newProc.Wait(ctx)
			require.NoError(t, err)
			assert.True(t, newProc.Complete(ctx))
		},
		"WaitGivesSuccessfulExitCode": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Args = testutil.TrueCreateOpts().Args
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			require.NotNil(t, proc)
			exitCode, err := proc.Wait(ctx)
			assert.NoError(t, err)
			assert.Equal(t, 0, exitCode)
		},
		"WaitGivesFailureExitCode": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Args = testutil.FalseCreateOpts().Args
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)
			require.NotNil(t, proc)
			exitCode, err := proc.Wait(ctx)
			assert.Error(t, err)
			assert.Equal(t, 1, exitCode)
		},
		"WaitGivesProperExitCodeOnSignalTerminate": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, testutil.SleepCreateOpts(100))
			require.NoError(t, err)
			require.NotNil(t, proc)
			sig := syscall.SIGTERM
			assert.NoError(t, proc.Signal(ctx, sig))
			exitCode, err := proc.Wait(ctx)
			assert.Error(t, err)
			if runtime.GOOS == "windows" {
				assert.Equal(t, 1, exitCode)
			} else {
				assert.Equal(t, int(sig), exitCode)
			}
		},
		"WaitGivesProperExitCodeOnSignalInterrupt": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, testutil.SleepCreateOpts(100))
			require.NoError(t, err)
			require.NotNil(t, proc)
			sig := syscall.SIGINT
			assert.NoError(t, proc.Signal(ctx, sig))
			exitCode, err := proc.Wait(ctx)
			assert.Error(t, err)
			if runtime.GOOS == "windows" {
				assert.Equal(t, 1, exitCode)
			} else {
				assert.Equal(t, int(sig), exitCode)
			}
		},
		"WaitGivesNegativeOneOnAlternativeError": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, testutil.SleepCreateOpts(100))
			require.NoError(t, err)
			require.NotNil(t, proc)

			var exitCode int
			waitFinished := make(chan bool)
			cctx, cancel := context.WithCancel(ctx)
			cancel()
			go func() {
				exitCode, err = proc.Wait(cctx)
				select {
				case waitFinished <- true:
				case <-ctx.Done():
				}
			}()
			select {
			case <-waitFinished:
				assert.Error(t, err)
				assert.Equal(t, -1, exitCode)
			case <-ctx.Done():
				assert.Fail(t, "call to Wait() took too long to finish")
			}
		},
		"InfoHasTimeoutWhenProcessTimesOut": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			opts.Args = testutil.SleepCreateOpts(100).Args
			opts.Timeout = time.Second
			opts.TimeoutSecs = 1
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)

			exitCode, err := proc.Wait(ctx)
			assert.Error(t, err)
			if runtime.GOOS == "windows" {
				assert.Equal(t, 1, exitCode)
			} else {
				assert.Equal(t, int(syscall.SIGKILL), exitCode)
			}
			assert.True(t, proc.Info(ctx).Timeout)
		},
		"SignalingCompletedProcessErrors": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)

			_, err = proc.Wait(ctx)
			assert.NoError(t, err)

			err = proc.Signal(ctx, syscall.SIGTERM)
			require.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), "cannot signal a process that has terminated"))
		},
		"CompleteIsTrueWhenProcessExits": func(ctx context.Context, t *testing.T, opts *options.Create, makeProc ProcessConstructor) {
			proc, err := makeProc(ctx, opts)
			require.NoError(t, err)

			_, err = proc.Wait(ctx)
			assert.NoError(t, err)

			assert.False(t, proc.Running(ctx))
			assert.True(t, proc.Complete(ctx))
		},
	}
}
