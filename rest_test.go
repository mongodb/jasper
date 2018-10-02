package jasper

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type neverJSON struct{}

func (n *neverJSON) MarshalJSON() ([]byte, error)  { return nil, errors.New("always error") }
func (n *neverJSON) UnmarshalJSON(in []byte) error { return errors.New("always error") }
func (n *neverJSON) Read(p []byte) (int, error)    { return 0, errors.New("always error") }
func (n *neverJSON) Close() error                  { return errors.New("always error") }

func TestRestService(t *testing.T) {
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
			_, err := makeBody(&neverJSON{})
			assert.Error(t, err)
			assert.Error(t, handleError(&http.Response{Body: &neverJSON{}, StatusCode: http.StatusTeapot}))
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

			_, err = client.getProcessInfo(ctx, "foo")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem building request")

			err = client.DownloadFile(ctx, "foo", "bar")
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

			_, err = client.getProcessInfo(ctx, "foo")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "problem making request")

			err = client.DownloadFile(ctx, "foo", "bar")
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
		"CreateProcessEndpointErrorsWithMalformedData": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			body, err := makeBody(map[string]int{"tags": 42})
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "", ioutil.NopCloser(body))
			require.NoError(t, err)
			rw := httptest.NewRecorder()
			srv.createProcess(rw, req)
			assert.Equal(t, http.StatusBadRequest, rw.Code)
		},
		"CreateFailPropogatesErrors": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			srv.manager = &MockManager{
				FailCreate: true,
			}
			proc, err := client.Create(ctx, trueCreateOpts())
			assert.Error(t, err)
			assert.Nil(t, proc)
			assert.Contains(t, err.Error(), "problem submitting request")
		},
		"CreateFailsForTriggerReasons": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			srv.manager = &MockManager{
				FailCreate: false,
				Process: &MockProcess{
					FailRegisterTrigger: true,
				},
			}
			proc, err := client.Create(ctx, trueCreateOpts())
			assert.Error(t, err)
			assert.Nil(t, proc)
			assert.Contains(t, err.Error(), "problem managing resources")
		},
		"InvalidFilterReturnsError": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			req, err := http.NewRequest(http.MethodGet, client.getURL("/list/%s", "foo"), nil)
			require.NoError(t, err)
			out, err := client.getListOfProcesses(req)
			assert.Error(t, err)
			assert.Nil(t, out)
		},
		"WaitForProcessThatDoesNotExistShouldError": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			proc := &restProcess{
				client: client,
				id:     "foo",
			}

			assert.Error(t, proc.Wait(ctx))
		},
		"SignalProcessThatDoesNotExistShouldError": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			proc := &restProcess{
				client: client,
				id:     "foo",
			}

			assert.Error(t, proc.Signal(ctx, syscall.SIGTERM))
		},
		"SignalErrorsWithInvalidSyscall": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			proc, err := client.Create(ctx, trueCreateOpts())
			require.NoError(t, err)

			assert.Error(t, proc.Signal(ctx, syscall.Signal(-1)))
		},
		"GetProcessWhenInvalid": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			srv.manager = &MockManager{
				Process: &MockProcess{},
			}

			_, err := client.Get(ctx, "foo")
			assert.Error(t, err)
		},
		"MetricsErrorForInvalidProcess": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			req, err := http.NewRequest(http.MethodGet, client.getURL("/process/%s/metrics", "foo"), nil)
			require.NoError(t, err)
			req = req.WithContext(ctx)
			res, err := httpClient.Do(req)
			require.NoError(t, err)

			assert.Equal(t, http.StatusNotFound, res.StatusCode)
		},
		"MetricsPopulatedForValidProcess": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			srv.manager = &MockManager{
				Process: &MockProcess{
					ProcID: "foo",
					ProcInfo: ProcessInfo{
						PID: os.Getpid(),
					},
				},
			}

			req, err := http.NewRequest(http.MethodGet, client.getURL("/process/%s/metrics", "foo"), nil)
			require.NoError(t, err)
			req = req.WithContext(ctx)
			res, err := httpClient.Do(req)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)
		},
		"AddTagsWithNoTagsSpecifiedShouldError": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			srv.manager = &MockManager{}

			req, err := http.NewRequest(http.MethodPost, client.getURL("/process/%s/tags", "foo"), nil)
			require.NoError(t, err)
			req = req.WithContext(ctx)
			res, err := httpClient.Do(req)
			require.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		},
		"SignalInPassingCase": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			srv.manager = &MockManager{
				Process: &MockProcess{},
			}
			proc := &restProcess{
				client: client,
				id:     "foo",
			}

			err := proc.Signal(ctx, syscall.SIGTERM)
			assert.NoError(t, err)

		},
		"SignalFailsToParsePid": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			req, err := http.NewRequest(http.MethodPost, client.getURL("/process/%s/signal/f", "foo"), nil)
			require.NoError(t, err)
			rw := httptest.NewRecorder()

			srv.signalProcess(rw, req)
			assert.Equal(t, http.StatusBadRequest, rw.Code)
		},
		"DownloadFileCreatesResource": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			file, err := ioutil.TempFile("build", "out.txt")
			require.NoError(t, err)
			defer os.Remove(file.Name())
			assert.NoError(t, client.DownloadFile(ctx, "https://google.com", file.Name()))

			info, err := os.Stat(file.Name())
			assert.NoError(t, err)
			assert.NotEqual(t, 0, info.Size())
		},
		"DownloadFileFailsWithBadURL": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			err := client.DownloadFile(ctx, "", "")
			assert.Error(t, err)
		},
		"ServiceDownloadFileFailsWithBadInfo": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			body, err := makeBody(struct {
				URL int `json:"url"`
			}{URL: 0})
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, client.getURL("/Download"), body)
			require.NoError(t, err)
			rw := httptest.NewRecorder()
			srv.downloadFile(rw, req)
			assert.Equal(t, http.StatusBadRequest, rw.Code)
		},
		"DownloadFileFailsForNonexistentURL": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			file, err := ioutil.TempFile("build", "out.txt")
			require.NoError(t, err)
			defer os.Remove(file.Name())
			assert.Error(t, client.DownloadFile(ctx, "https://google.com/foo", file.Name()))
		},
		"DownloadFileFailsForInsufficientPermissions": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {
			if os.Geteuid() == 0 {
				t.Skip("cannot test download permissions as root")
			} else if runtime.GOOS == "windows" {
				t.Skip("cannot test download permissions on windows")
			}
			assert.Error(t, client.DownloadFile(ctx, "https://google.com", "/foo/bar"))
		},
		// "": func(ctx context.Context, t *testing.T, srv *Service, client *restClient) {},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
			defer cancel()

			srv, port := makeAndStartService(ctx, httpClient)
			require.NotNil(t, srv)

			client := &restClient{
				prefix: fmt.Sprintf("http://localhost:%d/jasper/v1", port),
				client: httpClient,
			}

			test(ctx, t, srv, client)
		})
	}
}
