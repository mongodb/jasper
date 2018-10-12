package jrpc

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/jrpc/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestJRPCService(t *testing.T) {
	for managerName, makeManager := range map[string]func() jasper.Manager{
		"Blocking":    jasper.NewLocalManagerBlockingProcesses,
		"Nonblocking": jasper.NewLocalManager,
	} {
		t.Run(managerName, func(t *testing.T) {
			for testName, testCase := range map[string]func(context.Context, *testing.T, jasper.CreateOptions, internal.JasperProcessManagerClient, string){
				"CreateWithLogFile": func(ctx context.Context, t *testing.T, opts jasper.CreateOptions, client internal.JasperProcessManagerClient, output string) {
					cwd, err := os.Getwd()
					require.NoError(t, err)
					file, err := ioutil.TempFile(filepath.Join(filepath.Dir(cwd), "build"), "out.txt")
					require.NoError(t, err)
					defer os.Remove(file.Name())

					logger := jasper.Logger{
						Type: jasper.LogFile,
						Options: jasper.LogOptions{
							FileName: file.Name(),
							Format:   jasper.LogFormatPlain,
						},
					}
					opts.Output.Loggers = []jasper.Logger{logger}

					procInfo, err := client.Create(ctx, internal.ConvertCreateOptions(&opts))
					assert.NoError(t, err)
					assert.NotNil(t, procInfo)

					outcome, err := client.Wait(ctx, &internal.JasperProcessID{Value: procInfo.Id})
					assert.NoError(t, err)
					assert.True(t, outcome.Success)

					info, err := os.Stat(file.Name())
					assert.NoError(t, err)
					assert.NotZero(t, info.Size())

					fileContents, err := ioutil.ReadFile(file.Name())
					assert.NoError(t, err)
					assert.Contains(t, string(fileContents), output)
				},
			} {
				t.Run(testName, func(t *testing.T) {
					tctx, tcancel := context.WithTimeout(context.Background(), taskTimeout)
					defer tcancel()
					output := "foobar"
					opts := jasper.CreateOptions{Args: []string{"echo", output}}

					manager := makeManager()
					addr, err := startJRPC(tctx, manager)
					require.NoError(t, err)

					conn, err := grpc.DialContext(tctx, addr, grpc.WithInsecure(), grpc.WithBlock())
					require.NoError(t, err)
					client := internal.NewJasperProcessManagerClient(conn)

					go func() {
						<-tctx.Done()
						conn.Close()
					}()

					testCase(tctx, t, opts, client, output)
				})
			}
		})
	}
}
