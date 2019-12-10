package remote

import (
	"context"
	"runtime"
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

func TestWireManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	factory := func(ctx context.Context, t *testing.T) Manager {
		mngr, err := jasper.NewSynchronizedManager(false)
		require.NoError(t, err)

		client, err := makeTestMDBServiceAndClient(ctx, mngr)
		require.NoError(t, err)
		return client
	}

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
			for _, test := range AddBasicClientTests(
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
						opts := testutil.TrueCreateOpts()
						modify.Options(opts)
						proc, err := client.CreateProcess(ctx, opts)
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
						opts := testutil.SleepCreateOpts(100)
						modify.Options(opts)
						proc, err := client.CreateProcess(ctx, opts)
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
			) {
				t.Run(test.Name, func(t *testing.T) {
					tctx, cancel := context.WithTimeout(ctx, testutil.RPCTestTimeout)
					defer cancel()
					test.Case(tctx, t, factory(tctx, t))
				})
			}
		})
	}
}
