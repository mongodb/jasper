package cli

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mongodb/jasper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestCLIManager(t *testing.T) {
	for remoteType, makeService := range map[string]func(ctx context.Context, t *testing.T, manager jasper.Manager) (jasper.CloseFunc, int){
		serviceREST: makeRESTService,
		serviceRPC:  makeRPCService,
	} {
		t.Run(remoteType, func(t *testing.T) {
			for testName, testCase := range map[string]func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string){
				"CreateCommandSucceeds": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					input, err := json.Marshal(CommandInput{
						Commands: [][]string{[]string{"echo", "hello", "world"}},
					})
					require.NoError(t, err)
					resp := &OutcomeResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, managerCreateCommand(), input, resp))
					require.True(t, resp.Successful())
				},
				"GetExistingIDSucceeds": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					input, err := json.Marshal(IDInput{jasperProcID})
					require.NoError(t, err)
					resp := &InfoResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, managerGet(), input, resp))
					require.True(t, resp.Successful())
					assert.Equal(t, jasperProcID, resp.Info.ID)
				},
				"GetNonexistentIDFails": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					input, err := json.Marshal(IDInput{nonexistentID})
					require.NoError(t, err)
					resp := &InfoResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, managerGet(), input, resp))
					require.False(t, resp.Successful())
					require.NotEmpty(t, resp.ErrorMessage())
				},
				"GetEmptyIDFails": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					input, err := json.Marshal(IDInput{""})
					require.NoError(t, err)
					assert.Error(t, execCLICommandInputOutput(t, c, managerGet(), input, &InfoResponse{}))
				},
				"ListValidFilterSucceeds": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					input, err := json.Marshal(FilterInput{jasper.All})
					require.NoError(t, err)
					resp := &InfosResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, managerList(), input, resp))
					require.True(t, resp.Successful())
					assert.Len(t, resp.Infos, 1)
					assert.Equal(t, jasperProcID, resp.Infos[0].ID)
				},
				"ListInvalidFilterFails": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					input, err := json.Marshal(FilterInput{jasper.Filter("foo")})
					require.NoError(t, err)
					assert.Error(t, execCLICommandInputOutput(t, c, managerList(), input, &InfosResponse{}))
				},
				"GroupFindsTaggedProcess": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					tag := "foo"
					require.True(t, tagProcess(t, c, jasperProcID, tag).Successful())

					input, err := json.Marshal(TagInput{Tag: tag})
					require.NoError(t, err)
					resp := &InfosResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, managerGroup(), input, resp))
					require.True(t, resp.Successful())
					require.Len(t, resp.Infos, 1)
					assert.Equal(t, jasperProcID, resp.Infos[0].ID)
				},
				"GroupEmptyTagFails": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					input, err := json.Marshal(TagInput{Tag: ""})
					require.NoError(t, err)
					assert.Error(t, execCLICommandInputOutput(t, c, managerGroup(), input, &InfosResponse{}))
				},
				"GroupNoMatchingTaggedProcessesReturnsEmpty": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					input, err := json.Marshal(TagInput{Tag: "foo"})
					require.NoError(t, err)
					resp := &InfosResponse{}
					require.NoError(t, execCLICommandInputOutput(t, c, managerGroup(), input, resp))
					require.True(t, resp.Successful())
					assert.Len(t, resp.Infos, 0)
				},
				"ClearSucceeds": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					resp := &OutcomeResponse{}
					require.NoError(t, execCLICommandOutput(t, c, managerClear(), resp))
					assert.True(t, resp.Successful())
				},
				"CloseSucceeds": func(ctx context.Context, t *testing.T, c *cli.Context, jasperProcID string) {
					resp := &OutcomeResponse{}
					require.NoError(t, execCLICommandOutput(t, c, managerClose(), resp))
					assert.True(t, resp.Successful())
				},
			} {
				t.Run(testName, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
					defer cancel()
					manager, err := jasper.NewLocalManager(false)
					require.NoError(t, err)
					closeService, port := makeService(ctx, t, manager)
					c := mockCLIContext(remoteType, port)
					require.NoError(t, err)
					defer func() {
						assert.NoError(t, closeService())
					}()

					cmdCtx := cli.NewContext(nil, nil, c)
					resp := &InfoResponse{}
					input, err := json.Marshal(trueCreateOpts())
					require.NoError(t, err)
					require.NoError(t, execCLICommandInputOutput(t, cmdCtx, managerCreateProcess(), input, resp))
					require.True(t, resp.Successful())
					require.NotZero(t, resp.Info.ID)

					testCase(ctx, t, cmdCtx, resp.Info.ID)
				})
			}
		})
	}
}
