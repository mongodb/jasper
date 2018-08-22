package jrpc

import (
	"context"
	"testing"

	"github.com/mongodb/jasper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJRPCManager(t *testing.T) {
	assert.NotPanics(t, func() {
		NewJRPCManager(nil)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for mname, factory := range map[string]func(ctx context.Context, t *testing.T) jasper.Manager{
		"Blocking": func(ctx context.Context, t *testing.T) jasper.Manager {
			mngr := jasper.NewLocalManagerBlockingProcesses()
			addr, err := startJRPC(ctx, mngr)
			require.NoError(t, err)

			client, err := getClient(ctx, addr)
			require.NoError(t, err)

			return client
		},
		"NotBlocking": func(ctx context.Context, t *testing.T) jasper.Manager {
			mngr := jasper.NewLocalManager()
			addr, err := startJRPC(ctx, mngr)
			require.NoError(t, err)

			client, err := getClient(ctx, addr)
			require.NoError(t, err)

			return client
		},
	} {
		t.Run(mname, func(t *testing.T) {
			for name, test := range map[string]func(context.Context, *testing.T, jasper.Manager){
				"ValidateFixture": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					assert.NotNil(t, ctx)
					assert.NotNil(t, manager)
				},
				"ListErrorsWhenEmpty": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					all, err := manager.List(ctx, jasper.All)
					assert.Error(t, err)
					assert.Len(t, all, 0)
					assert.Contains(t, err.Error(), "no processes")
				},
				"CreateProcessFails": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					proc, err := manager.Create(ctx, &jasper.CreateOptions{})
					assert.Error(t, err)
					assert.Nil(t, proc)
				},
				"ListAllReturnsErrorWithCancledContext": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					cctx, cancel := context.WithCancel(ctx)
					created, err := createProcs(ctx, trueCreateOpts(), manager, 10)
					assert.NoError(t, err)
					assert.Len(t, created, 10)
					cancel()
					output, err := manager.List(cctx, jasper.All)
					assert.Error(t, err)
					assert.Nil(t, output)
				},
				"LongRunningOperationsAreListedAsRunning": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					procs, err := createProcs(ctx, sleepCreateOpts(20), manager, 10)
					assert.NoError(t, err)
					assert.Len(t, procs, 10)

					procs, err = manager.List(ctx, jasper.All)
					assert.NoError(t, err)
					assert.Len(t, procs, 10)

					procs, err = manager.List(ctx, jasper.Running)
					assert.NoError(t, err)
					assert.Len(t, procs, 10)

					procs, err = manager.List(ctx, jasper.Successful)
					assert.Error(t, err)
					assert.Len(t, procs, 0)
				},
				"ListReturnsOneSuccessfulCommand": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					proc, err := manager.Create(ctx, trueCreateOpts())
					require.NoError(t, err)

					assert.NoError(t, proc.Wait(ctx))

					listOut, err := manager.List(ctx, jasper.Successful)
					assert.NoError(t, err)

					if assert.Len(t, listOut, 1) {
						assert.Equal(t, listOut[0].ID(), proc.ID())
					}
				},
				"GetMethodErrorsWithNoResponse": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					proc, err := manager.Get(ctx, "foo")
					assert.Error(t, err)
					assert.Nil(t, proc)
				},
				"GetMethodReturnsMatchingDoc": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					proc, err := manager.Create(ctx, trueCreateOpts())
					require.NoError(t, err)

					ret, err := manager.Get(ctx, proc.ID())
					if assert.NoError(t, err) {
						assert.Equal(t, ret.ID(), proc.ID())
					}
				},
				"GroupErrorsWithoutResults": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					procs, err := manager.Group(ctx, "foo")
					assert.Error(t, err)
					assert.Len(t, procs, 0)
					assert.Contains(t, err.Error(), "no jobs")
				},
				"GroupErrorsForCanceledContexts": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					_, err := manager.Create(ctx, trueCreateOpts())
					assert.NoError(t, err)

					cctx, cancel := context.WithCancel(ctx)
					cancel()
					procs, err := manager.Group(cctx, "foo")
					assert.Error(t, err)
					assert.Len(t, procs, 0)
					assert.Contains(t, err.Error(), "canceled")
				},
				"GroupPropgatesMatching": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					proc, err := manager.Create(ctx, trueCreateOpts())
					require.NoError(t, err)

					proc.Tag("foo")

					procs, err := manager.Group(ctx, "foo")
					require.NoError(t, err)
					require.Len(t, procs, 1)
					assert.Equal(t, procs[0].ID(), proc.ID())
				},
				"CloseEmptyManagerNoops": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					assert.NoError(t, manager.Close(ctx))
				},
				"ClosersWithoutTriggersTerminatesProcesses": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					_, err := createProcs(ctx, sleepCreateOpts(100), manager, 10)
					assert.NoError(t, err)
					assert.NoError(t, manager.Close(ctx))
				},
				"CloseErrorsWithCanceledContext": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					_, err := createProcs(ctx, sleepCreateOpts(100), manager, 10)
					assert.NoError(t, err)

					cctx, cancel := context.WithCancel(ctx)
					cancel()

					err = manager.Close(cctx)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "canceled")
				},
				"CloseErrorsWithTerminatedProcesses": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					procs, err := createProcs(ctx, trueCreateOpts(), manager, 10)
					for _, p := range procs {
						assert.NoError(t, p.Wait(ctx))
					}

					assert.NoError(t, err)
					assert.Error(t, manager.Close(ctx))
				},
				// "": func(ctx context.Context, t *testing.T, manager jasper.Manager) {},
				// "": func(ctx context.Context, t *testing.T, manager jasper.Manager) {},
			} {
				t.Run(name, func(t *testing.T) {
					tctx, cancel := context.WithTimeout(ctx, taskTimeout)
					defer cancel()
					test(tctx, t, factory(tctx, t))
				})
			}
		})

	}

}
