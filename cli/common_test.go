package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

var nextPort <-chan int

func init() {
	nextPort = func() <-chan int {
		out := make(chan int, 25)
		go func() {
			id := 3000
			for {
				id++
				out <- id
			}
		}()
		return out
	}()
}

func noWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func getNextPort() int {
	return <-nextPort
}

const testTimeout = 1 * time.Second

func buildDir(t *testing.T) string {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(filepath.Dir(cwd), "build")
}

func trueCreateOpts() *jasper.CreateOptions {
	return &jasper.CreateOptions{Args: []string{"true"}}
}

func yesCreateOpts(timeoutSecs int) *jasper.CreateOptions {
	return &jasper.CreateOptions{Args: []string{"yes"}, TimeoutSecs: timeoutSecs}
}

func mockCLIContext(service string, port int) *cli.Context {
	flags := &flag.FlagSet{}
	_ = flags.String(serviceFlagName, service, "")
	_ = flags.Int(portFlagName, port, "")
	_ = flags.String(hostFlagName, "localhost", "")
	_ = flags.String(certFilePathFlagName, "", "")
	return cli.NewContext(nil, flags, nil)
}

type mockInput struct {
	Value     string `json:"value"`
	validated bool
}

func (m *mockInput) Validate() error {
	m.validated = true
	return nil
}

type mockOutput struct {
	Value string `json:"value"`
}

// mockRequest returns a function that returns a mockOutput with the given
// value val.
func mockRequest(val string) func(context.Context, jasper.RemoteClient) interface{} {
	return func(context.Context, jasper.RemoteClient) interface{} {
		return mockOutput{val}
	}
}

// withMockStdin runs the operation with a stdin that contains the given input.
// It passes the mocked stdin as a parameter to the operation.
func withMockStdin(t *testing.T, input string, operation func(*os.File) error) error {
	stdin := os.Stdin
	defer func() {
		os.Stdin = stdin
	}()
	tmpFile, err := ioutil.TempFile(buildDir(t), "mock_stdin.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString(input)
	require.NoError(t, err)
	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)
	os.Stdin = tmpFile
	return operation(os.Stdin)
}

// withMockStdout runs the operation with a stdout that can be inspected as a
// regular file. It passes the mocked stdout to the operation.
func withMockStdout(t *testing.T, operation func(*os.File) error) error {
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()
	tmpFile, err := ioutil.TempFile(buildDir(t), "mock_stdout.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	os.Stdout = tmpFile
	return operation(os.Stdout)
}

// waitForRESTService waits until the REST service becomes available to serve
// requests or the context times out.
func waitForRESTService(ctx context.Context, t *testing.T, url string) {
	// Block until the service comes up
	timeoutInterval := 10 * time.Millisecond
	timer := time.NewTimer(timeoutInterval)
	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "test timed out before REST service was available")
			return
		case <-timer.C:
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				timer.Reset(timeoutInterval)
				continue
			}
			req = req.WithContext(ctx)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				timer.Reset(timeoutInterval)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				timer.Reset(timeoutInterval)
				continue
			}
			return
		}
	}
}

func TestRemoteClientInvalidService(t *testing.T) {
	ctx := context.Background()
	client, err := remoteClient(ctx, "invalid", "localhost", getNextPort(), "")
	require.Error(t, err)
	require.Nil(t, client)
}

func TestRemoteClient(t *testing.T) {
	for remoteType, makeServiceAndClient := range map[string]func(ctx context.Context, t *testing.T, port int, manager jasper.Manager) (jasper.CloseFunc, jasper.RemoteClient){
		serviceREST: func(ctx context.Context, t *testing.T, port int, manager jasper.Manager) (jasper.CloseFunc, jasper.RemoteClient) {
			closeService := func() error { return nil }
			client, err := remoteClient(ctx, serviceREST, "localhost", port, "")
			require.NoError(t, err)
			return closeService, client
		},
		serviceRPC: func(ctx context.Context, t *testing.T, port int, manager jasper.Manager) (jasper.CloseFunc, jasper.RemoteClient) {
			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", "localhost", port))
			require.NoError(t, err)
			closeService, err := rpc.StartService(ctx, manager, addr, "", "")
			require.NoError(t, err)
			client, err := remoteClient(ctx, serviceRPC, "localhost", port, "")
			require.NoError(t, err)
			return closeService, client
		},
	} {
		t.Run(remoteType, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()
			manager, err := jasper.NewLocalManager(false)
			require.NoError(t, err)
			closeService, client := makeServiceAndClient(ctx, t, getNextPort(), manager)
			assert.NoError(t, closeService())
			assert.NoError(t, client.CloseConnection())
		})
	}
}

func TestCLICommon(t *testing.T) {
	for remoteType, makeServiceAndClient := range map[string]func(ctx context.Context, t *testing.T, port int, manager jasper.Manager) (jasper.CloseFunc, jasper.RemoteClient){
		serviceREST: func(ctx context.Context, t *testing.T, port int, manager jasper.Manager) (jasper.CloseFunc, jasper.RemoteClient) {
			srv := jasper.NewManagerService(manager)
			app := srv.App(ctx)
			app.SetPrefix("jasper")
			require.NoError(t, app.SetPort(port))

			go func() {
				assert.NoError(t, app.Run(ctx))
			}()

			waitForRESTService(ctx, t, fmt.Sprintf("http://localhost:%d/jasper/v1", port))

			closeService := func() error { return nil }
			client, err := remoteClient(ctx, serviceREST, "localhost", port, "")
			require.NoError(t, err)
			return closeService, client
		},
		serviceRPC: func(ctx context.Context, t *testing.T, port int, manager jasper.Manager) (jasper.CloseFunc, jasper.RemoteClient) {
			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", "localhost", port))
			require.NoError(t, err)

			closeService, err := rpc.StartService(ctx, manager, addr, "", "")
			require.NoError(t, err)

			client, err := remoteClient(ctx, serviceRPC, "localhost", port, "")
			require.NoError(t, err)
			return closeService, client
		},
	} {
		t.Run(remoteType, func(t *testing.T) {
			for testName, testCase := range map[string]func(ctx context.Context, t *testing.T, c *cli.Context, client jasper.RemoteClient){
				"CreateProcessWithConnection": func(ctx context.Context, t *testing.T, c *cli.Context, client jasper.RemoteClient) {
					withConnection(ctx, c, func(client jasper.RemoteClient) error {
						proc, err := client.CreateProcess(ctx, trueCreateOpts())
						require.NoError(t, err)
						require.NotNil(t, proc)
						assert.NotZero(t, proc.Info(ctx).PID)
						return nil
					})
				},
				"DoPassthroughInputOutputReadsFromStdin": func(ctx context.Context, t *testing.T, c *cli.Context, client jasper.RemoteClient) {
					withMockStdin(t, `{"value":"foo"}`, func(stdin *os.File) error {
						input := &mockInput{}
						require.NoError(t, doPassthroughInputOutput(c, input, mockRequest("")))
						output, err := ioutil.ReadAll(stdin)
						require.NoError(t, err)
						assert.Len(t, output, 0)
						return nil
					})
				},
				"DoPassthroughInputOutputSetsAndValidatesInput": func(ctx context.Context, t *testing.T, c *cli.Context, client jasper.RemoteClient) {
					expectedInput := "foo"
					withMockStdin(t, fmt.Sprintf(`{"value":"%s"}`, expectedInput), func(*os.File) error {
						input := &mockInput{}
						require.NoError(t, doPassthroughInputOutput(c, input, mockRequest("")))
						assert.Equal(t, expectedInput, input.Value)
						assert.True(t, input.validated)
						return nil
					})
				},
				"DoPassthroughInputOutputWritesResponseToStdout": func(ctx context.Context, t *testing.T, c *cli.Context, client jasper.RemoteClient) {
					withMockStdin(t, `{"value":"foo"}`, func(*os.File) error {
						return withMockStdout(t, func(stdout *os.File) error {
							input := &mockInput{}
							outputVal := "bar"
							require.NoError(t, doPassthroughInputOutput(c, input, mockRequest(outputVal)))
							assert.Equal(t, "foo", input.Value)
							assert.True(t, input.validated)

							expectedOutput := `{"value":"bar"}`
							_, err := stdout.Seek(0, 0)
							require.NoError(t, err)
							output, err := ioutil.ReadAll(stdout)
							require.NoError(t, err)
							assert.Equal(t, noWhitespace(expectedOutput), noWhitespace(string(output)))
							return nil
						})
					})
				},
				"DoPassthroughOutputIgnoresStdin": func(ctx context.Context, t *testing.T, c *cli.Context, client jasper.RemoteClient) {
					input := "foo"
					withMockStdin(t, input, func(stdin *os.File) error {
						require.NoError(t, doPassthroughOutput(c, mockRequest("")))
						output, err := ioutil.ReadAll(stdin)
						require.NoError(t, err)
						assert.Len(t, output, len(input))
						return nil
					})
				},
				"DoPassthroughOutputWritesResponseToStdout": func(ctx context.Context, t *testing.T, c *cli.Context, client jasper.RemoteClient) {
					withMockStdout(t, func(stdout *os.File) error {
						outputVal := "bar"
						require.NoError(t, doPassthroughOutput(c, mockRequest(outputVal)))

						expectedOutput := `{"value": "bar"}`
						_, err := stdout.Seek(0, 0)
						require.NoError(t, err)
						output, err := ioutil.ReadAll(stdout)
						require.NoError(t, err)
						assert.Equal(t, noWhitespace(expectedOutput), noWhitespace(string(output)))
						return nil
					})
				},
				// "": func(ctx context.Context, t *testing.T, c *cli.Context, client jasper.RemoteClient) {},
			} {
				t.Run(testName, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
					defer cancel()
					port := getNextPort()
					c := mockCLIContext(remoteType, port)
					manager, err := jasper.NewLocalManager(false)
					require.NoError(t, err)
					closeService, client := makeServiceAndClient(ctx, t, port, manager)
					defer func() {
						assert.NoError(t, client.CloseConnection())
						assert.NoError(t, closeService())
					}()

					testCase(ctx, t, c, client)
				})
			}
		})
	}
}

func execCLICommandInputOutput(t *testing.T, c *cli.Context, cmd cli.Command, input []byte, output interface{}) error {
	return withMockStdin(t, string(input), func(*os.File) error {
		return execCLICommandOutput(t, c, cmd, output)
	})
}

func execCLICommandOutput(t *testing.T, c *cli.Context, cmd cli.Command, output interface{}) error {
	return withMockStdout(t, func(stdout *os.File) error {
		if err := cli.HandleAction(cmd.Action, c); err != nil {
			return err
		}
		if _, err := stdout.Seek(0, 0); err != nil {
			return err
		}
		resp, err := ioutil.ReadAll(stdout)
		if err != nil {
			return err
		}
		return json.Unmarshal(resp, output)
	})
}
