package jrpc

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: these tests are largely copied directly from the top level
// package into this package to avoid an import cycle.

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
					t.Skip("for now,")
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
					if runtime.GOOS == "windows" {
						t.Skip("the sleep tests don't block correctly on windows")
					}

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
					if runtime.GOOS == "windows" {
						t.Skip("context times out on windows")
					}

					procs, err := createProcs(ctx, trueCreateOpts(), manager, 10)
					for _, p := range procs {
						assert.NoError(t, p.Wait(ctx))
					}

					assert.NoError(t, err)
					assert.Error(t, manager.Close(ctx))
				},
				// "": func(ctx context.Context, t *testing.T, manager jasper.Manager) {},
				// "": func(ctx context.Context, t *testing.T, manager jasper.Manager) {},

				///////////////////////////////////
				//
				// The following test cases are added
				// specifically for the jrpc case

				"RegisterIsDisabled": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					err := manager.Register(ctx, nil)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "cannot register")
				},
				"ListErrorsWithEmpty": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					procs, err := manager.List(ctx, jasper.All)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "no processes")
					assert.Len(t, procs, 0)
				},
				"CreateProcessReturnsCorrectExample": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					proc, err := manager.Create(ctx, trueCreateOpts())
					assert.NoError(t, err)
					assert.NotNil(t, proc)
					assert.NotZero(t, proc.ID())

					fetched, err := manager.Get(ctx, proc.ID())
					assert.NoError(t, err)
					assert.NotNil(t, fetched)
					assert.Equal(t, proc.ID(), fetched.ID())
				},
				"DownloadFileCreatesResource": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					jrpcManager, ok := manager.(*jrpcManager)
					require.True(t, ok)

					cwd, err := os.Getwd()
					require.NoError(t, err)
					file, err := ioutil.TempFile(filepath.Join(filepath.Dir(cwd), "build"), "out.txt")
					require.NoError(t, err)
					defer os.Remove(file.Name())

					err = jrpcManager.DownloadFile(ctx, "https://google.com", file.Name())
					assert.NoError(t, err)
					info, err := os.Stat(file.Name())
					require.NoError(t, err)
					assert.NotEqual(t, 0, info.Size())
				},
				"DownloadFileFailsWithBadURL": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					jrpcManager, ok := manager.(*jrpcManager)
					require.True(t, ok)

					err := jrpcManager.DownloadFile(ctx, "", "")
					assert.Error(t, err)
				},
				"DownloadFileFailsForNonexistentURL": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					jrpcManager, ok := manager.(*jrpcManager)
					require.True(t, ok)

					err := jrpcManager.DownloadFile(ctx, "https://google.com/foo", "out.txt")
					assert.Error(t, err)
				},
				"DownloadFileFailsForInvalidPath": func(ctx context.Context, t *testing.T, manager jasper.Manager) {
					jrpcManager, ok := manager.(*jrpcManager)
					require.True(t, ok)

					require.NotEqual(t, 0, os.Geteuid())
					err := jrpcManager.DownloadFile(ctx, "https://google.com", "/foo/bar")
					assert.Error(t, err)
				},
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

type processConstructor func(context.Context, *jasper.CreateOptions) (jasper.Process, error)

func TestJRPCProcess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for cname, makeProc := range map[string]processConstructor{
		"Blocking": func(ctx context.Context, opts *jasper.CreateOptions) (jasper.Process, error) {
			mngr := jasper.NewLocalManagerBlockingProcesses()
			addr, err := startJRPC(ctx, mngr)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			client, err := getClient(ctx, addr)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return client.Create(ctx, opts)

		},
		"NotBlocking": func(ctx context.Context, opts *jasper.CreateOptions) (jasper.Process, error) {
			mngr := jasper.NewLocalManager()
			addr, err := startJRPC(ctx, mngr)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			client, err := getClient(ctx, addr)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return client.Create(ctx, opts)
		},
	} {
		t.Run(cname, func(t *testing.T) {
			for name, testCase := range map[string]func(context.Context, *testing.T, *jasper.CreateOptions, processConstructor){
				"WithPopulatedArgsCommandCreationPasses": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					assert.NotZero(t, opts.Args)
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					assert.NotNil(t, proc)
				},
				"ErrorToCreateWithInvalidArgs": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					opts.Args = []string{}
					proc, err := makep(ctx, opts)
					assert.Error(t, err)
					assert.Nil(t, proc)
				},
				"WithCancledContextProcessCreationFailes": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					pctx, pcancel := context.WithCancel(ctx)
					pcancel()
					proc, err := makep(pctx, opts)
					assert.Error(t, err)
					assert.Nil(t, proc)
				},
				"CanceledContextTimesOutEarly": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					if runtime.GOOS == "windows" {
						t.Skip("the sleep tests don't block correctly on windows")
					}

					pctx, pcancel := context.WithTimeout(ctx, 200*time.Millisecond)
					defer pcancel()
					startAt := time.Now()
					opts.Args = []string{"sleep", "101"}
					proc, err := makep(pctx, opts)
					assert.NoError(t, err)

					time.Sleep(100 * time.Millisecond) // let time pass...
					require.NotNil(t, proc)
					assert.False(t, proc.Info(ctx).Successful)
					assert.True(t, time.Since(startAt) < 400*time.Millisecond)
				},
				"ProcessLacksTagsByDefault": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					tags := proc.GetTags()
					assert.Empty(t, tags)
				},
				"ProcessTagsPersist": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					opts.Tags = []string{"foo"}
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					tags := proc.GetTags()
					assert.Contains(t, tags, "foo")
				},
				"InfoHasMatchingID": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					if assert.NoError(t, err) {
						assert.Equal(t, proc.ID(), proc.Info(ctx).ID)
					}
				},
				"ResetTags": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					proc.Tag("foo")
					assert.Contains(t, proc.GetTags(), "foo")
					proc.ResetTags()
					assert.Len(t, proc.GetTags(), 0)
				},
				"TagsAreSetLike": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)

					for i := 0; i < 100; i++ {
						proc.Tag("foo")
					}

					assert.Len(t, proc.GetTags(), 1)
					proc.Tag("bar")
					assert.Len(t, proc.GetTags(), 2)
				},
				"CompleteIsTrueAfterWait": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					time.Sleep(10 * time.Millisecond) // give the process time to start background machinery
					assert.NoError(t, proc.Wait(ctx))
					assert.True(t, proc.Complete(ctx))
				},
				"WaitReturnsWithCancledContext": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					opts.Args = []string{"sleep", "101"}
					pctx, pcancel := context.WithCancel(ctx)
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					assert.True(t, proc.Running(ctx))
					assert.NoError(t, err)
					pcancel()
					assert.Error(t, proc.Wait(pctx))
				},
				"RegisterTriggerErrorsForNil": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					assert.Error(t, proc.RegisterTrigger(ctx, nil))
				},
				// "": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {},
				// "": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {},

				///////////////////////////////////
				//
				// The following test cases are added
				// specifically for the jrpc case

				"CompleteReturnsFalseForProcessThatDoesntExist": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)

					firstID := proc.ID()
					assert.NoError(t, proc.Wait(ctx))
					assert.True(t, proc.Complete(ctx))
					proc.(*jrpcProcess).info.Id += "_foo"
					proc.(*jrpcProcess).info.Complete = false
					require.NotEqual(t, firstID, proc.ID())
					assert.False(t, proc.Complete(ctx), proc.ID())
				},

				"RunningReturnsFalseForProcessThatDoesntExist": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)

					firstID := proc.ID()
					assert.NoError(t, proc.Wait(ctx))
					proc.(*jrpcProcess).info.Id += "_foo"
					proc.(*jrpcProcess).info.Complete = false
					require.NotEqual(t, firstID, proc.ID())
					assert.False(t, proc.Running(ctx), proc.ID())
				},

				"CompleteAlwaysReturnsTrueWhenProcessIsComplete": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)

					assert.NoError(t, proc.Wait(ctx))

					assert.True(t, proc.Complete(ctx))

				},
				// "": func(ctx context.Context, t *testing.T, opts *jasper.CreateOptions, makep processConstructor) {},
			} {
				t.Run(name, func(t *testing.T) {
					tctx, cancel := context.WithTimeout(ctx, taskTimeout)
					defer cancel()

					opts := &jasper.CreateOptions{Args: []string{"ls"}}
					testCase(tctx, t, opts, makeProc)
				})
			}
		})
	}
}
