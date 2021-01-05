package jasper

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mongodb/grip"
	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ProcessTestCase represents a testing function including a constructor and
// options to create a process.
type ProcessTestCase func(context.Context, *testing.T, *options.Create, ProcessConstructor)

// ProcessTests returns the common test suite for the Process interface. This
// should be used for testing purposes only.
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

type ManagerConstructor func(context.Context)

// ManagerTestCase represents a testing function including a constructor and
// options to create a manager, as well as a function to modify process creation
// options.
type ManagerTestCase struct {
	Name string
	Case func(context.Context, *testing.T, Manager, testutil.OptsModify)
}

// ManagerTests returns the common test suite for the Manager interface. This
// should be used for testing purposes only.
func ManagerTests() []ManagerTestCase {
	return []ManagerTestCase{
		{
			Name: "ValidateFixture",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				assert.NotNil(t, ctx)
				assert.NotNil(t, mngr)
				// kim: NOTE: this is not present in the remote tests.
				assert.NotNil(t, mngr.LoggingCache(ctx))
			},
		},
		{
			Name: "IDReturnsNonempty",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				assert.NotEmpty(t, mngr.ID())
			},
		},
		{
			Name: "ProcEnvVarMatchesManagerID",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())
				proc, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)
				info := proc.Info(ctx)
				require.NotEmpty(t, info.Options.Environment)
				assert.Equal(t, mngr.ID(), info.Options.Environment[ManagerEnvironID])
			},
		},
		{
			Name: "CreateProcessFailsWithEmptyOptions",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(&options.Create{})
				proc, err := mngr.CreateProcess(ctx, opts)
				require.Error(t, err)
				assert.Nil(t, proc)
			},
		},
		{
			Name: "CreateSimpleProcess",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())
				proc, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)
				assert.NotNil(t, proc)
				info := proc.Info(ctx)
				assert.True(t, info.IsRunning || info.Complete)
			},
		},
		{
			Name: "CreateProcessFails",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(&options.Create{})
				proc, err := mngr.CreateProcess(ctx, opts)
				require.Error(t, err)
				assert.Nil(t, proc)
			},
		},
		{
			Name: "ListWithoutResultsDoesNotError",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				procs, err := mngr.List(ctx, options.All)
				require.NoError(t, err)
				assert.Len(t, procs, 0)
			},
		},
		{
			Name: "ListAllProcesses",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())
				created, err := createProcs(ctx, opts, mngr, 10)
				require.NoError(t, err)
				assert.Len(t, created, 10)
				all, err := mngr.List(ctx, options.All)
				require.NoError(t, err)
				assert.Len(t, all, 10)
			},
		},
		{
			Name: "ListAllErrorsWithCanceledContext",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())
				created, err := createProcs(ctx, opts, mngr, 10)
				require.NoError(t, err)
				assert.Len(t, created, 10)

				cctx, cancel := context.WithCancel(ctx)
				cancel()
				procs, err := mngr.List(cctx, options.All)
				require.Error(t, err)
				assert.Nil(t, procs)
			},
		},
		{
			Name: "LongRunningProcessesAreListedAsRunning",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.SleepCreateOpts(20))
				procs, err := createProcs(ctx, opts, mngr, 10)
				require.NoError(t, err)
				assert.Len(t, procs, 10)

				procs, err = mngr.List(ctx, options.All)
				require.NoError(t, err)
				assert.Len(t, procs, 10)

				procs, err = mngr.List(ctx, options.Running)
				require.NoError(t, err)
				assert.Len(t, procs, 10)

				procs, err = mngr.List(ctx, options.Successful)
				require.NoError(t, err)
				assert.Len(t, procs, 0)
			},
		},
		{
			Name: "ListReturnsOneSuccessfulProcess",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())

				proc, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)

				_, err = proc.Wait(ctx)
				require.NoError(t, err)

				procs, err := mngr.List(ctx, options.Successful)
				require.NoError(t, err)

				require.Len(t, procs, 1)
				assert.Equal(t, proc.ID(), procs[0].ID())
			},
		},
		{
			Name: "ListReturnsOneFailedProcess",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.FalseCreateOpts())

				proc, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)
				_, err = proc.Wait(ctx)
				require.Error(t, err)

				procs, err := mngr.List(ctx, options.Failed)
				require.NoError(t, err)

				require.Len(t, procs, 1)
				assert.Equal(t, proc.ID(), procs[0].ID())
			},
		},
		{
			Name: "ListErrorsWithInvalidFilter",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				procs, err := mngr.List(ctx, options.Filter("foo"))
				assert.Error(t, err)
				assert.Empty(t, procs)
			},
		},
		{
			Name: "GetProcessErrorsWithNonexistentProcess",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				proc, err := mngr.Get(ctx, "foo")
				require.Error(t, err)
				assert.Nil(t, proc)
			},
		},
		{
			Name: "GetProcessReturnsMatchingProcess",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())
				proc, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)

				getProc, err := mngr.Get(ctx, proc.ID())
				require.NoError(t, err)
				assert.Equal(t, proc.ID(), getProc.ID())
			},
		},
		{
			Name: "GroupWithoutResultsDoesNotError",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				procs, err := mngr.Group(ctx, "foo")
				require.NoError(t, err)
				assert.Len(t, procs, 0)
			},
		},
		{
			Name: "GroupErrorsWithCanceledContext",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())
				_, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)

				cctx, cancel := context.WithCancel(ctx)
				cancel()
				procs, err := mngr.Group(cctx, "foo")
				require.Error(t, err)
				assert.Len(t, procs, 0)
			},
		},
		{
			Name: "GroupReturnsMatchingProcesses",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())
				proc, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)
				proc.Tag("foo")

				procs, err := mngr.Group(ctx, "foo")
				require.NoError(t, err)
				require.Len(t, procs, 1)
				assert.Equal(t, proc.ID(), procs[0].ID())
			},
		},
		{
			Name: "CloseEmptyManagerNoops",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, moidfyOpts testutil.OptsModify) {
				assert.NoError(t, mngr.Close(ctx))
			},
		},
		{
			Name: "CloseErrorsWithCanceledContext",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.SleepCreateOpts(5))
				_, err := createProcs(ctx, opts, mngr, 10)
				require.NoError(t, err)

				cctx, cancel := context.WithCancel(ctx)
				cancel()

				assert.Error(t, mngr.Close(cctx))
			},
		},
		{
			Name: "CloseSucceedsWithTerminatedProcesses",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())
				procs, err := createProcs(ctx, opts, mngr, 10)
				for _, proc := range procs {
					_, err = proc.Wait(ctx)
					require.NoError(t, err)
				}

				require.NoError(t, err)
				assert.NoError(t, mngr.Close(ctx))
			},
		},
		{
			Name: "CloserWithoutTriggersTerminatesProcesses",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				if runtime.GOOS == "windows" {
					t.Skip("manager close tests will error due to process termination on Windows")
				}
				opts := modifyOpts(testutil.SleepCreateOpts(5))

				_, err := createProcs(ctx, opts, mngr, 10)
				require.NoError(t, err)
				assert.NoError(t, mngr.Close(ctx))
			},
		},
		{
			Name: "ClearCausesDeletionOfCompletedProcesses",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())
				proc, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)

				sameProc, err := mngr.Get(ctx, proc.ID())
				require.NoError(t, err)
				require.Equal(t, proc.ID(), sameProc.ID())

				_, err = proc.Wait(ctx)
				require.NoError(t, err)

				mngr.Clear(ctx)
				nilProc, err := mngr.Get(ctx, proc.ID())
				require.Error(t, err)
				assert.Nil(t, nilProc)
			},
		},
		{
			Name: "ClearIsANoopForRunningProcesses",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.SleepCreateOpts(5))
				proc, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)

				mngr.Clear(ctx)
				sameProc, err := mngr.Get(ctx, proc.ID())
				require.NoError(t, err)
				assert.Equal(t, proc.ID(), sameProc.ID())
			},
		},
		{
			Name: "ClearSelectivelyDeletesOnlyCompletedProcesses",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				trueOpts := modifyOpts(testutil.TrueCreateOpts())
				trueProc, err := mngr.CreateProcess(ctx, trueOpts)
				require.NoError(t, err)

				sleepOpts := modifyOpts(testutil.SleepCreateOpts(5))
				sleepProc, err := mngr.CreateProcess(ctx, sleepOpts)
				require.NoError(t, err)

				_, err = trueProc.Wait(ctx)
				require.NoError(t, err)

				require.True(t, sleepProc.Running(ctx))

				mngr.Clear(ctx)

				sameSleepProc, err := mngr.Get(ctx, sleepProc.ID())
				require.NoError(t, err)
				assert.Equal(t, sleepProc.ID(), sameSleepProc.ID())

				nilProc, err := mngr.Get(ctx, trueProc.ID())
				require.Error(t, err)
				assert.Nil(t, nilProc)
			},
		},
		{
			Name: "CreateCommandPasses",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modify testutil.OptsModify) {
				cmd := mngr.CreateCommand(ctx)
				cmd.Add(testutil.TrueCreateOpts().Args)
				assert.NoError(t, cmd.Run(ctx))
			},
		},
		// kim: NOTE: this is a local test, not sure if it works on remote
		// interface
		{
			Name: "RunningCommandCreatesNewProcesses",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modify testutil.OptsModify) {
				cmd := mngr.CreateCommand(ctx)
				trueCmd := testutil.TrueCreateOpts().Args
				subCmds := [][]string{trueCmd, trueCmd, trueCmd}
				cmd.Extend(subCmds)
				require.NoError(t, cmd.Run(ctx))

				allProcs, err := mngr.List(ctx, options.All)
				require.NoError(t, err)

				assert.Len(t, allProcs, len(subCmds))
			},
		},
		// kim: NOTE: this is a local test, not sure if it works on remote
		// interface
		{

			Name: "CommandProcessIDsMatchManagedProcessIDs",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				cmd := mngr.CreateCommand(ctx)
				trueCmd := testutil.TrueCreateOpts().Args
				cmd.Extend([][]string{trueCmd, trueCmd, trueCmd})
				require.NoError(t, cmd.Run(ctx))

				allProcs, err := mngr.List(ctx, options.All)
				require.NoError(t, err)

				procsContainID := func(procs []Process, procID string) bool {
					for _, proc := range procs {
						if proc.ID() == procID {
							return true
						}
					}
					return false
				}

				for _, procID := range cmd.GetProcIDs() {
					assert.True(t, procsContainID(allProcs, procID))
				}
			},
		},
		{
			Name: "WriteFileSucceeds",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				tmpFile, err := ioutil.TempFile(testutil.BuildDirectory(), filepath.Base(t.Name()))
				require.NoError(t, err)
				defer func() {
					assert.NoError(t, os.RemoveAll(tmpFile.Name()))
				}()
				require.NoError(t, tmpFile.Close())

				opts := options.WriteFile{Path: tmpFile.Name(), Content: []byte("foo")}
				require.NoError(t, mngr.WriteFile(ctx, opts))

				content, err := ioutil.ReadFile(tmpFile.Name())
				require.NoError(t, err)

				assert.Equal(t, opts.Content, content)
			},
		},
		{
			Name: "WriteFileAcceptsContentFromReader",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				tmpFile, err := ioutil.TempFile(testutil.BuildDirectory(), filepath.Base(t.Name()))
				require.NoError(t, err)
				defer func() {
					assert.NoError(t, os.RemoveAll(tmpFile.Name()))
				}()
				require.NoError(t, tmpFile.Close())

				buf := []byte("foo")
				opts := options.WriteFile{Path: tmpFile.Name(), Reader: bytes.NewBuffer(buf)}
				require.NoError(t, mngr.WriteFile(ctx, opts))

				content, err := ioutil.ReadFile(tmpFile.Name())
				require.NoError(t, err)

				assert.Equal(t, buf, content)
			},
		},
		{
			Name: "WriteFileSucceedsWithLargeContent",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				tmpFile, err := ioutil.TempFile(testutil.BuildDirectory(), filepath.Base(t.Name()))
				require.NoError(t, err)
				defer func() {
					assert.NoError(t, os.RemoveAll(tmpFile.Name()))
				}()
				require.NoError(t, tmpFile.Close())

				const mb = 1024 * 1024
				opts := options.WriteFile{Path: tmpFile.Name(), Content: bytes.Repeat([]byte("foo"), mb)}
				require.NoError(t, mngr.WriteFile(ctx, opts))

				content, err := ioutil.ReadFile(tmpFile.Name())
				require.NoError(t, err)

				assert.Equal(t, opts.Content, content)
			},
		},
		{
			Name: "WriteFileSucceedsWithLargeContentFromReader",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				tmpFile, err := ioutil.TempFile(testutil.BuildDirectory(), filepath.Base(t.Name()))
				require.NoError(t, err)
				defer func() {
					assert.NoError(t, tmpFile.Close())
					assert.NoError(t, os.RemoveAll(tmpFile.Name()))
				}()

				const mb = 1024 * 1024
				buf := bytes.Repeat([]byte("foo"), 2*mb)
				opts := options.WriteFile{Path: tmpFile.Name(), Reader: bytes.NewBuffer(buf)}
				require.NoError(t, mngr.WriteFile(ctx, opts))

				content, err := ioutil.ReadFile(tmpFile.Name())
				require.NoError(t, err)

				assert.Equal(t, buf, content)
			},
		},
		{
			Name: "WriteFileSucceedsWithNoContent",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				path := filepath.Join(testutil.BuildDirectory(), filepath.Base(t.Name()))
				require.NoError(t, os.RemoveAll(path))
				defer func() {
					assert.NoError(t, os.RemoveAll(path))
				}()

				opts := options.WriteFile{Path: path}
				require.NoError(t, mngr.WriteFile(ctx, opts))

				stat, err := os.Stat(path)
				require.NoError(t, err)

				assert.Zero(t, stat.Size())
			},
		},
		{
			Name: "WriteFileFailsWithInvalidPath",
			Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify) {
				opts := options.WriteFile{Content: []byte("foo")}
				assert.Error(t, mngr.WriteFile(ctx, opts))
			},
		},
	}
}

// createProcs makes N identical processes with the given manager.
func createProcs(ctx context.Context, opts *options.Create, mngr Manager, num int) ([]Process, error) {
	catcher := grip.NewBasicCatcher()
	var procs []Process
	for i := 0; i < num; i++ {
		optsCopy := *opts

		proc, err := mngr.CreateProcess(ctx, &optsCopy)
		catcher.Add(err)
		if proc != nil {
			procs = append(procs, proc)
		}
	}

	return procs, catcher.Resolve()
}
