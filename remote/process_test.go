package remote

import (
	"context"
	"fmt"
	"syscall"
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

	testCases := append(jasper.ProcessTests(), []jasper.ProcessTestCase{
		{
			Name: "RegisterTriggerFails",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makeProc jasper.ProcessConstructor) {
				opts.Args = testutil.SleepCreateOpts(3).Args
				proc, err := makeProc(ctx, opts)
				require.NoError(t, err)
				assert.Error(t, proc.RegisterTrigger(ctx, func(jasper.ProcessInfo) {}))
			},
		},
		{
			Name: "RegisterSignalTriggerFails",
			Case: func(ctx context.Context, t *testing.T, opts *options.Create, makeProc jasper.ProcessConstructor) {
				opts.Args = testutil.SleepCreateOpts(3).Args
				proc, err := makeProc(ctx, opts)
				require.NoError(t, err)
				assert.Error(t, proc.RegisterSignalTrigger(ctx, func(jasper.ProcessInfo, syscall.Signal) bool {
					return false
				}))
			},
		},
	}...)

	for procName, makeProc := range map[string]jasper.ProcessConstructor{
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
		t.Run(procName, func(t *testing.T) {
			for optsTestName, modifyOpts := range map[string]testutil.OptsModify{
				"BlockingProcess": func(opts *options.Create) *options.Create {
					opts.Implementation = options.ProcessImplementationBlocking
					return opts
				},
				"Basic": func(opts *options.Create) *options.Create {
					opts.Implementation = options.ProcessImplementationBasic
					return opts
				},
			} {
				t.Run(optsTestName, func(t *testing.T) {
					for _, testCase := range testCases {
						t.Run(testCase.Name, func(t *testing.T) {
							tctx, tcancel := context.WithTimeout(ctx, testutil.RPCTestTimeout)
							defer tcancel()

							opts := &options.Create{Args: []string{"ls"}}
							opts = modifyOpts(opts)
							testCase.Case(tctx, t, opts, makeProc)
						})
					}
				})
			}
		})
	}
}
