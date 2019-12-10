package remote

import (
	"context"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/send"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	sender := grip.GetSender()
	grip.Error(sender.SetLevel(send.LevelInfo{Threshhold: level.Info}))
	grip.Error(grip.SetLevel(sender))
}

type ClientTestCase struct {
	Name string
	Case func(context.Context, *testing.T, Manager)
}

func AddBasicClientTests(tests ...ClientTestCase) []ClientTestCase {
	return append([]ClientTestCase{
		{
			Name: "ValidateFixture",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				assert.NotNil(t, ctx)
				assert.NotNil(t, client)
			},
		},
		{
			Name: "IDReturnsNonempty",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				assert.NotEmpty(t, client.ID())
			},
		},
		{
			Name: "ProcEnvVarMatchesManagerID",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				opts := testutil.TrueCreateOpts()
				proc, err := client.CreateProcess(ctx, opts)
				require.NoError(t, err)
				info := proc.Info(ctx)
				require.NotEmpty(t, info.Options.Environment)
				assert.Equal(t, client.ID(), info.Options.Environment[jasper.ManagerEnvironID])
			},
		},
		{
			Name: "ListDoesNotErrorWhenEmpty",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				all, err := client.List(ctx, options.All)
				require.NoError(t, err)
				assert.Len(t, all, 0)
			},
		},
		{
			Name: "CreateProcessFails",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				proc, err := client.CreateProcess(ctx, &options.Create{})
				require.Error(t, err)
				assert.Nil(t, proc)
			},
		},
		{
			Name: "ListAllReturnsErrorWithCanceledContext",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				cctx, cancel := context.WithCancel(ctx)
				created, err := createProcs(ctx, testutil.TrueCreateOpts(), client, 10)
				require.NoError(t, err)
				assert.Len(t, created, 10)
				cancel()
				output, err := client.List(cctx, options.All)
				require.Error(t, err)
				assert.Nil(t, output)
			},
		},
		{
			Name: "LongRunningOperationsAreListedAsRunning",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				procs, err := createProcs(ctx, testutil.SleepCreateOpts(20), client, 10)
				require.NoError(t, err)
				assert.Len(t, procs, 10)

				procs, err = client.List(ctx, options.All)
				require.NoError(t, err)
				assert.Len(t, procs, 10)

				procs, err = client.List(ctx, options.Running)
				require.NoError(t, err)
				assert.Len(t, procs, 10)

				procs, err = client.List(ctx, options.Successful)
				require.NoError(t, err)
				assert.Len(t, procs, 0)
			},
		},
		{
			Name: "ListReturnsOneSuccessfulCommand",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				proc, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
				require.NoError(t, err)

				_, err = proc.Wait(ctx)
				require.NoError(t, err)

				listOut, err := client.List(ctx, options.Successful)
				require.NoError(t, err)

				if assert.Len(t, listOut, 1) {
					assert.Equal(t, listOut[0].ID(), proc.ID())
				}
			},
		},
		{
			Name: "GetMethodErrorsWithNoResponse",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				proc, err := client.Get(ctx, "foo")
				require.Error(t, err)
				assert.Nil(t, proc)
			},
		},
		{
			Name: "GetMethodReturnsMatchingDoc",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				proc, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
				require.NoError(t, err)

				ret, err := client.Get(ctx, proc.ID())
				require.NoError(t, err)
				assert.Equal(t, ret.ID(), proc.ID())
			},
		},
		{
			Name: "GroupDoesNotErrorWithoutResults",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				procs, err := client.Group(ctx, "foo")
				require.NoError(t, err)
				assert.Len(t, procs, 0)
			},
		},
		{
			Name: "GroupErrorsForCanceledContexts",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				_, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
				require.NoError(t, err)

				cctx, cancel := context.WithCancel(ctx)
				cancel()
				procs, err := client.Group(cctx, "foo")
				require.Error(t, err)
				assert.Len(t, procs, 0)
				assert.Contains(t, err.Error(), "canceled")
			},
		},
		{
			Name: "GroupPropagatesMatching",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				proc, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
				require.NoError(t, err)

				proc.Tag("foo")

				procs, err := client.Group(ctx, "foo")
				require.NoError(t, err)
				require.Len(t, procs, 1)
				assert.Equal(t, procs[0].ID(), proc.ID())
			},
		},
		{
			Name: "CloseEmptyManagerNoops",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				require.NoError(t, client.Close(ctx))
			},
		},
		{
			Name: "ClosersWithoutTriggersTerminatesProcesses",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				if runtime.GOOS == "windows" {
					t.Skip("the sleep tests don't block correctly on windows")
				}

				_, err := createProcs(ctx, testutil.SleepCreateOpts(100), client, 10)
				require.NoError(t, err)
				assert.NoError(t, client.Close(ctx))
			},
		},
		{
			Name: "CloseErrorsWithCanceledContext",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				_, err := createProcs(ctx, testutil.SleepCreateOpts(100), client, 10)
				require.NoError(t, err)

				cctx, cancel := context.WithCancel(ctx)
				cancel()

				err = client.Close(cctx)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "canceled")
			},
		},
		{
			Name: "CloseSucceedsWithTerminatedProcesses",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				procs, err := createProcs(ctx, testutil.TrueCreateOpts(), client, 10)
				for _, p := range procs {
					_, err = p.Wait(ctx)
					require.NoError(t, err)
				}

				require.NoError(t, err)
				assert.NoError(t, client.Close(ctx))
			},
		},
		{
			Name: "WaitingOnNonExistentProcessErrors",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				proc, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
				require.NoError(t, err)

				_, err = proc.Wait(ctx)
				require.NoError(t, err)

				client.Clear(ctx)

				_, err = proc.Wait(ctx)
				require.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), "problem finding process"))
			},
		},
		{
			Name: "ClearCausesDeletionOfProcesses",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				opts := testutil.TrueCreateOpts()
				proc, err := client.CreateProcess(ctx, opts)
				require.NoError(t, err)
				sameProc, err := client.Get(ctx, proc.ID())
				require.NoError(t, err)
				require.Equal(t, proc.ID(), sameProc.ID())
				_, err = proc.Wait(ctx)
				require.NoError(t, err)
				client.Clear(ctx)
				nilProc, err := client.Get(ctx, proc.ID())
				require.Error(t, err)
				assert.Nil(t, nilProc)
			},
		},
		{
			Name: "ClearIsANoopForActiveProcesses",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				opts := testutil.SleepCreateOpts(20)
				proc, err := client.CreateProcess(ctx, opts)
				require.NoError(t, err)
				client.Clear(ctx)
				sameProc, err := client.Get(ctx, proc.ID())
				require.NoError(t, err)
				assert.Equal(t, proc.ID(), sameProc.ID())
				require.NoError(t, jasper.Terminate(ctx, proc)) // Clean up
			},
		},
		{
			Name: "ClearSelectivelyDeletesOnlyDeadProcesses",
			Case: func(ctx context.Context, t *testing.T, client Manager) {
				trueOpts := testutil.TrueCreateOpts()
				lsProc, err := client.CreateProcess(ctx, trueOpts)
				require.NoError(t, err)

				sleepOpts := testutil.SleepCreateOpts(20)
				sleepProc, err := client.CreateProcess(ctx, sleepOpts)
				require.NoError(t, err)

				_, err = lsProc.Wait(ctx)
				require.NoError(t, err)

				client.Clear(ctx)

				sameSleepProc, err := client.Get(ctx, sleepProc.ID())
				require.NoError(t, err)
				assert.Equal(t, sleepProc.ID(), sameSleepProc.ID())

				nilProc, err := client.Get(ctx, lsProc.ID())
				require.Error(t, err)
				assert.Nil(t, nilProc)
				require.NoError(t, jasper.Terminate(ctx, sleepProc)) // Clean up
			},
		},
	}, tests...)
}

type ProcessTestCase struct {
	Name string
	Case func(context.Context, *testing.T, *options.Create, jasper.ProcessConstructor)
}

func AddBasicProcessTests(tests ...ProcessTestCase) []ProcessTestCase {
	return append([]ProcessTestCase{
		{
			Name: "WithPopulatedArgsCommandCreationPasses",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				assert.NotZero(t, opts.Args)
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				assert.NotNil(t, proc)
			},
		},
		{
			Name: "ErrorToCreateWithInvalidArgs",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				opts.Args = []string{}
				proc, err := makep(ctx, opts)
				require.Error(t, err)
				assert.Nil(t, proc)
			},
		},
		{
			Name: "WithCanceledContextProcessCreationFails",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				pctx, pcancel := context.WithCancel(ctx)
				pcancel()
				proc, err := makep(pctx, opts)
				require.Error(t, err)
				assert.Nil(t, proc)
			},
		},
		{
			Name: "CanceledContextTimesOutEarly",
			Case: func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
				pctx, pcancel := context.WithTimeout(ctx, 5*time.Second)
				defer pcancel()
				startAt := time.Now()
				opts := testutil.SleepCreateOpts(20)
				proc, err := makep(pctx, opts)
				require.NoError(t, err)
				require.NotNil(t, proc)

				time.Sleep(5 * time.Millisecond) // let time pass...
				assert.False(t, proc.Info(ctx).Successful)
				assert.True(t, time.Since(startAt) < 20*time.Second)
			},
		},
		{
			Name: "ProcessLacksTagsByDefault",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				tags := proc.GetTags()
				assert.Empty(t, tags)
			},
		},
		{
			Name: "ProcessTagsPersist",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				opts.Tags = []string{"foo"}
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				tags := proc.GetTags()
				assert.Contains(t, tags, "foo")
			},
		},
		{
			Name: "InfoHasMatchingID",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				_, err = proc.Wait(ctx)
				require.NoError(t, err)
				assert.Equal(t, proc.ID(), proc.Info(ctx).ID)
			},
		},
		{
			Name: "ResetTags",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				proc.Tag("foo")
				assert.Contains(t, proc.GetTags(), "foo")
				proc.ResetTags()
				assert.Len(t, proc.GetTags(), 0)
			},
		},
		{
			Name: "TagsAreSetLike",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)

				for i := 0; i < 100; i++ {
					proc.Tag("foo")
				}

				assert.Len(t, proc.GetTags(), 1)
				proc.Tag("bar")
				assert.Len(t, proc.GetTags(), 2)
			},
		},
		{
			Name: "CompleteIsTrueAfterWait",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				time.Sleep(10 * time.Millisecond) // give the process time to start background machinery
				_, err = proc.Wait(ctx)
				assert.NoError(t, err)
				assert.True(t, proc.Complete(ctx))
			},
		},
		{
			Name: "WaitReturnsWithCanceledContext",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				opts.Args = []string{"sleep", "10"}
				pctx, pcancel := context.WithCancel(ctx)
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				assert.True(t, proc.Running(ctx))
				assert.NoError(t, err)
				pcancel()
				_, err = proc.Wait(pctx)
				assert.Error(t, err)
			},
		},
		{
			Name: "RegisterTriggerErrorsForNil",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				assert.Error(t, proc.RegisterTrigger(ctx, nil))
			},
		},
		{
			Name: "RegisterSignalTriggerIDErrorsForExitedProcess",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				_, err = proc.Wait(ctx)
				assert.NoError(t, err)
				assert.Error(t, proc.RegisterSignalTriggerID(ctx, jasper.CleanTerminationSignalTrigger))
			},
		},
		{
			Name: "RegisterSignalTriggerIDFailsWithInvalidTriggerID",
			Case: func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
				opts := testutil.SleepCreateOpts(3)
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				assert.Error(t, proc.RegisterSignalTriggerID(ctx, jasper.SignalTriggerID(-1)))
			},
		},
		{
			Name: "RegisterSignalTriggerIDPassesWithValidTriggerID",
			Case: func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
				opts := testutil.SleepCreateOpts(3)
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				assert.NoError(t, proc.RegisterSignalTriggerID(ctx, jasper.CleanTerminationSignalTrigger))
			},
		},
		{
			Name: "WaitOnRespawnedProcessDoesNotError",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				require.NotNil(t, proc)
				_, err = proc.Wait(ctx)
				require.NoError(t, err)

				newProc, err := proc.Respawn(ctx)
				require.NoError(t, err)
				_, err = newProc.Wait(ctx)
				assert.NoError(t, err)
			},
		},
		{
			Name: "RespawnedProcessGivesSameResult",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				require.NotNil(t, proc)

				_, err = proc.Wait(ctx)
				require.NoError(t, err)
				procExitCode := proc.Info(ctx).ExitCode

				newProc, err := proc.Respawn(ctx)
				require.NoError(t, err)
				_, err = newProc.Wait(ctx)
				require.NoError(t, err)
				assert.Equal(t, procExitCode, newProc.Info(ctx).ExitCode)
			},
		},
		{
			Name: "RespawningFinishedProcessIsOK",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				require.NotNil(t, proc)
				_, err = proc.Wait(ctx)
				require.NoError(t, err)

				newProc, err := proc.Respawn(ctx)
				assert.NoError(t, err)
				_, err = newProc.Wait(ctx)
				require.NoError(t, err)
				assert.True(t, newProc.Info(ctx).Successful)
			},
		},
		{
			Name: "RespawningRunningProcessIsOK",
			Case: func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
				opts := testutil.SleepCreateOpts(2)
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				require.NotNil(t, proc)

				newProc, err := proc.Respawn(ctx)
				assert.NoError(t, err)
				_, err = newProc.Wait(ctx)
				require.NoError(t, err)
				assert.True(t, newProc.Info(ctx).Successful)
			},
		},
		{
			Name: "RespawnShowsConsistentStateValues",
			Case: func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
				opts := testutil.SleepCreateOpts(3)
				proc, err := makep(ctx, opts)
				require.NoError(t, err)
				require.NotNil(t, proc)
				_, err = proc.Wait(ctx)
				require.NoError(t, err)

				newProc, err := proc.Respawn(ctx)
				require.NoError(t, err)
				assert.True(t, newProc.Running(ctx))
				_, err = newProc.Wait(ctx)
				require.NoError(t, err)
				assert.True(t, proc.Complete(ctx))
			},
		},
		{
			Name: "WaitGivesSuccessfulExitCode",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, testutil.TrueCreateOpts())
				require.NoError(t, err)
				require.NotNil(t, proc)
				exitCode, err := proc.Wait(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 0, exitCode)
			},
		},
		{
			Name: "WaitGivesFailureExitCode",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, testutil.FalseCreateOpts())
				require.NoError(t, err)
				require.NotNil(t, proc)
				exitCode, err := proc.Wait(ctx)
				require.Error(t, err)
				assert.Equal(t, 1, exitCode)
			},
		},
		{
			Name: "WaitGivesProperExitCodeOnSignalDeath",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, testutil.SleepCreateOpts(100))
				require.NoError(t, err)
				require.NotNil(t, proc)
				sig := syscall.SIGTERM
				require.NoError(t, proc.Signal(ctx, sig))
				exitCode, err := proc.Wait(ctx)
				require.Error(t, err)
				if runtime.GOOS == "windows" {
					assert.Equal(t, 1, exitCode)
				} else {
					assert.Equal(t, int(sig), exitCode)
				}
			},
		},
		{
			Name: "WaitGivesNegativeOneOnAlternativeError",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				cctx, cancel := context.WithCancel(ctx)
				proc, err := makep(ctx, testutil.SleepCreateOpts(100))
				require.NoError(t, err)
				require.NotNil(t, proc)

				var exitCode int
				waitFinished := make(chan bool)
				go func() {
					exitCode, err = proc.Wait(cctx)
					waitFinished <- true
				}()
				cancel()
				select {
				case <-waitFinished:
					require.Error(t, err)
					assert.Equal(t, -1, exitCode)
				case <-ctx.Done():
					assert.Fail(t, "call to Wait() took too long to finish")
				}
				require.NoError(t, jasper.Terminate(ctx, proc)) // Clean up.

			},
		},
		{
			Name: "InfoHasTimeoutWhenProcessTimesOut",
			Case: func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
				opts := testutil.SleepCreateOpts(100)
				opts.Timeout = time.Second
				opts.TimeoutSecs = 1
				proc, err := makep(ctx, opts)
				require.NoError(t, err)

				exitCode, err := proc.Wait(ctx)
				assert.Error(t, err)
				if runtime.GOOS == "windows" {
					assert.Equal(t, 1, exitCode)
				} else {
					assert.Equal(t, int(syscall.SIGKILL), exitCode)
				}
				info := proc.Info(ctx)
				assert.True(t, info.Timeout)
			},
		},
		{
			Name: "CallingSignalOnDeadProcessDoesError",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
				proc, err := makep(ctx, opts)
				require.NoError(t, err)

				_, err = proc.Wait(ctx)
				assert.NoError(t, err)

				err = proc.Signal(ctx, syscall.SIGTERM)
				require.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), "cannot signal a process that has terminated"))
			},
		},
	}, tests...)

}
