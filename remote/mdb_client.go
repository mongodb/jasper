package remote

import (
	"context"
	"net"
	"time"

	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/evergreen-ci/mrpc/shell"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

type mdbClient struct {
	conn      net.Conn
	namespace string
	timeout   time.Duration
}

const (
	namespace = "jasper.$cmd"
)

// NewMDBClient returns a remote client for connection to a MongoDB wire protocol
// service. reqTimeout specifies the timeout for a request, or uses a default
// timeout if zero.
func NewMDBClient(ctx context.Context, addr net.Addr, reqTimeout time.Duration) (Manager, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return nil, errors.Wrapf(err, "establishing connection to '%s' service at address '%s'", addr.Network(), addr.String())
	}
	timeout := reqTimeout
	if timeout.Seconds() == 0 {
		timeout = 30 * time.Second
	}
	return &mdbClient{conn: conn, namespace: namespace, timeout: timeout}, nil
}

func (c *mdbClient) ID() string {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, &idRequest{ID: 1})
	if err != nil {
		grip.Warning(message.WrapError(err, "creating request"))
		return ""
	}
	msg, err := c.doRequest(context.Background(), req)
	if err != nil {
		grip.Warning(message.WrapError(err, "making request"))
		return ""
	}
	var resp idResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		grip.Warning(message.WrapError(err, "converting wire message to response"))
		return ""
	}
	if err := resp.SuccessOrError(); err != nil {
		grip.Warning(message.WrapError(err, "response contained error"))
		return ""
	}
	return resp.ID
}

func (c *mdbClient) CreateProcess(ctx context.Context, opts *options.Create) (jasper.Process, error) {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, createProcessRequest{Options: *opts})
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}
	var resp infoResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return nil, errors.Wrap(err, "converting wire message to response")
	}
	if err := resp.SuccessOrError(); err != nil {
		return nil, errors.Wrap(err, "response contained error")
	}
	return &mdbProcess{info: resp.Info, doRequest: c.doRequest}, nil
}

func (c *mdbClient) CreateCommand(ctx context.Context) *jasper.Command {
	return jasper.NewCommand().ProcConstructor(c.CreateProcess)
}

func (c *mdbClient) LoggingCache(ctx context.Context) jasper.LoggingCache {
	return &mdbLoggingCache{
		client: c,
		ctx:    ctx,
	}
}

func (c *mdbClient) SendMessages(ctx context.Context, lp options.LoggingPayload) error {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, &sendMessagesRequest{Payload: lp})
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}

	resp := &shell.ErrorResponse{}
	if err = shell.MessageToResponse(msg, resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}

	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (c *mdbClient) Register(ctx context.Context, proc jasper.Process) error {
	return errors.New("cannot register local processes on remote process managers")
}

func (c *mdbClient) List(ctx context.Context, f options.Filter) ([]jasper.Process, error) {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, listRequest{Filter: f})
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}
	var resp infosResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return nil, errors.Wrap(err, "converting wire message to response")
	}
	if err := resp.SuccessOrError(); err != nil {
		return nil, errors.Wrap(err, "response contained error")
	}
	infos := resp.Infos
	procs := make([]jasper.Process, 0, len(infos))
	for _, info := range infos {
		procs = append(procs, &mdbProcess{info: info, doRequest: c.doRequest})
	}
	return procs, nil
}

func (c *mdbClient) Group(ctx context.Context, tag string) ([]jasper.Process, error) {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, groupRequest{Tag: tag})
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}
	var resp infosResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return nil, errors.Wrap(err, "converting wire message to response")
	}
	if err := resp.SuccessOrError(); err != nil {
		return nil, errors.Wrap(err, "response contained error")
	}
	infos := resp.Infos
	procs := make([]jasper.Process, 0, len(infos))
	for _, info := range infos {
		procs = append(procs, &mdbProcess{info: info, doRequest: c.doRequest})
	}
	return procs, nil
}

func (c *mdbClient) Get(ctx context.Context, id string) (jasper.Process, error) {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, &getProcessRequest{ID: id})
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}
	var resp infoResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return nil, errors.Wrap(err, "converting wire message to response")
	}
	if err := resp.SuccessOrError(); err != nil {
		return nil, errors.Wrap(err, "response contained error")
	}
	info := resp.Info
	return &mdbProcess{info: info, doRequest: c.doRequest}, nil
}

func (c *mdbClient) Clear(ctx context.Context) {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, &clearRequest{Clear: 1})
	if err != nil {
		grip.Warning(message.WrapError(err, "creating request"))
		return
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		grip.Warning(message.WrapError(err, "making request"))
		return
	}
	var resp shell.ErrorResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		grip.Warning(message.WrapError(shell.MessageToResponse(msg, &resp), "converting wire message to response"))
	}
	grip.Warning(message.WrapError(resp.SuccessOrError(), "response contained error"))
}

func (c *mdbClient) Close(ctx context.Context) error {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, &closeRequest{Close: 1})
	if err != nil {
		return errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}
	var resp shell.ErrorResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}
	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (c *mdbClient) WriteFile(ctx context.Context, opts options.WriteFile) error {
	sendOpts := func(opts options.WriteFile) error {
		req, err := shell.RequestToMessage(mongowire.OP_QUERY, writeFileRequest{Options: opts})
		if err != nil {
			return errors.Wrap(err, "creating request")
		}
		msg, err := c.doRequest(ctx, req)
		if err != nil {
			return errors.Wrap(err, "making request")
		}
		var resp shell.ErrorResponse
		if err := shell.MessageToResponse(msg, &resp); err != nil {
			return errors.Wrap(err, "converting wire message to response")
		}
		return errors.Wrap(resp.SuccessOrError(), "response contained error")
	}
	return opts.WriteBufferedContent(sendOpts)
}

// CloseConnection closes the client connection. Callers are expected to call
// this when finished with the client.
func (c *mdbClient) CloseConnection() error {
	return c.conn.Close()
}

func (c *mdbClient) ConfigureCache(ctx context.Context, opts options.Cache) error {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, configureCacheRequest{Options: opts})
	if err != nil {
		return errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}
	var resp shell.ErrorResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}
	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (c *mdbClient) DownloadFile(ctx context.Context, opts options.Download) error {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, downloadFileRequest{Options: opts})
	if err != nil {
		return errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}
	var resp shell.ErrorResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}
	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (c *mdbClient) DownloadMongoDB(ctx context.Context, opts options.MongoDBDownload) error {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, downloadMongoDBRequest{Options: opts})
	if err != nil {
		return errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}
	var resp shell.ErrorResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}
	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (c *mdbClient) GetLogStream(ctx context.Context, id string, count int) (jasper.LogStream, error) {
	r := getLogStreamRequest{}
	r.Params.ID = id
	r.Params.Count = count
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, r)
	if err != nil {
		return jasper.LogStream{}, errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return jasper.LogStream{}, errors.Wrap(err, "making request")
	}
	var resp getLogStreamResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return jasper.LogStream{}, errors.Wrap(err, "converting wire message to response")
	}
	if err := resp.SuccessOrError(); err != nil {
		return jasper.LogStream{}, errors.Wrap(err, "response contained error")
	}
	return resp.LogStream, nil
}

func (c *mdbClient) GetBuildloggerURLs(ctx context.Context, id string) ([]string, error) {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, getBuildloggerURLsRequest{ID: id})
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}
	var resp getBuildloggerURLsResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return nil, errors.Wrap(err, "converting wire message to response")
	}
	if err := resp.SuccessOrError(); err != nil {
		return nil, errors.Wrap(err, "response contained error")
	}
	return resp.URLs, nil
}

func (c *mdbClient) SignalEvent(ctx context.Context, name string) error {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, signalEventRequest{Name: name})
	if err != nil {
		return errors.Wrap(err, "creating request")
	}
	msg, err := c.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}
	var resp shell.ErrorResponse
	if err := shell.MessageToResponse(msg, &resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}
	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

// doRequest sends the given request and reads the response.
func (c *mdbClient) doRequest(ctx context.Context, req mongowire.Message) (mongowire.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	if err := mongowire.SendMessage(ctx, req, c.conn); err != nil {
		return nil, errors.Wrap(err, "sending request message")
	}
	msg, err := mongowire.ReadMessage(ctx, c.conn)
	if err != nil {
		return nil, errors.Wrap(err, "reading response message")
	}
	return msg, nil
}
