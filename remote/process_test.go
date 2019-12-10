package remote

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessImplementations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient := testutil.GetHTTPClient()
	defer testutil.PutHTTPClient(httpClient)

	for cname, makeProc := range map[string]jasper.ProcessConstructor{
		"REST": func(ctx context.Context, opts *options.Create) (jasper.Process, error) {
			_, port, err := startRESTService(ctx, httpClient)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			client := &restClient{
				prefix: fmt.Sprintf("http://localhost:%d/jasper/v1", port),
				client: httpClient,
			}

			return client.CreateProcess(ctx, opts)
		},
		"MDB": func(ctx context.Context, opts *options.Create) (jasper.Process, error) {
			mngr, err := jasper.NewSynchronizedManager(false)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			client, err := makeTestMDBServiceAndClient(ctx, mngr)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return client.CreateProcess(ctx, opts)
		},
		"RPC/TLS": func(ctx context.Context, opts *options.Create) (jasper.Process, error) {
			mngr, err := jasper.NewSynchronizedManager(false)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			client, err := makeTLSRPCServiceAndClient(ctx, mngr)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return client.CreateProcess(ctx, opts)
		},
		"RPC/Insecure": func(ctx context.Context, opts *options.Create) (jasper.Process, error) {
			mngr, err := jasper.NewSynchronizedManager(false)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			client, err := makeInsecureRPCServiceAndClient(ctx, mngr)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return client.CreateProcess(ctx, opts)
		},
	} {
		t.Run(cname, func(t *testing.T) {
			for _, modify := range []struct {
				Name    string
				Options testutil.OptsModify
			}{
				{
					Name: "Blocking",
					Options: func(opts *options.Create) {
						opts.Implementation = options.ProcessImplementationBlocking
					},
				},
				{
					Name: "Basic",
					Options: func(opts *options.Create) {
						opts.Implementation = options.ProcessImplementationBasic
					},
				},
				{
					Name:    "Default",
					Options: func(opts *options.Create) {},
				},
			} {
				t.Run(modify.Name, func(t *testing.T) {
					for _, test := range AddBasicProcessTests(
						ProcessTestCase{
							Name:       "CompleteReturnsFalseForProcessThatDoesntExist",
							ShouldSkip: !strings.Contains(cname, "RPC"),
							Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
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
						},
						ProcessTestCase{
							Name:       "RunningReturnsFalseForProcessThatDoesntExist",
							ShouldSkip: !strings.Contains(cname, "RPC"),
							Case: func(ctx context.Context, t *testing.T, opts *options.Create, makep jasper.ProcessConstructor) {
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
						},
					) {
						if test.ShouldSkip {
							continue
						}
						t.Run(test.Name, func(t *testing.T) {
							tctx, cancel := context.WithTimeout(ctx, testutil.RPCTestTimeout)
							defer cancel()

							opts := &options.Create{Args: []string{"ls"}}
							modify.Options(opts)
							test.Case(tctx, t, opts, makeProc)
						})
					}
				})
			}
		})
	}
}
