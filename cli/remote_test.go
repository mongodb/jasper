package cli

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
	testoptions "github.com/mongodb/jasper/testutil/options"
	"github.com/mongodb/jasper/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestCLIRemote(t *testing.T) {
	for remoteType, makeService := range map[string]func(ctx context.Context, t *testing.T, port int, manager jasper.Manager) util.CloseFunc{
		RESTService: makeTestRESTService,
		RPCService:  makeTestRPCService,
	} {
		t.Run(remoteType, func(t *testing.T) {
			for testName, testCase := range map[string]func(ctx context.Context, t *testing.T, c *cli.Context){
				"ConfigureCacheSucceeds": func(ctx context.Context, t *testing.T, c *cli.Context) {
					input, err := json.Marshal(options.Cache{})
					require.NoError(t, err)
					resp := &OutcomeResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, remoteConfigureCache(), input, resp))
					assert.True(t, resp.Successful())
				},
				"DownloadFileSucceeds": func(ctx context.Context, t *testing.T, c *cli.Context) {
					tmpFile, err := os.CreateTemp(testutil.BuildDirectory(), "out.txt")
					require.NoError(t, err)
					defer func() {
						assert.NoError(t, tmpFile.Close())
						assert.NoError(t, os.RemoveAll(tmpFile.Name()))
					}()

					input, err := json.Marshal(options.Download{
						URL:  "https://example.com",
						Path: tmpFile.Name(),
					})
					require.NoError(t, err)

					resp := &OutcomeResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, remoteDownloadFile(), input, resp))

					info, err := os.Stat(tmpFile.Name())
					require.NoError(t, err)
					assert.NotZero(t, info.Size)
				},
				"DownloadMongoDBSucceeds": func(ctx context.Context, t *testing.T, c *cli.Context) {
					tmpDir, err := os.MkdirTemp(testutil.BuildDirectory(), "out")
					require.NoError(t, err)
					defer func() {
						assert.NoError(t, os.RemoveAll(tmpDir))
					}()

					opts := testoptions.ValidMongoDBDownloadOptions()
					opts.Path = tmpDir
					input, err := json.Marshal(opts)
					require.NoError(t, err)
					resp := &OutcomeResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, remoteDownloadMongoDB(), input, resp))
				},
				"GetBuildloggerURLsFailsWithNonexistentProcess": func(ctx context.Context, t *testing.T, c *cli.Context) {
					input, err := json.Marshal(IDInput{ID: "foo"})
					require.NoError(t, err)
					resp := &OutcomeResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, remoteGetBuildloggerURLs(), input, resp))
					assert.False(t, resp.Successful())
				},
				"GetLogStreamSucceeds": func(ctx context.Context, t *testing.T, c *cli.Context) {
					logger, err := jasper.NewInMemoryLogger(10)
					require.NoError(t, err)
					opts := testoptions.TrueCreateOpts()
					opts.Output.Loggers = []*options.LoggerConfig{logger}
					createInput, err := json.Marshal(opts)
					require.NoError(t, err)
					createResp := &InfoResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, managerCreateProcess(), createInput, createResp))

					input, err := json.Marshal(LogStreamInput{ID: createResp.Info.ID, Count: 100})
					require.NoError(t, err)
					resp := &LogStreamResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, remoteGetLogStream(), input, resp))

					assert.True(t, resp.Successful())
				},
			} {
				t.Run(testName, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), testutil.TestTimeout)
					defer cancel()

					port := testutil.GetPortNumber()
					c := mockCLIContext(remoteType, port)
					manager, err := jasper.NewSynchronizedManager(false)
					require.NoError(t, err)
					closeService := makeService(ctx, t, port, manager)
					require.NoError(t, err)
					defer func() {
						assert.NoError(t, closeService())
					}()

					testCase(ctx, t, c)
				})
			}
		})
	}
}
