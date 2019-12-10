package remote

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

	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: these tests are largely copied directly from the top level package into
// this package to avoid an import cycle.

func TestRPCClient(t *testing.T) {
	assert.NotPanics(t, func() {
		newRPCClient(nil)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for setupMethod, makeTestServiceAndClient := range map[string]func(ctx context.Context, mngr jasper.Manager) (Manager, error){
		"Insecure": makeInsecureServiceAndClient,
		"TLS":      makeTLSServiceAndClient,
	} {
		t.Run(setupMethod, func(t *testing.T) {
			for mname, factory := range map[string]func(ctx context.Context, t *testing.T) Manager{
				"Basic": func(ctx context.Context, t *testing.T) Manager {
					mngr, err := jasper.NewSynchronizedManager(false)
					require.NoError(t, err)

					client, err := makeTestServiceAndClient(ctx, mngr)
					require.NoError(t, err)
					return client
				},
			} {
				t.Run(mname, func(t *testing.T) {
					for name, test := range map[string]func(context.Context, *testing.T, Manager){

						// The following test cases are added specifically for the
						// RemoteClient.

						"WithInMemoryLogger": func(ctx context.Context, t *testing.T, client Manager) {
							output := "foo"
							opts := &options.Create{
								Args: []string{"echo", output},
								Output: options.Output{
									Loggers: []options.Logger{
										{
											Type:    options.LogInMemory,
											Options: options.Log{InMemoryCap: 100, Format: options.LogFormatPlain},
										},
									},
								},
							}

							for testName, testCase := range map[string]func(ctx context.Context, t *testing.T, proc jasper.Process){
								"GetLogStreamFailsForInvalidCount": func(ctx context.Context, t *testing.T, proc jasper.Process) {
									stream, err := client.GetLogStream(ctx, proc.ID(), -1)
									assert.Error(t, err)
									assert.Zero(t, stream)
								},
								"GetLogStreamReturnsOutputOnSuccess": func(ctx context.Context, t *testing.T, proc jasper.Process) {
									logs := []string{}
									for stream, err := client.GetLogStream(ctx, proc.ID(), 1); !stream.Done; stream, err = client.GetLogStream(ctx, proc.ID(), 1) {
										require.NoError(t, err)
										require.NotEmpty(t, stream.Logs)
										logs = append(logs, stream.Logs...)
									}
									assert.Contains(t, logs, output)
								},
							} {
								t.Run(testName, func(t *testing.T) {
									proc, err := client.CreateProcess(ctx, opts)
									require.NoError(t, err)
									require.NotNil(t, proc)

									_, err = proc.Wait(ctx)
									require.NoError(t, err)
									testCase(ctx, t, proc)
								})
							}
						},
						"GetLogStreamFromNonexistentProcessFails": func(ctx context.Context, t *testing.T, client Manager) {
							stream, err := client.GetLogStream(ctx, "foo", 1)
							assert.Error(t, err)
							assert.Zero(t, stream)
						},
						"GetLogStreamFailsWithoutInMemoryLogger": func(ctx context.Context, t *testing.T, client Manager) {
							opts := &options.Create{Args: []string{"echo", "foo"}}

							proc, err := client.CreateProcess(ctx, opts)
							require.NoError(t, err)
							require.NotNil(t, proc)

							_, err = proc.Wait(ctx)
							require.NoError(t, err)

							stream, err := client.GetLogStream(ctx, proc.ID(), 1)
							assert.Error(t, err)
							assert.Zero(t, stream)
						},
						// "": func(ctx context.Context, t *testing.T, client Manager) {},

						///////////////////////////////////
						//
						// The following test cases are added
						// specifically for the rpc case

						"RegisterIsDisabled": func(ctx context.Context, t *testing.T, client Manager) {
							err := client.Register(ctx, nil)
							require.Error(t, err)
							assert.Contains(t, err.Error(), "cannot register")
						},
						"CreateProcessReturnsCorrectExample": func(ctx context.Context, t *testing.T, client Manager) {
							proc, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
							require.NoError(t, err)
							assert.NotNil(t, proc)
							assert.NotZero(t, proc.ID())

							fetched, err := client.Get(ctx, proc.ID())
							assert.NoError(t, err)
							assert.NotNil(t, fetched)
							assert.Equal(t, proc.ID(), fetched.ID())
						},
						"WaitOnSigKilledProcessReturnsProperExitCode": func(ctx context.Context, t *testing.T, client Manager) {
							proc, err := client.CreateProcess(ctx, testutil.SleepCreateOpts(100))
							require.NoError(t, err)
							require.NotNil(t, proc)
							require.NotZero(t, proc.ID())

							require.NoError(t, proc.Signal(ctx, syscall.SIGKILL))

							exitCode, err := proc.Wait(ctx)
							require.Error(t, err)
							if runtime.GOOS == "windows" {
								assert.Equal(t, 1, exitCode)
							} else {
								assert.Equal(t, 9, exitCode)
							}
						},
						"StandardInput": func(ctx context.Context, t *testing.T, client Manager) {
							for subTestName, subTestCase := range map[string]func(ctx context.Context, t *testing.T, opts *options.Create, expectedOutput string, stdin []byte){
								"ReaderIsIgnored": func(ctx context.Context, t *testing.T, opts *options.Create, expectedOutput string, stdin []byte) {
									opts.StandardInput = bytes.NewBuffer(stdin)

									proc, err := client.CreateProcess(ctx, opts)
									require.NoError(t, err)

									_, err = proc.Wait(ctx)
									require.NoError(t, err)

									logs, err := client.GetLogStream(ctx, proc.ID(), 1)
									require.NoError(t, err)
									assert.Empty(t, logs.Logs)
								},
								"BytesSetsStandardInput": func(ctx context.Context, t *testing.T, opts *options.Create, expectedOutput string, stdin []byte) {
									opts.StandardInputBytes = stdin

									proc, err := client.CreateProcess(ctx, opts)
									require.NoError(t, err)

									_, err = proc.Wait(ctx)
									require.NoError(t, err)

									logs, err := client.GetLogStream(ctx, proc.ID(), 1)
									require.NoError(t, err)

									require.Len(t, logs.Logs, 1)
									assert.Equal(t, expectedOutput, strings.TrimSpace(logs.Logs[0]))
								},
								"BytesCopiedByRespawnedProcess": func(ctx context.Context, t *testing.T, opts *options.Create, expectedOutput string, stdin []byte) {
									opts.StandardInputBytes = stdin

									proc, err := client.CreateProcess(ctx, opts)
									require.NoError(t, err)

									_, err = proc.Wait(ctx)
									require.NoError(t, err)

									logs, err := client.GetLogStream(ctx, proc.ID(), 1)
									require.NoError(t, err)

									require.Len(t, logs.Logs, 1)
									assert.Equal(t, expectedOutput, strings.TrimSpace(logs.Logs[0]))

									newProc, err := proc.Respawn(ctx)
									require.NoError(t, err)

									_, err = newProc.Wait(ctx)
									require.NoError(t, err)

									logs, err = client.GetLogStream(ctx, newProc.ID(), 1)
									require.NoError(t, err)

									require.Len(t, logs.Logs, 1)
									assert.Equal(t, expectedOutput, strings.TrimSpace(logs.Logs[0]))
								},
							} {
								t.Run(subTestName, func(t *testing.T) {
									opts := &options.Create{
										Args: []string{"bash", "-s"},
										Output: options.Output{
											Loggers: []options.Logger{jasper.NewInMemoryLogger(1)},
										},
									}
									expectedOutput := "foobar"
									stdin := []byte("echo " + expectedOutput)
									subTestCase(ctx, t, opts, expectedOutput, stdin)
								})
							}
						},
						"WriteFileSucceeds": func(ctx context.Context, t *testing.T, client Manager) {
							tmpFile, err := ioutil.TempFile(buildDir(t), filepath.Base(t.Name()))
							require.NoError(t, err)
							defer func() {
								assert.NoError(t, tmpFile.Close())
								assert.NoError(t, os.RemoveAll(tmpFile.Name()))
							}()

							opts := options.WriteFile{Path: tmpFile.Name(), Content: []byte("foo")}
							require.NoError(t, client.WriteFile(ctx, opts))

							content, err := ioutil.ReadFile(tmpFile.Name())
							require.NoError(t, err)

							assert.Equal(t, opts.Content, content)
						},
						"WriteFileAcceptsContentFromReader": func(ctx context.Context, t *testing.T, client Manager) {
							tmpFile, err := ioutil.TempFile(buildDir(t), filepath.Base(t.Name()))
							require.NoError(t, err)
							defer func() {
								assert.NoError(t, tmpFile.Close())
								assert.NoError(t, os.RemoveAll(tmpFile.Name()))
							}()

							buf := []byte("foo")
							opts := options.WriteFile{Path: tmpFile.Name(), Reader: bytes.NewBuffer(buf)}
							require.NoError(t, client.WriteFile(ctx, opts))

							content, err := ioutil.ReadFile(tmpFile.Name())
							require.NoError(t, err)

							assert.Equal(t, buf, content)
						},
						"WriteFileSucceedsWithLargeContent": func(ctx context.Context, t *testing.T, client Manager) {
							tmpFile, err := ioutil.TempFile(buildDir(t), filepath.Base(t.Name()))
							require.NoError(t, err)
							defer func() {
								assert.NoError(t, tmpFile.Close())
								assert.NoError(t, os.RemoveAll(tmpFile.Name()))
							}()

							const mb = 1024 * 1024
							opts := options.WriteFile{Path: tmpFile.Name(), Content: bytes.Repeat([]byte("foo"), mb)}
							require.NoError(t, client.WriteFile(ctx, opts))

							content, err := ioutil.ReadFile(tmpFile.Name())
							require.NoError(t, err)

							assert.Equal(t, opts.Content, content)
						},
						"WriteFileFailsWithInvalidPath": func(ctx context.Context, t *testing.T, client Manager) {
							opts := options.WriteFile{Content: []byte("foo")}
							assert.Error(t, client.WriteFile(ctx, opts))
						},
						"WriteFileSucceedsWithNoContent": func(ctx context.Context, t *testing.T, client Manager) {
							path := filepath.Join(buildDir(t), filepath.Base(t.Name()))
							require.NoError(t, os.RemoveAll(path))
							defer func() {
								assert.NoError(t, os.RemoveAll(path))
							}()

							opts := options.WriteFile{Path: path}
							require.NoError(t, client.WriteFile(ctx, opts))

							stat, err := os.Stat(path)
							require.NoError(t, err)

							assert.Zero(t, stat.Size())
						},
						// "": func(ctx context.Context, t *testing.T, client Manager) {},
					} {
						t.Run(name, func(t *testing.T) {
							tctx, cancel := context.WithTimeout(ctx, testutil.RPCTestTimeout)
							defer cancel()
							test(tctx, t, factory(tctx, t))
						})
					}
				})
			}
		})
	}
}

func TestRPCProcess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for setupMethod, makeTestServiceAndClient := range map[string]func(ctx context.Context, mngr jasper.Manager) (Manager, error){
		"Insecure": makeInsecureServiceAndClient,
		"TLS":      makeTLSServiceAndClient,
	} {
		t.Run(setupMethod, func(t *testing.T) {
			for cname, makeProc := range map[string]jasper.ProcessConstructor{
				"Basic": func(ctx context.Context, opts *options.Create) (jasper.Process, error) {
					mngr, err := jasper.NewSynchronizedManager(false)
					if err != nil {
						return nil, errors.WithStack(err)
					}

					client, err := makeTestServiceAndClient(ctx, mngr)
					if err != nil {
						return nil, errors.WithStack(err)
					}

					return client.CreateProcess(ctx, opts)
				},
			} {
				t.Run(cname, func(t *testing.T) {
					for name, testCase := range map[string]func(context.Context, *testing.T, *options.Create, jasper.ProcessConstructor){
						"WithPopulatedArgsCommandCreationPasses": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							assert.NotZero(t, opts.Args)
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							assert.NotNil(t, proc)
						},
						"ErrorToCreateWithInvalidArgs": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							opts.Args = []string{}
							proc, err := makep(ctx, opts)
							require.Error(t, err)
							assert.Nil(t, proc)
						},
						"WithCanceledContextProcessCreationFails": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							pctx, pcancel := context.WithCancel(ctx)
							pcancel()
							proc, err := makep(pctx, opts)
							require.Error(t, err)
							assert.Nil(t, proc)
						},
						"CanceledContextTimesOutEarly": func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
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
						"ProcessLacksTagsByDefault": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							tags := proc.GetTags()
							assert.Empty(t, tags)
						},
						"ProcessTagsPersist": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							opts.Tags = []string{"foo"}
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							tags := proc.GetTags()
							assert.Contains(t, tags, "foo")
						},
						"InfoHasMatchingID": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							_, err = proc.Wait(ctx)
							require.NoError(t, err)
							assert.Equal(t, proc.ID(), proc.Info(ctx).ID)
						},
						"ResetTags": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							proc.Tag("foo")
							assert.Contains(t, proc.GetTags(), "foo")
							proc.ResetTags()
							assert.Len(t, proc.GetTags(), 0)
						},
						"TagsAreSetLike": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)

							for i := 0; i < 100; i++ {
								proc.Tag("foo")
							}

							assert.Len(t, proc.GetTags(), 1)
							proc.Tag("bar")
							assert.Len(t, proc.GetTags(), 2)
						},
						"CompleteIsTrueAfterWait": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							time.Sleep(10 * time.Millisecond) // give the process time to start background machinery
							_, err = proc.Wait(ctx)
							assert.NoError(t, err)
							assert.True(t, proc.Complete(ctx))
						},
						"WaitReturnsWithCanceledContext": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
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
						"RegisterTriggerErrorsForNil": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							assert.Error(t, proc.RegisterTrigger(ctx, nil))
						},
						"RegisterSignalTriggerIDErrorsForExitedProcess": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							_, err = proc.Wait(ctx)
							assert.NoError(t, err)
							assert.Error(t, proc.RegisterSignalTriggerID(ctx, jasper.CleanTerminationSignalTrigger))
						},
						"RegisterSignalTriggerIDFailsWithInvalidTriggerID": func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
							opts := testutil.SleepCreateOpts(3)
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							assert.Error(t, proc.RegisterSignalTriggerID(ctx, jasper.SignalTriggerID(-1)))
						},
						"RegisterSignalTriggerIDPassesWithValidTriggerID": func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
							opts := testutil.SleepCreateOpts(3)
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							assert.NoError(t, proc.RegisterSignalTriggerID(ctx, jasper.CleanTerminationSignalTrigger))
						},
						"WaitOnRespawnedProcessDoesNotError": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
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
						"RespawnedProcessGivesSameResult": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
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
						"RespawningFinishedProcessIsOK": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
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
						"RespawningRunningProcessIsOK": func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
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
						"RespawnShowsConsistentStateValues": func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
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
						"WaitGivesSuccessfulExitCode": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, testutil.TrueCreateOpts())
							require.NoError(t, err)
							require.NotNil(t, proc)
							exitCode, err := proc.Wait(ctx)
							assert.NoError(t, err)
							assert.Equal(t, 0, exitCode)
						},
						"WaitGivesFailureExitCode": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, testutil.FalseCreateOpts())
							require.NoError(t, err)
							require.NotNil(t, proc)
							exitCode, err := proc.Wait(ctx)
							require.Error(t, err)
							assert.Equal(t, 1, exitCode)
						},
						"WaitGivesProperExitCodeOnSignalDeath": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
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
						"WaitGivesNegativeOneOnAlternativeError": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
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
						"InfoHasTimeoutWhenProcessTimesOut": func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
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
						"CallingSignalOnDeadProcessDoesError": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)

							_, err = proc.Wait(ctx)
							assert.NoError(t, err)

							err = proc.Signal(ctx, syscall.SIGTERM)
							require.Error(t, err)
							assert.True(t, strings.Contains(err.Error(), "cannot signal a process that has terminated"))
						},

						// "": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {},

						///////////////////////////////////
						//
						// The following test cases are added
						// specifically for the rpc case

						"CompleteReturnsFalseForProcessThatDoesntExist": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)

							firstID := proc.ID()
							_, err = proc.Wait(ctx)
							assert.NoError(t, err)
							assert.True(t, proc.Complete(ctx))
							proc.(*rpcProcess).info.Id += "_foo"
							proc.(*rpcProcess).info.Complete = false
							require.NotEqual(t, firstID, proc.ID())
							assert.False(t, proc.Complete(ctx), proc.ID())
						},
						"RunningReturnsFalseForProcessThatDoesntExist": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)

							firstID := proc.ID()
							_, err = proc.Wait(ctx)
							assert.NoError(t, err)
							proc.(*rpcProcess).info.Id += "_foo"
							proc.(*rpcProcess).info.Complete = false
							require.NotEqual(t, firstID, proc.ID())
							assert.False(t, proc.Running(ctx), proc.ID())
						},
						"CompleteAlwaysReturnsTrueWhenProcessIsComplete": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
							proc, err := makep(ctx, opts)
							require.NoError(t, err)

							_, err = proc.Wait(ctx)
							assert.NoError(t, err)

							assert.True(t, proc.Complete(ctx))
						},
						"RegisterSignalTriggerFails": func(ctx context.Context, t *testing.T, _ *options.Create, makep jasper.ProcessConstructor) {
							opts := testutil.SleepCreateOpts(3)
							proc, err := makep(ctx, opts)
							require.NoError(t, err)
							assert.Error(t, proc.RegisterSignalTrigger(ctx, func(_ jasper.ProcessInfo, _ syscall.Signal) bool {
								return false
							}))
						},
						// "": func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {},
					} {
						t.Run(name, func(t *testing.T) {
							tctx, cancel := context.WithTimeout(ctx, testutil.RPCTestTimeout)
							defer cancel()

							opts := &options.Create{Args: []string{"ls"}}
							testCase(tctx, t, opts, makeProc)
						})
					}
				})
			}

		})
	}
}
