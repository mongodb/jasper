package remote

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

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
	grip.Error(sender.SetLevel(send.LevelInfo{Default: level.Info, Threshold: level.Info}))
	grip.Error(grip.SetSender(sender))
}

// addBasicClientTests contains all the manager tests found in the root package
// TestManagerImplementations, minus the ones that are not compatible with
// remote interfaces. Other than incompatible tests, these tests should exactly
// mirror the ones in the root package.
// func addBasicClientTests(modify testutil.OptsModify, tests ...clientTestCase) []jasper.ManagerTestCase {
//     return append([]jasper.ManagerTestCase{
//         // {
//         //     Name: "ValidateFixture",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         assert.NotNil(t, ctx)
//         //         assert.NotNil(t, client)
//         //     },
//         // },
//         // {
//         //     Name: "IDReturnsNonempty",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         assert.NotEmpty(t, client.ID())
//         //     },
//         // },
//         // {
//         //     Name: "ProcEnvVarMatchesManagerID",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.TrueCreateOpts()
//         //         modify(opts)
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.NoError(t, err)
//         //         info := proc.Info(ctx)
//         //         require.NotEmpty(t, info.Options.Environment)
//         //         assert.Equal(t, client.ID(), info.Options.Environment[jasper.ManagerEnvironID])
//         //     },
//         // },
//         // {
//         //     Name: "CreateProcessFailsWithEmptyOptions",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := &options.Create{}
//         //         modify(opts)
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.Error(t, err)
//         //         assert.Nil(t, proc)
//         //     },
//         // },
//         // {
//         //     Name: "CreateSimpleProcess",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.TrueCreateOpts()
//         //         modify(opts)
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.NoError(t, err)
//         //         assert.NotNil(t, proc)
//         //         info := proc.Info(ctx)
//         //         assert.True(t, info.IsRunning || info.Complete)
//         //     },
//         // },
//         // {
//         //     Name: "CreateProcessFails",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := &options.Create{}
//         //         modify(opts)
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.Error(t, err)
//         //         assert.Nil(t, proc)
//         //     },
//         // },
//         // {
//         //     Name: "ListDoesNotErrorWhenEmpty",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         all, err := client.List(ctx, options.All)
//         //         require.NoError(t, err)
//         //         assert.Len(t, all, 0)
//         //     },
//         // },
//         // {
//         //     Name: "ListAllOperations",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.TrueCreateOpts()
//         //         modify(opts)
//         //         created, err := createProcs(ctx, opts, client, 10)
//         //         require.NoError(t, err)
//         //         assert.Len(t, created, 10)
//         //         output, err := client.List(ctx, options.All)
//         //         require.NoError(t, err)
//         //         assert.Len(t, output, 10)
//         //     },
//         // },
//         // {
//         //     Name: "ListAllReturnsErrorWithCanceledContext",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         cctx, cancel := context.WithCancel(ctx)
//         //         opts := testutil.TrueCreateOpts()
//         //         modify(opts)
//         //
//         //         created, err := createProcs(ctx, opts, client, 10)
//         //         require.NoError(t, err)
//         //         assert.Len(t, created, 10)
//         //         cancel()
//         //         output, err := client.List(cctx, options.All)
//         //         require.Error(t, err)
//         //         assert.Nil(t, output)
//         //     },
//         // },
//         // {
//         //     Name: "LongRunningOperationsAreListedAsRunning",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.SleepCreateOpts(20)
//         //         modify(opts)
//         //         procs, err := createProcs(ctx, opts, client, 10)
//         //         require.NoError(t, err)
//         //         assert.Len(t, procs, 10)
//         //
//         //         procs, err = client.List(ctx, options.All)
//         //         require.NoError(t, err)
//         //         assert.Len(t, procs, 10)
//         //
//         //         procs, err = client.List(ctx, options.Running)
//         //         require.NoError(t, err)
//         //         assert.Len(t, procs, 10)
//         //
//         //         procs, err = client.List(ctx, options.Successful)
//         //         require.NoError(t, err)
//         //         assert.Len(t, procs, 0)
//         //     },
//         // },
//         // {
//         //     Name: "ListReturnsOneSuccessfulCommand",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.TrueCreateOpts()
//         //         modify(opts)
//         //
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.NoError(t, err)
//         //
//         //         _, err = proc.Wait(ctx)
//         //         require.NoError(t, err)
//         //
//         //         listOut, err := client.List(ctx, options.Successful)
//         //         require.NoError(t, err)
//         //
//         //         require.Len(t, listOut, 1)
//         //         assert.Equal(t, listOut[0].ID(), proc.ID())
//         //     },
//         // },
//         // {
//         //     Name: "ListReturnsOneFailedCommand",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.FalseCreateOpts()
//         //         modify(opts)
//         //
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.NoError(t, err)
//         //         _, err = proc.Wait(ctx)
//         //         require.Error(t, err)
//         //
//         //         listOut, err := client.List(ctx, options.Failed)
//         //         require.NoError(t, err)
//         //
//         //         require.Len(t, listOut, 1)
//         //         assert.Equal(t, listOut[0].ID(), proc.ID())
//         //     },
//         // },
//         // {
//         //     Name: "ListErrorsWithInvalidFilter",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         procs, err := client.List(ctx, options.Filter("foo"))
//         //         assert.Error(t, err)
//         //         assert.Empty(t, procs)
//         //     },
//         // },
//         // {
//         //     Name: "GetMethodErrorsWithNonexistentProcess",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         proc, err := client.Get(ctx, "foo")
//         //         require.Error(t, err)
//         //         assert.Nil(t, proc)
//         //     },
//         // },
//         // {
//         //     Name: "GetMethodReturnsMatchingProcess",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.TrueCreateOpts()
//         //         modify(opts)
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.NoError(t, err)
//         //
//         //         ret, err := client.Get(ctx, proc.ID())
//         //         require.NoError(t, err)
//         //         assert.Equal(t, ret.ID(), proc.ID())
//         //     },
//         // },
//         // {
//         //     Name: "GroupDoesNotErrorWhenEmptyResult",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         procs, err := client.Group(ctx, "foo")
//         //         require.NoError(t, err)
//         //         assert.Len(t, procs, 0)
//         //     },
//         // },
//         // {
//         //     Name: "GroupErrorsForCanceledContext",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.TrueCreateOpts()
//         //         modify(opts)
//         //         _, err := client.CreateProcess(ctx, opts)
//         //         require.NoError(t, err)
//         //
//         //         cctx, cancel := context.WithCancel(ctx)
//         //         cancel()
//         //         procs, err := client.Group(cctx, "foo")
//         //         require.Error(t, err)
//         //         assert.Len(t, procs, 0)
//         //         assert.Contains(t, err.Error(), "canceled")
//         //     },
//         // },
//         // {
//         //     Name: "GroupPropagatesMatching",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.TrueCreateOpts()
//         //         modify(opts)
//         //
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.NoError(t, err)
//         //
//         //         proc.Tag("foo")
//         //
//         //         procs, err := client.Group(ctx, "foo")
//         //         require.NoError(t, err)
//         //         require.Len(t, procs, 1)
//         //         assert.Equal(t, procs[0].ID(), proc.ID())
//         //     },
//         // },
//         // {
//         //     Name: "CloseEmptyManagerNoops",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         assert.NoError(t, client.Close(ctx))
//         //     },
//         // },
//         // {
//         //     Name: "CloseErrorsWithCanceledContext",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.SleepCreateOpts(100)
//         //         modify(opts)
//         //
//         //         _, err := createProcs(ctx, opts, client, 10)
//         //         require.NoError(t, err)
//         //
//         //         cctx, cancel := context.WithCancel(ctx)
//         //         cancel()
//         //
//         //         err = client.Close(cctx)
//         //         require.Error(t, err)
//         //         assert.Contains(t, err.Error(), "canceled")
//         //     },
//         // },
//         // {
//         //     Name: "CloseSucceedsWithTerminatedProcesses",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         procs, err := createProcs(ctx, testutil.TrueCreateOpts(), client, 10)
//         //         for _, p := range procs {
//         //             _, err = p.Wait(ctx)
//         //             require.NoError(t, err)
//         //         }
//         //
//         //         require.NoError(t, err)
//         //         assert.NoError(t, client.Close(ctx))
//         //     },
//         // },
//         // {
//         //     Name: "CloserWithoutTriggersTerminatesProcesses",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         if runtime.GOOS == "windows" {
//         //             t.Skip("manager close tests will error due to process termination on Windows")
//         //         }
//         //         opts := testutil.SleepCreateOpts(100)
//         //         modify(opts)
//         //
//         //         _, err := createProcs(ctx, opts, client, 10)
//         //         require.NoError(t, err)
//         //         assert.NoError(t, client.Close(ctx))
//         //     },
//         // },
//         // {
//         //     Name: "ClearCausesDeletionOfProcesses",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.TrueCreateOpts()
//         //         modify(opts)
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.NoError(t, err)
//         //         sameProc, err := client.Get(ctx, proc.ID())
//         //         require.NoError(t, err)
//         //         require.Equal(t, proc.ID(), sameProc.ID())
//         //         _, err = proc.Wait(ctx)
//         //         require.NoError(t, err)
//         //         client.Clear(ctx)
//         //         nilProc, err := client.Get(ctx, proc.ID())
//         //         require.Error(t, err)
//         //         assert.Nil(t, nilProc)
//         //     },
//         // },
//         // {
//         //     Name: "ClearIsANoopForActiveProcesses",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         opts := testutil.SleepCreateOpts(20)
//         //         modify(opts)
//         //         proc, err := client.CreateProcess(ctx, opts)
//         //         require.NoError(t, err)
//         //         client.Clear(ctx)
//         //         sameProc, err := client.Get(ctx, proc.ID())
//         //         require.NoError(t, err)
//         //         assert.Equal(t, proc.ID(), sameProc.ID())
//         //         require.NoError(t, jasper.Terminate(ctx, proc)) // Clean up
//         //     },
//         // },
//         // {
//         //     Name: "ClearSelectivelyDeletesOnlyDeadProcesses",
//         //     Case: func(ctx context.Context, t *testing.T, mngr Manager, modifyOpts testutil.OptsModify){
//         //         trueOpts := testutil.TrueCreateOpts()
//         //         modify(trueOpts)
//         //         lsProc, err := client.CreateProcess(ctx, trueOpts)
//         //         require.NoError(t, err)
//         //
//         //         sleepOpts := testutil.SleepCreateOpts(20)
//         //         modify(sleepOpts)
//         //         sleepProc, err := client.CreateProcess(ctx, sleepOpts)
//         //         require.NoError(t, err)
//         //
//         //         _, err = lsProc.Wait(ctx)
//         //         require.NoError(t, err)
//         //
//         //         client.Clear(ctx)
//         //
//         //         sameSleepProc, err := client.Get(ctx, sleepProc.ID())
//         //         require.NoError(t, err)
//         //         assert.Equal(t, sleepProc.ID(), sameSleepProc.ID())
//         //
//         //         nilProc, err := client.Get(ctx, lsProc.ID())
//         //         require.Error(t, err)
//         //         assert.Nil(t, nilProc)
//         //         require.NoError(t, jasper.Terminate(ctx, sleepProc)) // Clean up
//         //     },
//         // },
//         //
//         // The tests below this are specific to the remote manager.
//         //
//         {
//             Name: "WaitingOnNonexistentProcessErrors",
//             Case: func(ctx context.Context, t *testing.T, mngr jasper.Manager, modifyOpts testutil.OptsModify) {
//                 opts := modifyOpts(testutil.TrueCreateOpts())
//
//                 proc, err := mngr.CreateProcess(ctx, opts)
//                 require.NoError(t, err)
//
//                 _, err = proc.Wait(ctx)
//                 require.NoError(t, err)
//
//                 mngr.Clear(ctx)
//
//                 _, err = proc.Wait(ctx)
//                 require.Error(t, err)
//                 procs, err := mngr.List(ctx, options.All)
//                 require.NoError(t, err)
//                 assert.Len(t, procs, 0)
//             },
//         },
//         {
//             Name: "RegisterAlwaysErrors",
//             Case: func(ctx context.Context, t *testing.T, mngr jasper.Manager, modifyOpts testutil.OptsModify) {
//                 proc, err := mngr.CreateProcess(ctx, &options.Create{Args: []string{"ls"}})
//                 assert.NotNil(t, proc)
//                 require.NoError(t, err)
//
//                 assert.Error(t, mngr.Register(ctx, nil))
//                 assert.Error(t, mngr.Register(ctx, proc))
//             },
//         },
//     }, tests...)
// }

func remoteManagerTestCases(httpClient *http.Client) map[string]func(context.Context, *testing.T) Manager {
	return map[string]func(context.Context, *testing.T) Manager{
		"MDB": func(ctx context.Context, t *testing.T) Manager {
			mngr, err := jasper.NewSynchronizedManager(false)
			require.NoError(t, err)

			client, err := makeTestMDBServiceAndClient(ctx, mngr)
			require.NoError(t, err)
			return client
		},
		"RPC/TLS": func(ctx context.Context, t *testing.T) Manager {
			mngr, err := jasper.NewSynchronizedManager(false)
			require.NoError(t, err)

			client, err := makeTLSRPCServiceAndClient(ctx, mngr)
			require.NoError(t, err)
			return client
		},
		"RPC/Insecure": func(ctx context.Context, t *testing.T) Manager {
			assert.NotPanics(t, func() {
				newRPCClient(nil)
			})

			mngr, err := jasper.NewSynchronizedManager(false)
			require.NoError(t, err)

			client, err := makeInsecureRPCServiceAndClient(ctx, mngr)
			require.NoError(t, err)
			return client
		},
		"REST": func(ctx context.Context, t *testing.T) Manager {
			_, port, err := startRESTService(ctx, httpClient)
			require.NoError(t, err)

			client := &restClient{
				prefix: fmt.Sprintf("http://localhost:%d/jasper/v1", port),
				client: httpClient,
			}
			return client
		},
	}
}

func TestManagerImplementations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient := testutil.GetHTTPClient()
	defer testutil.PutHTTPClient(httpClient)

	testCases := jasper.ManagerTests()
	for _, testCase := range []jasper.ManagerTestCase{
		{
			Name: "WaitingOnNonexistentProcessErrors",
			Case: func(ctx context.Context, t *testing.T, mngr jasper.Manager, modifyOpts testutil.OptsModify) {
				opts := modifyOpts(testutil.TrueCreateOpts())

				proc, err := mngr.CreateProcess(ctx, opts)
				require.NoError(t, err)

				_, err = proc.Wait(ctx)
				require.NoError(t, err)

				mngr.Clear(ctx)

				_, err = proc.Wait(ctx)
				require.Error(t, err)
				procs, err := mngr.List(ctx, options.All)
				require.NoError(t, err)
				assert.Len(t, procs, 0)
			},
		},
		{
			Name: "RegisterAlwaysErrors",
			Case: func(ctx context.Context, t *testing.T, mngr jasper.Manager, modifyOpts testutil.OptsModify) {
				proc, err := mngr.CreateProcess(ctx, &options.Create{Args: []string{"ls"}})
				assert.NotNil(t, proc)
				require.NoError(t, err)

				assert.Error(t, mngr.Register(ctx, nil))
				assert.Error(t, mngr.Register(ctx, proc))
			},
		},
	} {
		testCases = append(testCases, testCase)
	}

	for managerName, makeManager := range remoteManagerTestCases(httpClient) {
		t.Run(managerName, func(t *testing.T) {
			for _, testCase := range testCases {
				t.Run(testCase.Name, func(t *testing.T) {
					for optsTestCase, modifyOpts := range map[string]testutil.OptsModify{
						"BlockingProcess": func(opts *options.Create) *options.Create {
							opts.Implementation = options.ProcessImplementationBlocking
							return opts
						},
						"BasicProcess": func(opts *options.Create) *options.Create {
							opts.Implementation = options.ProcessImplementationBasic
							return opts
						},
					} {
						t.Run(optsTestCase, func(t *testing.T) {
							tctx, tcancel := context.WithTimeout(ctx, testutil.RPCTestTimeout)
							defer tcancel()
							mngr := makeManager(tctx, t)
							defer func() {
								assert.NoError(t, mngr.CloseConnection())
							}()
							testCase.Case(tctx, t, mngr, modifyOpts)
						})
					}
				})
			}
		})
	}
}

type clientTestCase struct {
	Name string
	Case func(context.Context, *testing.T, Manager)
}

func TestClientImplementations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient := testutil.GetHTTPClient()
	defer testutil.PutHTTPClient(httpClient)

	for managerName, makeManager := range remoteManagerTestCases(httpClient) {
		t.Run(managerName, func(t *testing.T) {
			for _, testCase := range []clientTestCase{
				{
					Name: "StandardInput",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						for subTestName, subTestCase := range map[string]func(ctx context.Context, t *testing.T, opts *options.Create, expectedOutput string, stdin []byte){
							"ReaderIsIgnored": func(ctx context.Context, t *testing.T, opts *options.Create, expectedOutput string, stdin []byte) {
								opts.StandardInput = bytes.NewBuffer(stdin)

								proc, err := mngr.CreateProcess(ctx, opts)
								require.NoError(t, err)

								_, err = proc.Wait(ctx)
								require.NoError(t, err)

								logs, err := mngr.GetLogStream(ctx, proc.ID(), 1)
								require.NoError(t, err)
								assert.Empty(t, logs.Logs)
							},
							"BytesSetsStandardInput": func(ctx context.Context, t *testing.T, opts *options.Create, expectedOutput string, stdin []byte) {
								opts.StandardInputBytes = stdin

								proc, err := mngr.CreateProcess(ctx, opts)
								require.NoError(t, err)

								_, err = proc.Wait(ctx)
								require.NoError(t, err)

								logs, err := mngr.GetLogStream(ctx, proc.ID(), 1)
								require.NoError(t, err)

								require.Len(t, logs.Logs, 1)
								assert.Equal(t, expectedOutput, strings.TrimSpace(logs.Logs[0]))
							},
							"BytesCopiedByRespawnedProcess": func(ctx context.Context, t *testing.T, opts *options.Create, expectedOutput string, stdin []byte) {
								opts.StandardInputBytes = stdin

								proc, err := mngr.CreateProcess(ctx, opts)
								require.NoError(t, err)

								_, err = proc.Wait(ctx)
								require.NoError(t, err)

								logs, err := mngr.GetLogStream(ctx, proc.ID(), 1)
								require.NoError(t, err)

								require.Len(t, logs.Logs, 1)
								assert.Equal(t, expectedOutput, strings.TrimSpace(logs.Logs[0]))

								newProc, err := proc.Respawn(ctx)
								require.NoError(t, err)

								_, err = newProc.Wait(ctx)
								require.NoError(t, err)

								logs, err = mngr.GetLogStream(ctx, newProc.ID(), 1)
								require.NoError(t, err)

								require.Len(t, logs.Logs, 1)
								assert.Equal(t, expectedOutput, strings.TrimSpace(logs.Logs[0]))
							},
						} {
							t.Run(subTestName, func(t *testing.T) {
								inMemLogger, err := jasper.NewInMemoryLogger(1)
								require.NoError(t, err)

								opts := &options.Create{
									Args: []string{"bash", "-s"},
									Output: options.Output{
										Loggers: []*options.LoggerConfig{inMemLogger},
									},
								}

								expectedOutput := "foobar"
								stdin := []byte("echo " + expectedOutput)
								subTestCase(ctx, t, opts, expectedOutput, stdin)
							})
						}
					},
				},
				{
					Name: "GetLogStreamFromNonexistentProcessFails",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						stream, err := mngr.GetLogStream(ctx, "foo", 1)
						assert.Error(t, err)
						assert.Zero(t, stream)
					},
				},
				{
					Name: "GetLogStreamFailsWithoutInMemoryLogger",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						opts := &options.Create{Args: []string{"echo", "foo"}}
						proc, err := mngr.CreateProcess(ctx, opts)
						require.NoError(t, err)
						require.NotNil(t, proc)

						_, err = proc.Wait(ctx)
						require.NoError(t, err)

						stream, err := mngr.GetLogStream(ctx, proc.ID(), 1)
						assert.Error(t, err)
						assert.Zero(t, stream)
					},
				},
				{
					Name: "WithInMemoryLogger",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						inMemLogger, err := jasper.NewInMemoryLogger(100)
						require.NoError(t, err)
						output := "foo"
						opts := &options.Create{
							Args: []string{"echo", output},
							Output: options.Output{
								Loggers: []*options.LoggerConfig{inMemLogger},
							},
						}

						for testName, testCase := range map[string]func(ctx context.Context, t *testing.T, proc jasper.Process){
							"GetLogStreamFailsForInvalidCount": func(ctx context.Context, t *testing.T, proc jasper.Process) {
								stream, err := mngr.GetLogStream(ctx, proc.ID(), -1)
								assert.Error(t, err)
								assert.Zero(t, stream)
							},
							"GetLogStreamReturnsOutputOnSuccess": func(ctx context.Context, t *testing.T, proc jasper.Process) {
								logs := []string{}
								for stream, err := mngr.GetLogStream(ctx, proc.ID(), 1); !stream.Done; stream, err = mngr.GetLogStream(ctx, proc.ID(), 1) {
									require.NoError(t, err)
									require.NotEmpty(t, stream.Logs)
									logs = append(logs, stream.Logs...)
								}
								assert.Contains(t, logs, output)
							},
						} {
							t.Run(testName, func(t *testing.T) {
								proc, err := mngr.CreateProcess(ctx, opts)
								require.NoError(t, err)
								require.NotNil(t, proc)

								_, err = proc.Wait(ctx)
								require.NoError(t, err)
								testCase(ctx, t, proc)
							})
						}
					},
				},
				{
					Name: "DownloadFile",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						for testName, testCase := range map[string]func(ctx context.Context, t *testing.T, mngr Manager, tempDir string){
							"CreatesFileIfNonexistent": func(ctx context.Context, t *testing.T, mngr Manager, tempDir string) {
								opts := options.Download{
									URL:  "https://example.com",
									Path: filepath.Join(tempDir, filepath.Base(t.Name())),
								}
								require.NoError(t, mngr.DownloadFile(ctx, opts))
								defer func() {
									assert.NoError(t, os.RemoveAll(opts.Path))
								}()

								fileInfo, err := os.Stat(opts.Path)
								require.NoError(t, err)
								assert.NotZero(t, fileInfo.Size())
							},
							"WritesFileIfExists": func(ctx context.Context, t *testing.T, mngr Manager, tempDir string) {
								file, err := ioutil.TempFile(tempDir, "out.txt")
								require.NoError(t, err)
								defer func() {
									assert.NoError(t, os.RemoveAll(file.Name()))
								}()
								require.NoError(t, file.Close())

								opts := options.Download{
									URL:  "https://example.com",
									Path: file.Name(),
								}
								require.NoError(t, mngr.DownloadFile(ctx, opts))
								defer func() {
									assert.NoError(t, os.RemoveAll(opts.Path))
								}()

								fileInfo, err := os.Stat(file.Name())
								require.NoError(t, err)
								assert.NotZero(t, fileInfo.Size())
							},
							"CreatesFileAndExtracts": func(ctx context.Context, t *testing.T, mngr Manager, tempDir string) {
								downloadDir, err := ioutil.TempDir(tempDir, "out")
								require.NoError(t, err)
								defer func() {
									assert.NoError(t, os.RemoveAll(downloadDir))
								}()

								fileServerDir, err := ioutil.TempDir(tempDir, "file_server")
								require.NoError(t, err)
								defer func() {
									assert.NoError(t, os.RemoveAll(fileServerDir))
								}()

								fileName := "foo.zip"
								fileContents := "foo"
								require.NoError(t, testutil.AddFileToDirectory(fileServerDir, fileName, fileContents))

								absDownloadDir, err := filepath.Abs(downloadDir)
								require.NoError(t, err)
								destFilePath := filepath.Join(absDownloadDir, fileName)
								destExtractDir := filepath.Join(absDownloadDir, "extracted")

								port := testutil.GetPortNumber()
								fileServerAddr := fmt.Sprintf("localhost:%d", port)
								fileServer := &http.Server{Addr: fileServerAddr, Handler: http.FileServer(http.Dir(fileServerDir))}
								defer func() {
									assert.NoError(t, fileServer.Close())
								}()
								listener, err := net.Listen("tcp", fileServerAddr)
								require.NoError(t, err)
								go func() {
									grip.Info(fileServer.Serve(listener))
								}()

								baseURL := fmt.Sprintf("http://%s", fileServerAddr)
								require.NoError(t, testutil.WaitForRESTService(ctx, baseURL))

								opts := options.Download{
									URL:  fmt.Sprintf("%s/%s", baseURL, fileName),
									Path: destFilePath,
									ArchiveOpts: options.Archive{
										ShouldExtract: true,
										Format:        options.ArchiveZip,
										TargetPath:    destExtractDir,
									},
								}
								require.NoError(t, mngr.DownloadFile(ctx, opts))

								fileInfo, err := os.Stat(destFilePath)
								require.NoError(t, err)
								assert.NotZero(t, fileInfo.Size())

								dirContents, err := ioutil.ReadDir(destExtractDir)
								require.NoError(t, err)

								assert.NotZero(t, len(dirContents))
							},
							"FailsForInvalidArchiveFormat": func(ctx context.Context, t *testing.T, mngr Manager, tempDir string) {
								file, err := ioutil.TempFile(tempDir, filepath.Base(t.Name()))
								require.NoError(t, err)
								defer func() {
									assert.NoError(t, os.RemoveAll(file.Name()))
								}()
								require.NoError(t, file.Close())
								extractDir, err := ioutil.TempDir(tempDir, filepath.Base(t.Name())+"_extract")
								require.NoError(t, err)
								defer func() {
									assert.NoError(t, os.RemoveAll(file.Name()))
								}()

								opts := options.Download{
									URL:  "https://example.com",
									Path: file.Name(),
									ArchiveOpts: options.Archive{
										ShouldExtract: true,
										Format:        options.ArchiveFormat("foo"),
										TargetPath:    extractDir,
									},
								}
								assert.Error(t, mngr.DownloadFile(ctx, opts))
							},
							"FailsForUnarchivedFile": func(ctx context.Context, t *testing.T, mngr Manager, tempDir string) {
								extractDir, err := ioutil.TempDir(tempDir, filepath.Base(t.Name())+"_extract")
								require.NoError(t, err)
								defer func() {
									assert.NoError(t, os.RemoveAll(extractDir))
								}()
								opts := options.Download{
									URL:  "https://example.com",
									Path: filepath.Join(tempDir, filepath.Base(t.Name())),
									ArchiveOpts: options.Archive{
										ShouldExtract: true,
										Format:        options.ArchiveAuto,
										TargetPath:    extractDir,
									},
								}
								assert.Error(t, mngr.DownloadFile(ctx, opts))

								dirContents, err := ioutil.ReadDir(extractDir)
								require.NoError(t, err)
								assert.Zero(t, len(dirContents))
							},
							"FailsForInvalidURL": func(ctx context.Context, t *testing.T, mngr Manager, tempDir string) {
								file, err := ioutil.TempFile(tempDir, filepath.Base(t.Name()))
								require.NoError(t, err)
								defer func() {
									assert.NoError(t, os.RemoveAll(file.Name()))
								}()
								require.NoError(t, file.Close())
								assert.Error(t, mngr.DownloadFile(ctx, options.Download{URL: "", Path: file.Name()}))
							},
							"FailsForNonexistentURL": func(ctx context.Context, t *testing.T, mngr Manager, tempDir string) {
								file, err := ioutil.TempFile(tempDir, "out.txt")
								require.NoError(t, err)
								defer func() {
									assert.NoError(t, os.RemoveAll(file.Name()))
								}()
								require.NoError(t, file.Close())
								assert.Error(t, mngr.DownloadFile(ctx, options.Download{URL: "https://example.com/foo", Path: file.Name()}))
							},
							"FailsForInsufficientPermissions": func(ctx context.Context, t *testing.T, mngr Manager, tempDir string) {
								if os.Geteuid() == 0 {
									t.Skip("cannot test download permissions as root")
								} else if runtime.GOOS == "windows" {
									t.Skip("cannot test download permissions on windows")
								}
								assert.Error(t, mngr.DownloadFile(ctx, options.Download{URL: "https://example.com", Path: "/foo/bar"}))
							},
						} {
							t.Run(testName, func(t *testing.T) {
								tempDir, err := ioutil.TempDir(testutil.BuildDirectory(), filepath.Base(t.Name()))
								require.NoError(t, err)
								defer func() {
									assert.NoError(t, os.RemoveAll(tempDir))
								}()
								testCase(ctx, t, mngr, tempDir)
							})
						}
					},
				},
				{
					Name: "GetBuildloggerURLsFailsWithoutBuildlogger",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						logger := &options.LoggerConfig{}
						require.NoError(t, logger.Set(&options.DefaultLoggerOptions{
							Base: options.BaseOptions{Format: options.LogFormatPlain},
						}))
						opts := &options.Create{
							Args: []string{"echo", "foobar"},
							Output: options.Output{
								Loggers: []*options.LoggerConfig{logger},
							},
						}

						info, err := mngr.CreateProcess(ctx, opts)
						require.NoError(t, err)
						id := info.ID()
						assert.NotEmpty(t, id)

						urls, err := mngr.GetBuildloggerURLs(ctx, id)
						assert.Error(t, err)
						assert.Nil(t, urls)
					},
				},
				{
					Name: "GetBuildloggerURLsFailsWithNonexistentProcess",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						urls, err := mngr.GetBuildloggerURLs(ctx, "foo")
						assert.Error(t, err)
						assert.Nil(t, urls)
					},
				},
				{
					Name: "CreateProcessWithLogFile",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						file, err := ioutil.TempFile(testutil.BuildDirectory(), filepath.Base(t.Name()))
						require.NoError(t, err)
						defer func() {
							assert.NoError(t, os.RemoveAll(file.Name()))
						}()
						require.NoError(t, file.Close())

						logger := &options.LoggerConfig{}
						require.NoError(t, logger.Set(&options.FileLoggerOptions{
							Filename: file.Name(),
							Base:     options.BaseOptions{Format: options.LogFormatPlain},
						}))
						output := "foobar"
						opts := &options.Create{
							Args: []string{"echo", output},
							Output: options.Output{
								Loggers: []*options.LoggerConfig{logger},
							},
						}

						proc, err := mngr.CreateProcess(ctx, opts)
						require.NoError(t, err)

						exitCode, err := proc.Wait(ctx)
						require.NoError(t, err)
						require.Zero(t, exitCode)

						info, err := os.Stat(file.Name())
						require.NoError(t, err)
						assert.NotZero(t, info.Size())

						fileContents, err := ioutil.ReadFile(file.Name())
						require.NoError(t, err)
						assert.Contains(t, string(fileContents), output)
					},
				},
				{
					Name: "RegisterSignalTriggerIDChecksForInvalidTriggerID",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						opts := testutil.SleepCreateOpts(1)
						proc, err := mngr.CreateProcess(ctx, opts)
						require.NoError(t, err)
						assert.True(t, proc.Running(ctx))

						assert.Error(t, proc.RegisterSignalTriggerID(ctx, jasper.SignalTriggerID("foo")))

						assert.NoError(t, proc.Signal(ctx, syscall.SIGTERM))
					},
				},
				{
					Name: "RegisterSignalTriggerIDPassesWithValidArgs",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						opts := testutil.SleepCreateOpts(1)
						proc, err := mngr.CreateProcess(ctx, opts)
						require.NoError(t, err)
						assert.True(t, proc.Running(ctx))

						assert.NoError(t, proc.RegisterSignalTriggerID(ctx, jasper.CleanTerminationSignalTrigger))

						assert.NoError(t, proc.Signal(ctx, syscall.SIGTERM))
					},
				},
				{
					Name: "SendMessagesFailsWithNonexistentLogger",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						payload := options.LoggingPayload{
							LoggerID: "nonexistent",
							Data:     "new log message",
							Priority: level.Warning,
							Format:   options.LoggingPayloadFormatString,
						}
						assert.Error(t, mngr.SendMessages(ctx, payload))
					},
				},
				{
					Name: "SendMessagesSucceeds",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						lc := mngr.LoggingCache(ctx)
						tmpDir, err := ioutil.TempDir(testutil.BuildDirectory(), "logging_cache")
						require.NoError(t, err)
						defer func() {
							assert.NoError(t, os.RemoveAll(tmpDir))
						}()
						tmpFile := filepath.Join(tmpDir, "send_messages")

						fileOpts := &options.FileLoggerOptions{
							Filename: tmpFile,
							Base: options.BaseOptions{
								Format: options.LogFormatPlain,
							},
						}
						config := &options.LoggerConfig{}
						require.NoError(t, config.Set(fileOpts))

						logger, err := lc.Create("logger", &options.Output{
							Loggers: []*options.LoggerConfig{config},
						})
						require.NoError(t, err)
						defer func() {
							assert.NoError(t, lc.Clear(ctx))
						}()

						payload := options.LoggingPayload{
							LoggerID: logger.ID,
							Data:     "new log message",
							Priority: level.Info,
							Format:   options.LoggingPayloadFormatString,
						}
						assert.NoError(t, mngr.SendMessages(ctx, payload))

						content, err := ioutil.ReadFile(tmpFile)
						require.NoError(t, err)
						assert.Equal(t, payload.Data, strings.TrimSpace(string(content)))
					},
				},
				{
					Name: "CreateScriptingSucceeds",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						tmpDir, err := ioutil.TempDir(testutil.BuildDirectory(), "scripting_tests")
						require.NoError(t, err)
						defer func() {
							assert.NoError(t, os.RemoveAll(tmpDir))
						}()
						sh := createTestScriptingHarness(ctx, t, mngr, tmpDir)
						assert.NotZero(t, sh)
					},
				},
				{
					Name: "CreateScriptingFailsWithInvalidOptions",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						sh, err := mngr.CreateScripting(ctx, &options.ScriptingGolang{})
						assert.Error(t, err)
						assert.Zero(t, sh)
					},
				},
				{
					Name: "GetScriptingWithNonexistentHarnessFails",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						_, err := mngr.GetScripting(ctx, "nonexistent")
						assert.Error(t, err)
					},
				},
				{
					Name: "GetScriptingWithExistingHarnessSucceeds",
					Case: func(ctx context.Context, t *testing.T, mngr Manager) {
						tmpDir, err := ioutil.TempDir(testutil.BuildDirectory(), "scripting_tests")
						require.NoError(t, err)
						defer func() {
							assert.NoError(t, os.RemoveAll(tmpDir))
						}()
						expectedHarness := createTestScriptingHarness(ctx, t, mngr, tmpDir)

						harness, err := mngr.GetScripting(ctx, expectedHarness.ID())
						require.NoError(t, err)
						assert.Equal(t, expectedHarness.ID(), harness.ID())
					},
				},
			} {
				t.Run(testCase.Name, func(t *testing.T) {
					tctx, tcancel := context.WithTimeout(ctx, testutil.RPCTestTimeout)
					defer tcancel()
					mngr := makeManager(tctx, t)
					defer func() {
						assert.NoError(t, mngr.CloseConnection())
					}()
					testCase.Case(tctx, t, mngr)
				})
			}
		})
	}
}

func createProcs(ctx context.Context, opts *options.Create, mngr Manager, num int) ([]jasper.Process, error) {
	catcher := grip.NewBasicCatcher()
	var procs []jasper.Process
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
