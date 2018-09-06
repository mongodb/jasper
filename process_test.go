package jasper

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type processConstructor func(context.Context, *CreateOptions) (Process, error)

func makeLockingProcess(pmake processConstructor) processConstructor {
	return func(ctx context.Context, opts *CreateOptions) (Process, error) {
		proc, err := pmake(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &localProcess{proc: proc}, nil
	}
}

func TestProcessImplementations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient := &http.Client{}

	for cname, makeProc := range map[string]processConstructor{
		"BlockingNoLock":   newBlockingProcess,
		"BlockingWithLock": makeLockingProcess(newBlockingProcess),
		"BasicNoLock":      newBasicProcess,
		"BasicWithLock":    makeLockingProcess(newBasicProcess),
		"REST": func(ctx context.Context, opts *CreateOptions) (Process, error) {

			srv, port := makeAndStartService(ctx, httpClient)
			if port < 100 || srv == nil {
				return nil, errors.New("fixture creation failure")
			}

			client := &restClient{
				prefix: fmt.Sprintf("http://localhost:%d/jasper/v1", port),
				client: httpClient,
			}

			return client.Create(ctx, opts)
		},
	} {
		t.Run(cname, func(t *testing.T) {
			for name, testCase := range map[string]func(context.Context, *testing.T, *CreateOptions, processConstructor){
				"WithPopulatedArgsCommandCreationPasses": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					assert.NotZero(t, opts.Args)
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					assert.NotNil(t, proc)
				},
				"ErrorToCreateWithInvalidArgs": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					opts.Args = []string{}
					proc, err := makep(ctx, opts)
					assert.Error(t, err)
					assert.Nil(t, proc)
				},
				"WithCancledContextProcessCreationFailes": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					pctx, pcancel := context.WithCancel(ctx)
					pcancel()
					proc, err := makep(pctx, opts)
					assert.Error(t, err)
					assert.Nil(t, proc)
				},
				"CanceledContextTimesOutEarly": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
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
				"ProcessLacksTagsByDefault": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					tags := proc.GetTags()
					assert.Empty(t, tags)
				},
				"ProcessTagsPersist": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					opts.Tags = []string{"foo"}
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					tags := proc.GetTags()
					assert.Contains(t, tags, "foo")
				},
				"InfoHasMatchingID": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					if assert.NoError(t, err) {
						assert.Equal(t, proc.ID(), proc.Info(ctx).ID)
					}
				},
				"ResetTags": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					proc.Tag("foo")
					assert.Contains(t, proc.GetTags(), "foo")
					proc.ResetTags()
					assert.Len(t, proc.GetTags(), 0)
				},
				"TagsAreSetLike": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)

					for i := 0; i < 100; i++ {
						proc.Tag("foo")
					}

					assert.Len(t, proc.GetTags(), 1)
					proc.Tag("bar")
					assert.Len(t, proc.GetTags(), 2)
				},
				"CompleteIsTrueAfterWait": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					time.Sleep(10 * time.Millisecond) // give the process time to start background machinery
					assert.NoError(t, proc.Wait(ctx))
					assert.True(t, proc.Complete(ctx))
				},
				"WaitReturnsWithCancledContext": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					opts.Args = []string{"sleep", "101"}
					pctx, pcancel := context.WithCancel(ctx)
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					assert.True(t, proc.Running(ctx))
					assert.NoError(t, err)
					pcancel()
					assert.Error(t, proc.Wait(pctx))
				},
				"RegisterTriggerErrorsForNil": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					proc, err := makep(ctx, opts)
					require.NoError(t, err)
					assert.Error(t, proc.RegisterTrigger(ctx, nil))
				},
				"DefaultTriggerSucceeds": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {
					if cname == "REST" {
						t.Skip("remote triggers are not supported on rest processes")
					}
					proc, err := makep(ctx, opts)
					assert.NoError(t, err)
					assert.NoError(t, proc.RegisterTrigger(ctx, makeDefaultTrigger(ctx, nil, opts, "foo")))
				},
				// "": func(ctx context.Context, t *testing.T, opts *CreateOptions, makep processConstructor) {},
			} {
				t.Run(name, func(t *testing.T) {
					tctx, cancel := context.WithTimeout(ctx, taskTimeout)
					defer cancel()

					opts := &CreateOptions{Args: []string{"ls"}}
					testCase(tctx, t, opts, makeProc)
				})
			}
		})
	}
}

func setupMongods(numProcs int) ([]CreateOptions, []string, error) {
	dbPaths := make([]string, numProcs)
	optslist := make([]CreateOptions, numProcs)
	for i := 0; i < numProcs; i++ {
		procName := "mongod"
		port := getPortNumber()

		dbPath, err := ioutil.TempDir("", procName)
		if err != nil {
			return nil, nil, err
		}
		dbPaths[i] = dbPath

		opts := CreateOptions{Args: []string{procName, "--port", fmt.Sprintf("%d", port), "--dbpath", dbPath}}
		optslist[i] = opts
	}

	return optslist, dbPaths, nil
}

func teardownMongods(dbPaths []string) {
	for _, dbPath := range dbPaths {
		os.RemoveAll(dbPath)
	}
}

func TestMongod(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for name, makeProc := range map[string]processConstructor{
		"Blocking": newBlockingProcess,
	} {
		t.Run(name, func(t *testing.T) {
			for _, test := range []struct {
				id          string
				numProcs    int
				signal      syscall.Signal
				sleepMillis time.Duration
				expectError bool
				errorString string
			}{
				{
					id:          "WithSingleMongod",
					numProcs:    1,
					signal:      syscall.SIGKILL,
					sleepMillis: 0,
					expectError: true,
					errorString: "operation failed",
				},
				{
					id:          "With50MongodsAndSigkill",
					numProcs:    50,
					signal:      syscall.SIGKILL,
					sleepMillis: 0,
					expectError: true,
					errorString: "operation failed",
				},
				{
					id:          "With100MongodsAndSigkill",
					numProcs:    100,
					signal:      syscall.SIGKILL,
					sleepMillis: 0,
					expectError: true,
					errorString: "operation failed",
				},
				{
					id:          "With50MongodsAndSigterm",
					numProcs:    50,
					signal:      syscall.SIGTERM,
					sleepMillis: 3000,
					expectError: false,
					errorString: "",
				},
			} {
				t.Run(test.id, func(t *testing.T) {
					optslist, dbPaths, err := setupMongods(test.numProcs)
					defer teardownMongods(dbPaths)
					require.NoError(t, err)

					// Spawn concurrent mongods
					procs := make([]Process, test.numProcs)
					for i, opts := range optslist {
						proc, err := makeProc(ctx, &opts)
						require.NoError(t, err)
						assert.True(t, proc.Running(ctx))
						procs[i] = proc
					}

					waitError := make(chan error, test.numProcs)
					for _, proc := range procs {
						go func(proc Process) {
							err := proc.Wait(ctx)
							waitError <- err
						}(proc)
					}

					// Signal to stop mongods
					time.Sleep(test.sleepMillis * time.Millisecond)
					for _, proc := range procs {
						err := proc.Signal(ctx, test.signal)
						assert.NoError(t, err)
					}

					// Check that the processes all received signal
					for i := 0; i < test.numProcs; i++ {
						err := <-waitError
						if test.expectError {
							assert.EqualError(t, err, test.errorString)
						} else {
							assert.NoError(t, err)
						}
					}
				})
			}
		})
	}
}
