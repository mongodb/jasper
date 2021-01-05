package jasper

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerImplementations(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for managerName, makeManager := range map[string]func(context.Context, *testing.T) Manager{
		"Basic/NoLock": func(_ context.Context, _ *testing.T) Manager {
			return &basicProcessManager{
				id:      "id",
				loggers: NewLoggingCache(),
				procs:   map[string]Process{},
			}
		},
		"Basic/Lock": func(_ context.Context, t *testing.T) Manager {
			synchronizedManager, err := NewSynchronizedManager(false)
			require.NoError(t, err)
			return synchronizedManager
		},
		"SelfClearing/NoLock": func(_ context.Context, t *testing.T) Manager {
			selfClearingManager, err := NewSelfClearingProcessManager(10, false)
			require.NoError(t, err)
			return selfClearingManager
		},
		"Remote/NoLock/NilOptions": func(_ context.Context, t *testing.T) Manager {
			m, err := newBasicProcessManager(map[string]Process{}, false, false)
			require.NoError(t, err)
			return NewRemoteManager(m, nil)
		},
		"Docker/NoLock": func(_ context.Context, t *testing.T) Manager {
			m, err := newBasicProcessManager(map[string]Process{}, false, false)
			require.NoError(t, err)
			image := os.Getenv("DOCKER_IMAGE")
			if image == "" {
				image = testutil.DefaultDockerImage
			}
			return NewDockerManager(m, &options.Docker{
				Image: image,
			})
		},
	} {
		if testutil.IsDockerCase(managerName) {
			testutil.SkipDockerIfUnsupported(t)
		}

		testCases := append(ManagerTests(), []ManagerTestCase{
			{
				Name: "CloseExecutesClosersForProcesses",
				Case: func(ctx context.Context, t *testing.T, manager Manager, modify testutil.OptsModify) {
					if runtime.GOOS == "windows" {
						t.Skip("manager close tests will error due to process termination on Windows")
					}
					opts := testutil.SleepCreateOpts(5)
					modify(opts)

					count := 0
					countIncremented := make(chan bool, 1)
					opts.RegisterCloser(func() (_ error) {
						defer close(countIncremented)
						count++
						countIncremented <- true
						return
					})

					_, err := manager.CreateProcess(ctx, opts)
					require.NoError(t, err)

					assert.Equal(t, count, 0)
					require.NoError(t, manager.Close(ctx))
					select {
					case <-ctx.Done():
						assert.Fail(t, "process took too long to run closers")
					case <-countIncremented:
						assert.Equal(t, 1, count)
					}
				},
			},
			{
				Name: "RegisterProcessErrorsForNilProcess",
				Case: func(ctx context.Context, t *testing.T, manager Manager, modify testutil.OptsModify) {
					err := manager.Register(ctx, nil)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "not defined")
				},
			},
			{
				Name: "RegisterProcessErrorsForCanceledContext",
				Case: func(ctx context.Context, t *testing.T, manager Manager, modify testutil.OptsModify) {
					cctx, cancel := context.WithCancel(ctx)
					cancel()

					opts := testutil.TrueCreateOpts()
					modify(opts)

					proc, err := newBlockingProcess(ctx, opts)
					require.NoError(t, err)
					err = manager.Register(cctx, proc)
					require.Error(t, err)
					assert.Contains(t, err.Error(), context.Canceled.Error())
				},
			},
			{
				Name: "RegisterProcessErrorsWhenMissingID",
				Case: func(ctx context.Context, t *testing.T, manager Manager, modify testutil.OptsModify) {
					proc := &blockingProcess{}
					assert.Equal(t, proc.ID(), "")
					err := manager.Register(ctx, proc)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "malformed")
				},
			},
			{
				Name: "RegisterProcessModifiesManagerState",
				Case: func(ctx context.Context, t *testing.T, manager Manager, modify testutil.OptsModify) {
					opts := testutil.TrueCreateOpts()
					modify(opts)

					proc, err := newBlockingProcess(ctx, opts)
					require.NoError(t, err)
					err = manager.Register(ctx, proc)
					require.NoError(t, err)

					procs, err := manager.List(ctx, options.All)
					require.NoError(t, err)
					require.True(t, len(procs) >= 1)

					assert.Equal(t, procs[0].ID(), proc.ID())
				},
			},
			{
				Name: "RegisterProcessErrorsForDuplicateProcess",
				Case: func(ctx context.Context, t *testing.T, manager Manager, modify testutil.OptsModify) {
					opts := testutil.TrueCreateOpts()
					modify(opts)

					proc, err := newBlockingProcess(ctx, opts)
					require.NoError(t, err)
					assert.NotEmpty(t, proc)
					err = manager.Register(ctx, proc)
					require.NoError(t, err)
					err = manager.Register(ctx, proc)
					assert.Error(t, err)
				},
			},
			{
				Name: "ManagerCallsOptionsCloseByDefault",
				Case: func(ctx context.Context, t *testing.T, manager Manager, modify testutil.OptsModify) {
					opts := &options.Create{}
					modify(opts)
					opts.Args = []string{"echo", "foobar"}
					count := 0
					countIncremented := make(chan bool, 1)
					opts.RegisterCloser(func() (_ error) {
						count++
						countIncremented <- true
						close(countIncremented)
						return
					})

					proc, err := manager.CreateProcess(ctx, opts)
					require.NoError(t, err)
					_, err = proc.Wait(ctx)
					require.NoError(t, err)

					select {
					case <-ctx.Done():
						assert.Fail(t, "process took too long to run closers")
					case <-countIncremented:
						assert.Equal(t, 1, count)
					}
				},
			},
		}...)
		t.Run(managerName, func(t *testing.T) {
			for _, testCase := range testCases {
				t.Run(testCase.Name, func(t *testing.T) {
					for procName, modifyOpts := range map[string]testutil.OptsModify{
						"BasicProcess": func(opts *options.Create) *options.Create {
							opts.Implementation = options.ProcessImplementationBasic
							return opts
						},
						"BlockingProcess": func(opts *options.Create) *options.Create {
							opts.Implementation = options.ProcessImplementationBlocking
							return opts
						},
					} {
						t.Run(procName, func(t *testing.T) {
							tctx, tcancel := context.WithTimeout(ctx, testutil.ManagerTestTimeout)
							defer tcancel()
							testCase.Case(tctx, t, makeManager(tctx, t), modifyOpts)
						})
					}
				})
			}
		})
	}
}

func TestTrackedManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for managerName, makeManager := range map[string]func() *basicProcessManager{
		"Basic": func() *basicProcessManager {
			return &basicProcessManager{
				procs:   map[string]Process{},
				loggers: NewLoggingCache(),
				tracker: &mockProcessTracker{
					Infos: []ProcessInfo{},
				},
			}
		},
	} {
		t.Run(managerName, func(t *testing.T) {
			for testName, testCase := range map[string]func(context.Context, *testing.T, *basicProcessManager, *options.Create){
				"CreateProcessTracksProcess": func(ctx context.Context, t *testing.T, manager *basicProcessManager, opts *options.Create) {
					proc, err := manager.CreateProcess(ctx, opts)
					require.NoError(t, err)
					assert.Len(t, manager.procs, 1)

					mockTracker, ok := manager.tracker.(*mockProcessTracker)
					require.True(t, ok)
					require.Len(t, mockTracker.Infos, 1)
					assert.Equal(t, proc.Info(ctx), mockTracker.Infos[0])
				},
				"CreateCommandTracksCommandAfterRun": func(ctx context.Context, t *testing.T, manager *basicProcessManager, opts *options.Create) {
					err := manager.CreateCommand(ctx).Add(opts.Args).Background(true).Run(ctx)
					require.NoError(t, err)
					assert.Len(t, manager.procs, 1)

					mockTracker, ok := manager.tracker.(*mockProcessTracker)
					require.True(t, ok)
					require.Len(t, mockTracker.Infos, 1)
					assert.NotZero(t, mockTracker.Infos[0])
				},
				"DoNotTrackProcessIfCreateProcessDoesNotMakeProcess": func(ctx context.Context, t *testing.T, manager *basicProcessManager, opts *options.Create) {
					opts.Args = []string{"foo"}
					_, err := manager.CreateProcess(ctx, opts)
					require.Error(t, err)
					assert.Len(t, manager.procs, 0)

					mockTracker, ok := manager.tracker.(*mockProcessTracker)
					require.True(t, ok)
					assert.Len(t, mockTracker.Infos, 0)
				},
				"DoNotTrackProcessIfCreateCommandDoesNotMakeProcess": func(ctx context.Context, t *testing.T, manager *basicProcessManager, opts *options.Create) {
					opts.Args = []string{"foo"}
					cmd := manager.CreateCommand(ctx).Add(opts.Args).Background(true)
					cmd.opts.Process = *opts
					err := cmd.Run(ctx)
					require.Error(t, err)
					assert.Len(t, manager.procs, 0)

					mockTracker, ok := manager.tracker.(*mockProcessTracker)
					require.True(t, ok)
					assert.Len(t, mockTracker.Infos, 0)
				},
				"CloseCleansUpProcesses": func(ctx context.Context, t *testing.T, manager *basicProcessManager, opts *options.Create) {
					cmd := manager.CreateCommand(ctx).Background(true).Add(opts.Args)
					cmd.opts.Process = *opts
					require.NoError(t, cmd.Run(ctx))
					assert.Len(t, manager.procs, 1)

					mockTracker, ok := manager.tracker.(*mockProcessTracker)
					require.True(t, ok)
					require.Len(t, mockTracker.Infos, 1)
					assert.NotZero(t, mockTracker.Infos[0])

					require.NoError(t, manager.Close(ctx))
					assert.Len(t, mockTracker.Infos, 0)
					require.NoError(t, manager.Close(ctx))
				},
				"CloseWithNoProcessesIsNotError": func(ctx context.Context, t *testing.T, manager *basicProcessManager, opts *options.Create) {
					mockTracker, ok := manager.tracker.(*mockProcessTracker)
					require.True(t, ok)

					require.NoError(t, manager.Close(ctx))
					assert.Len(t, mockTracker.Infos, 0)
					require.NoError(t, manager.Close(ctx))
					assert.Len(t, mockTracker.Infos, 0)
				},
				"DoubleCloseIsNotError": func(ctx context.Context, t *testing.T, manager *basicProcessManager, opts *options.Create) {
					cmd := manager.CreateCommand(ctx).Background(true).Add(opts.Args)
					cmd.opts.Process = *opts

					require.NoError(t, cmd.Run(ctx))
					assert.Len(t, manager.procs, 1)

					mockTracker, ok := manager.tracker.(*mockProcessTracker)
					require.True(t, ok)
					require.Len(t, mockTracker.Infos, 1)
					assert.NotZero(t, mockTracker.Infos[0])

					require.NoError(t, manager.Close(ctx))
					assert.Len(t, mockTracker.Infos, 0)
					require.NoError(t, manager.Close(ctx))
					assert.Len(t, mockTracker.Infos, 0)
				},
				// "": func(ctx context.Context, t *testing.T, manager Manager, modify testutil.OptsModify) {},
			} {
				t.Run(testName, func(t *testing.T) {
					for procName, modifyOpts := range map[string]testutil.OptsModify{
						"BasicProcess": func(opts *options.Create) *options.Create {
							opts.Implementation = options.ProcessImplementationBasic
							return opts
						},
						"BlockingProcess": func(opts *options.Create) *options.Create {
							opts.Implementation = options.ProcessImplementationBlocking
							return opts
						},
					} {
						t.Run(procName, func(t *testing.T) {
							tctx, cancel := context.WithTimeout(ctx, testutil.ManagerTestTimeout)
							defer cancel()
							opts := testutil.SleepCreateOpts(1)
							opts = modifyOpts(opts)
							testCase(tctx, t, makeManager(), opts)
						})
					}
				})
			}
		})
	}
}
