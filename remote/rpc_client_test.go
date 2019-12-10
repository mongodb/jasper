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

	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
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
		"Insecure": makeInsecureRPCServiceAndClient,
		"TLS":      makeTLSRPCServiceAndClient,
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
					for _, test := range AddBasicClientTests(
						// The following test cases are added specifically for the
						// RemoteClient.
						ClientTestCase{
							Name: "WithInMemoryLogger",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
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
						},
						ClientTestCase{
							Name: "GetLogStreamFromNonexistentProcessFails",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
								stream, err := client.GetLogStream(ctx, "foo", 1)
								assert.Error(t, err)
								assert.Zero(t, stream)
							},
						},
						ClientTestCase{
							Name: "GetLogStreamFailsWithoutInMemoryLogger",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
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
						},

						///////////////////////////////////
						//
						// The following test cases are added
						// specifically for the rpc case

						ClientTestCase{
							Name: "RegisterIsDisabled",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
								err := client.Register(ctx, nil)
								require.Error(t, err)
								assert.Contains(t, err.Error(), "cannot register")
							},
						},
						ClientTestCase{
							Name: "CreateProcessReturnsCorrectExample",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
								proc, err := client.CreateProcess(ctx, testutil.TrueCreateOpts())
								require.NoError(t, err)
								assert.NotNil(t, proc)
								assert.NotZero(t, proc.ID())

								fetched, err := client.Get(ctx, proc.ID())
								assert.NoError(t, err)
								assert.NotNil(t, fetched)
								assert.Equal(t, proc.ID(), fetched.ID())
							},
						},
						ClientTestCase{
							Name: "WaitOnSigKilledProcessReturnsProperExitCode",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
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
						},
						ClientTestCase{
							Name: "StandardInput",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
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
						},
						ClientTestCase{
							Name: "WriteFileSucceeds",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
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
						},
						ClientTestCase{
							Name: "WriteFileAcceptsContentFromReader",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
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
						},
						ClientTestCase{
							Name: "WriteFileSucceedsWithLargeContent",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
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
						},
						ClientTestCase{
							Name: "WriteFileFailsWithInvalidPath",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
								opts := options.WriteFile{Content: []byte("foo")}
								assert.Error(t, client.WriteFile(ctx, opts))
							},
						},
						ClientTestCase{
							Name: "WriteFileSucceedsWithNoContent",
							Case: func(ctx context.Context, t *testing.T, client Manager) {
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
						},
					) {
						t.Run(test.Name, func(t *testing.T) {
							tctx, cancel := context.WithTimeout(ctx, testutil.RPCTestTimeout)
							defer cancel()
							test.Case(tctx, t, factory(tctx, t))
						})
					}
				})
			}
		})
	}
}
