package remote

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ClientTestCase struct {
	Name string
	Case func(context.Context, *testing.T, Manager)
}

func BasicClientTests() []ClientTestCase {
	return []ClientTestCase{
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
			"ListDoesNotErrorWhenEmpty": func(ctx context.Context, t *testing.T, client Manager) {
				all, err := client.List(ctx, options.All)
				require.NoError(t, err)
				assert.Len(t, all, 0)
			},
		},
		{
			"CreateProcessFails": func(ctx context.Context, t *testing.T, client Manager) {
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

		"GetMethodReturnsMatchingDoc": func(ctx context.Context, t *testing.T, client Manager) {
			proc, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
			require.NoError(t, err)

			ret, err := client.Get(ctx, proc.ID())
			require.NoError(t, err)
			assert.Equal(t, ret.ID(), proc.ID())
		},
		"GroupDoesNotErrorWithoutResults": func(ctx context.Context, t *testing.T, client Manager) {
			procs, err := client.Group(ctx, "foo")
			require.NoError(t, err)
			assert.Len(t, procs, 0)
		},
		"GroupErrorsForCanceledContexts": func(ctx context.Context, t *testing.T, client Manager) {
			_, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
			require.NoError(t, err)

			cctx, cancel := context.WithCancel(ctx)
			cancel()
			procs, err := client.Group(cctx, "foo")
			require.Error(t, err)
			assert.Len(t, procs, 0)
			assert.Contains(t, err.Error(), "canceled")
		},
		"GroupPropagatesMatching": func(ctx context.Context, t *testing.T, client Manager) {
			proc, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
			require.NoError(t, err)

			proc.Tag("foo")

			procs, err := client.Group(ctx, "foo")
			require.NoError(t, err)
			require.Len(t, procs, 1)
			assert.Equal(t, procs[0].ID(), proc.ID())
		},
		"CloseEmptyManagerNoops": func(ctx context.Context, t *testing.T, client Manager) {
			require.NoError(t, client.Close(ctx))
		},
		"ClosersWithoutTriggersTerminatesProcesses": func(ctx context.Context, t *testing.T, client Manager) {
			if runtime.GOOS == "windows" {
				t.Skip("the sleep tests don't block correctly on windows")
			}

			_, err := createProcs(ctx, testutil.SleepCreateOpts(100), client, 10)
			require.NoError(t, err)
			assert.NoError(t, client.Close(ctx))
		},
		"CloseErrorsWithCanceledContext": func(ctx context.Context, t *testing.T, client Manager) {
			_, err := createProcs(ctx, testutil.SleepCreateOpts(100), client, 10)
			require.NoError(t, err)

			cctx, cancel := context.WithCancel(ctx)
			cancel()

			err = client.Close(cctx)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "canceled")
		},
		"CloseSucceedsWithTerminatedProcesses": func(ctx context.Context, t *testing.T, client Manager) {
			procs, err := createProcs(ctx, testutil.TrueCreateOpts(), client, 10)
			for _, p := range procs {
				_, err = p.Wait(ctx)
				require.NoError(t, err)
			}

			require.NoError(t, err)
			assert.NoError(t, client.Close(ctx))
		},
		"WaitingOnNonExistentProcessErrors": func(ctx context.Context, t *testing.T, client Manager) {
			proc, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
			require.NoError(t, err)

			_, err = proc.Wait(ctx)
			require.NoError(t, err)

			client.Clear(ctx)

			_, err = proc.Wait(ctx)
			require.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), "problem finding process"))
		},
		"ClearCausesDeletionOfProcesses": func(ctx context.Context, t *testing.T, client Manager) {
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
		"ClearIsANoopForActiveProcesses": func(ctx context.Context, t *testing.T, client Manager) {
			opts := testutil.SleepCreateOpts(20)
			proc, err := client.CreateProcess(ctx, opts)
			require.NoError(t, err)
			client.Clear(ctx)
			sameProc, err := client.Get(ctx, proc.ID())
			require.NoError(t, err)
			assert.Equal(t, proc.ID(), sameProc.ID())
			require.NoError(t, jasper.Terminate(ctx, proc)) // Clean up
		},
		"ClearSelectivelyDeletesOnlyDeadProcesses": func(ctx context.Context, t *testing.T, client Manager) {
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
	}

}
