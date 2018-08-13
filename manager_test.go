package jasper

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mongodb/grip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func trueCreateOpts() *CreateOptions {
	return &CreateOptions{
		Args: []string{"true"},
	}
}

func falseCreateOpts() *CreateOptions {
	return &CreateOptions{
		Args: []string{"false"},
	}
}

func sleepCreateOpts(num int) *CreateOptions {
	return &CreateOptions{
		Args: []string{"sleep", fmt.Sprint(num)},
	}
}

func createProcs(ctx context.Context, opts *CreateOptions, manager Manager, num int) ([]Process, error) {
	catcher := grip.NewBasicCatcher()
	out := []Process{}
	for i := 0; i < num; i++ {
		proc, err := manager.Create(ctx, opts)
		catcher.Add(err)
		if proc != nil {
			out = append(out, proc)
		}
	}

	return out, catcher.Resolve()
}

func TestManagerInterface(t *testing.T) {
	httpClient := &http.Client{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srvPort := 3000

	for mname, factory := range map[string]func() Manager{
		"Basic/NoLock": func() Manager {
			return &basicProcessManager{
				procs: map[string]Process{},
			}
		},
		"Basic/Lock": func() Manager {
			return &localProcessManager{
				manager: &basicProcessManager{
					procs: map[string]Process{},
				},
			}
		},
		"Basic/NoLock/BlockingProcs": func() Manager {
			return &basicProcessManager{
				procs:    map[string]Process{},
				blocking: true,
			}
		},
		"Basic/Lock/BlockingProcs": func() Manager {
			return &localProcessManager{
				manager: &basicProcessManager{
					procs:    map[string]Process{},
					blocking: true,
				},
			}
		},
		"REST": func() Manager {
			srv := NewManagerService(NewLocalManager()).App()
			srvPort++
			srv.SetPrefix("jasper")

			require.NoError(t, srv.SetPort(srvPort))
			go func() {
				srv.Run(ctx)
			}()

			time.Sleep(100 * time.Millisecond)
			return &restClient{
				prefix: fmt.Sprintf("http://localhost:%d/jasper/v1", srvPort),
				client: httpClient,
			}
		},
	} {
		t.Run(mname, func(t *testing.T) {
			for name, test := range map[string]func(context.Context, *testing.T, Manager){
				"ValidateFixture": func(ctx context.Context, t *testing.T, manager Manager) {
					assert.NotNil(t, ctx)
					assert.NotNil(t, manager)
				},
				"ListErrorsWhenEmpty": func(ctx context.Context, t *testing.T, manager Manager) {
					all, err := manager.List(ctx, All)
					assert.Error(t, err)
					assert.Len(t, all, 0)
					assert.Contains(t, err.Error(), "no processes")
				},
				"CreateSimpleProcess": func(ctx context.Context, t *testing.T, manager Manager) {
					if mname == "REST" {
						t.Skip("test case not compatible with rest interfaces")
					}

					opts := trueCreateOpts()
					proc, err := manager.Create(ctx, opts)
					assert.NoError(t, err)
					assert.NotNil(t, proc)
					assert.True(t, opts.started)
				},
				"CreateProcessFails": func(ctx context.Context, t *testing.T, manager Manager) {
					proc, err := manager.Create(ctx, &CreateOptions{})
					assert.Error(t, err)
					assert.Nil(t, proc)
				},
				"ListAllOperations": func(ctx context.Context, t *testing.T, manager Manager) {
					t.Skip("this often deadlocks")
					created, err := createProcs(ctx, trueCreateOpts(), manager, 10)
					assert.NoError(t, err)
					assert.Len(t, created, 10)
					output, err := manager.List(ctx, All)
					assert.NoError(t, err)
					assert.Len(t, output, 10)
				},
				"ListAllReturnsErrorWithCancledContext": func(ctx context.Context, t *testing.T, manager Manager) {
					cctx, cancel := context.WithCancel(ctx)
					created, err := createProcs(ctx, trueCreateOpts(), manager, 10)
					assert.NoError(t, err)
					assert.Len(t, created, 10)
					cancel()
					output, err := manager.List(cctx, All)
					assert.Error(t, err)
					assert.Nil(t, output)
				},
				"LongRunningOperationsAreListedAsRunning": func(ctx context.Context, t *testing.T, manager Manager) {
					procs, err := createProcs(ctx, sleepCreateOpts(1), manager, 10)
					assert.NoError(t, err)
					assert.Len(t, procs, 10)

					procs, err = manager.List(ctx, Running)
					assert.NoError(t, err)
					assert.Len(t, procs, 10)

					procs, err = manager.List(ctx, Successful)
					assert.Error(t, err)
					assert.Len(t, procs, 0)
				},
				"ListReturnsOneSuccessfulCommand": func(ctx context.Context, t *testing.T, manager Manager) {
					proc, err := manager.Create(ctx, trueCreateOpts())
					require.NoError(t, err)

					assert.NoError(t, proc.Wait(ctx))

					listOut, err := manager.List(ctx, Successful)
					assert.NoError(t, err)

					if assert.Len(t, listOut, 1) {
						assert.Equal(t, listOut[0].ID(), proc.ID())
					}
				},
				"ListReturnsOneFailedCommand": func(ctx context.Context, t *testing.T, manager Manager) {
					proc, err := manager.Create(ctx, falseCreateOpts())
					require.NoError(t, err)
					assert.Error(t, proc.Wait(ctx))

					listOut, err := manager.List(ctx, Failed)
					assert.NoError(t, err)

					if assert.Len(t, listOut, 1) {
						assert.Equal(t, listOut[0].ID(), proc.ID())
					}
				},
				"GetMethodErrorsWithNoResponse": func(ctx context.Context, t *testing.T, manager Manager) {
					proc, err := manager.Get(ctx, "foo")
					assert.Error(t, err)
					assert.Nil(t, proc)
				},
				"GetMethodReturnsMatchingDoc": func(ctx context.Context, t *testing.T, manager Manager) {
					proc, err := manager.Create(ctx, trueCreateOpts())
					require.NoError(t, err)

					ret, err := manager.Get(ctx, proc.ID())
					assert.NoError(t, err)
					assert.Equal(t, ret.ID(), proc.ID())
				},
				"GroupErrorsWithoutResults": func(ctx context.Context, t *testing.T, manager Manager) {
					procs, err := manager.Group(ctx, "foo")
					assert.Error(t, err)
					assert.Len(t, procs, 0)
					assert.Contains(t, err.Error(), "no jobs")
				},
				"GroupErrorsForCanceledContexts": func(ctx context.Context, t *testing.T, manager Manager) {
					_, err := manager.Create(ctx, trueCreateOpts())
					assert.NoError(t, err)

					cctx, cancel := context.WithCancel(ctx)
					cancel()
					procs, err := manager.Group(cctx, "foo")
					assert.Error(t, err)
					assert.Len(t, procs, 0)
					assert.Contains(t, err.Error(), "canceled")
				},
				"GroupPropgatesMatching": func(ctx context.Context, t *testing.T, manager Manager) {
					proc, err := manager.Create(ctx, trueCreateOpts())
					require.NoError(t, err)

					proc.Tag("foo")

					procs, err := manager.Group(ctx, "foo")
					require.NoError(t, err)
					require.Len(t, procs, 1)
					assert.Equal(t, procs[0].ID(), proc.ID())
				},
				"CloseEmptyManagerNoops": func(ctx context.Context, t *testing.T, manager Manager) {
					assert.NoError(t, manager.Close(ctx))
				},
				"ClosersWithoutTriggersTerminatesProcesses": func(ctx context.Context, t *testing.T, manager Manager) {
					_, err := createProcs(ctx, sleepCreateOpts(100), manager, 10)
					assert.NoError(t, err)
					assert.NoError(t, manager.Close(ctx))
				},
				"CloseErrorsWithCanceledContext": func(ctx context.Context, t *testing.T, manager Manager) {
					_, err := createProcs(ctx, sleepCreateOpts(100), manager, 10)
					assert.NoError(t, err)

					cctx, cancel := context.WithCancel(ctx)
					cancel()

					err = manager.Close(cctx)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "canceled")
				},
				"CloseErrorsWithTerminatedProcesses": func(ctx context.Context, t *testing.T, manager Manager) {
					procs, err := createProcs(ctx, trueCreateOpts(), manager, 10)
					for _, p := range procs {
						assert.NoError(t, p.Wait(ctx))
					}

					assert.NoError(t, err)
					assert.Error(t, manager.Close(ctx))
				},
				"CloseExcutesClosersForProcesses": func(ctx context.Context, t *testing.T, manager Manager) {
					if mname == "REST" {
						t.Skip("not supported on rest interfaces")
					}

					opts := sleepCreateOpts(100)
					counter := 0
					opts.closers = append(opts.closers, func() {
						counter++
					})
					_, err := manager.Create(ctx, opts)
					assert.NoError(t, err)

					assert.Equal(t, counter, 0)
					assert.NoError(t, manager.Close(ctx))
					assert.Equal(t, 1, counter)
				},
				"RegisterProcessErrorsForNilProcess": func(ctx context.Context, t *testing.T, manager Manager) {
					if mname == "REST" {
						t.Skip("not supported on rest interfaces")
					}

					err := manager.Register(ctx, nil)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "not defined")
				},
				"RegisterProcessErrorsForCancledContext": func(ctx context.Context, t *testing.T, manager Manager) {
					if mname == "REST" {
						t.Skip("not supported on rest interfaces")
					}

					cctx, cancel := context.WithCancel(ctx)
					cancel()
					proc, err := newBasicProcess(ctx, trueCreateOpts())
					assert.NoError(t, err)
					err = manager.Register(cctx, proc)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "canceled")
				},
				"RegisterProcessErrorsWhenMissingID": func(ctx context.Context, t *testing.T, manager Manager) {
					if mname == "REST" {
						t.Skip("not supported on rest interfaces")
					}

					proc := &basicProcess{}
					assert.Equal(t, proc.ID(), "")
					err := manager.Register(ctx, proc)
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "malformed")
				},
				"RegisterProcessModifiesManagerState": func(ctx context.Context, t *testing.T, manager Manager) {
					if mname == "REST" {
						t.Skip("not supported on rest interfaces")
					}

					proc, err := newBasicProcess(ctx, trueCreateOpts())
					require.NoError(t, err)
					err = manager.Register(ctx, proc)
					assert.NoError(t, err)

					procs, err := manager.List(ctx, All)
					assert.NoError(t, err)
					require.True(t, len(procs) >= 1)

					assert.Equal(t, procs[0].ID(), proc.ID())
				},
				"RegisterProcessErrorsForDuplicateProcess": func(ctx context.Context, t *testing.T, manager Manager) {
					if mname == "REST" {
						t.Skip("not supported on rest interfaces")
					}

					proc, err := newBasicProcess(ctx, trueCreateOpts())
					assert.NoError(t, err)
					assert.NotEmpty(t, proc)
					err = manager.Register(ctx, proc)
					assert.NoError(t, err)
					err = manager.Register(ctx, proc)
					assert.Error(t, err)
				},
				// "": func(ctx context.Context, t *testing.T, manager Manager) {},
				// "": func(ctx context.Context, t *testing.T, manager Manager) {},
			} {
				t.Run(name, func(t *testing.T) {
					tctx, cancel := context.WithCancel(ctx)
					defer cancel()
					test(tctx, t, factory())
				})
			}
		})
	}
}
