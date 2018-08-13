package jasper

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestService(t *testing.T) {
	srvPort := 4000
	httpClient := &http.Client{}

	for name, test := range map[string]func(context.Context, *testing.T, *Service, *restClient){
		"VerifyFixtures": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			assert.NotNil(t, srv)
			assert.NotNil(t, client)
			assert.NotNil(t, srv.manager)
			assert.NotNil(t, client.client)
			assert.NotZero(t, client.prefix)

			// no good other place to put this assertion
			// about the constructor
			newm := NewManagerService(&basicProcessManager{})
			assert.IsType(t, &localProcessManager{}, srv.manager)
			assert.IsType(t, &localProcessManager{}, newm.manager)

			// similarly about helper functions
			client.prefix = ""
			assert.Equal(t, "/foo", client.getURL("foo"))
		},
		"EmptyCreateOpts": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			proc, err := client.Create(ctx, &CreateOptions{})
			assert.Error(t, err)
			assert.Nil(t, proc)
		},
		"WithOnlyTimeoutValue": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			proc, err := client.Create(ctx, &CreateOptions{Args: []string{"ls"}, TimeoutSecs: 300})
			assert.NoError(t, err)
			assert.NotNil(t, proc)
		},
		"ListErrorsWithInvalidFilter": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			list, err := client.List(ctx, "foo")
			assert.Error(t, err)
			assert.Nil(t, list)
		},
		"RegisterAlwaysErrors": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			proc, err := newBasicProcess(ctx, trueCreateOpts())
			require.NoError(t, err)

			assert.Error(t, client.Register(ctx, nil))
			assert.Error(t, client.Register(nil, nil))
			assert.Error(t, client.Register(ctx, proc))
			assert.Error(t, client.Register(nil, proc))
		},
		"ClientMethodsErrorWithBadUrl": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			client.prefix = strings.Replace(client.prefix, "http://", "://", 1)

			_, err := client.List(ctx, All)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem building request")

			_, err = client.Create(ctx, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem building request")

			_, err = client.Group(ctx, "foo")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem building request")

			_, err = client.Get(ctx, "foo")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem building request")

			err = client.Close(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem building request")
		},
		"ClientRequestsFailWithMalformedURL": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			client.prefix = strings.Replace(client.prefix, "http://", "http;//", 1)

			_, err := client.List(ctx, All)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem making request")

			_, err = client.Group(ctx, "foo")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem making request")

			_, err = client.Create(ctx, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem making request")

			_, err = client.Get(ctx, "foo")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem making request")

			err = client.Close(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem making request")
		},
		"ProcessMethodsWithBadUrl": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			client.prefix = strings.Replace(client.prefix, "http://", "://", 1)

			proc := &restProcess{
				client: client,
				id:     "foo",
			}

			err := proc.Signal(ctx, syscall.SIGTERM)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem building request")

			err = proc.Wait(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem building request")

			proc.Tag("a")

			out := proc.GetTags()
			assert.Nil(t, out)

			proc.ResetTags()
		},
		"ProcessRequestsFailWithBadURL": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {

			client.prefix = strings.Replace(client.prefix, "http://", "http;//", 1)

			proc := &restProcess{
				client: client,
				id:     "foo",
			}

			err := proc.Signal(ctx, syscall.SIGTERM)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem making request")

			err = proc.Wait(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem making request")

			proc.Tag("a")

			out := proc.GetTags()
			assert.Nil(t, out)

			proc.ResetTags()
		},
		"CheckSafetyOfTagMethodsForBrokenTasks": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			proc := &restProcess{
				client: client,
				id:     "foo",
			}

			proc.Tag("a")

			out := proc.GetTags()
			assert.Nil(t, out)

			proc.ResetTags()
		},
		"SignalFailsForTaskThatDoesNotExist": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			proc := &restProcess{
				client: client,
				id:     "foo",
			}

			err := proc.Signal(ctx, syscall.SIGTERM)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no process")

		},
		// "": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			srv := NewManagerService(NewLocalManager())
			app := srv.App()
			srvPort++
			app.SetPrefix("jasper")
			require.NoError(t, app.SetPort(srvPort))

			go func() {
				app.Run(ctx)
			}()

			time.Sleep(10 * time.Millisecond)
			client := &restClient{
				prefix: fmt.Sprintf("http://localhost:%d/jasper/v1", srvPort),
				client: httpClient,
			}

			test(ctx, t, srv, client)
		})
	}
}
