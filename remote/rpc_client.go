package remote

import (
	"context"
	"io"
	"net"

	"github.com/evergreen-ci/certdepot"
	"github.com/mongodb/grip"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/remote/internal"
	"github.com/mongodb/jasper/util"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type rpcClient struct {
	client       internal.JasperProcessManagerClient
	clientCloser util.CloseFunc
}

// NewRPCClient creates a connection to the RPC service with the specified
// address addr. If creds is non-nil, the credentials will be used to establish
// a secure TLS connection with the service; otherwise, it will establish an
// insecure connection. The caller is responsible for closing the connection
// using the returned jasper.CloseFunc.
func NewRPCClient(ctx context.Context, addr net.Addr, creds *certdepot.Credentials) (Manager, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	}
	if creds != nil {
		tlsConf, err := creds.Resolve()
		if err != nil {
			return nil, errors.Wrap(err, "resolving credentials into TLS config")
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, addr.String(), opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "establishing connection to '%s' service at address '%s'", addr.Network(), addr.String())
	}

	return newRPCClient(conn), nil
}

// NewRPCClientWithFile is the same as NewRPCClient but the credentials will
// be read from the file given by filePath if the filePath is non-empty. The
// credentials file should contain the JSON-encoded bytes from
// (*certdepot.Credentials).Export().
func NewRPCClientWithFile(ctx context.Context, addr net.Addr, filePath string) (Manager, error) {
	var creds *certdepot.Credentials
	if filePath != "" {
		var err error
		creds, err = certdepot.NewCredentialsFromFile(filePath)
		if err != nil {
			return nil, errors.Wrap(err, "getting credentials from file")
		}
	}

	return NewRPCClient(ctx, addr, creds)
}

// newRPCClient is a constructor for an RPC client.
func newRPCClient(cc *grpc.ClientConn) Manager {
	return &rpcClient{
		client:       internal.NewJasperProcessManagerClient(cc),
		clientCloser: cc.Close,
	}
}

func (c *rpcClient) ID() string {
	resp, err := c.client.ID(context.Background(), &emptypb.Empty{})
	if err != nil {
		return ""
	}
	return resp.Value
}

func (c *rpcClient) CreateProcess(ctx context.Context, opts *options.Create) (jasper.Process, error) {
	convertedOpts, err := internal.ConvertCreateOptions(opts)
	if err != nil {
		return nil, errors.Wrap(err, "converting create options")
	}
	proc, err := c.client.Create(ctx, convertedOpts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &rpcProcess{client: c.client, info: proc}, nil
}

func (c *rpcClient) CreateCommand(ctx context.Context) *jasper.Command {
	return jasper.NewCommand().ProcConstructor(c.CreateProcess)
}

func (c *rpcClient) Register(ctx context.Context, proc jasper.Process) error {
	return errors.New("cannot register local processes on remote process managers")
}

func (c *rpcClient) List(ctx context.Context, f options.Filter) ([]jasper.Process, error) {
	procs, err := c.client.List(ctx, internal.ConvertFilter(f))
	if err != nil {
		return nil, errors.Wrap(err, "getting streaming client")
	}

	out := []jasper.Process{}
	for {
		info, err := procs.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "receiving process list")
		}

		out = append(out, &rpcProcess{
			client: c.client,
			info:   info,
		})
	}

	return out, nil
}

func (c *rpcClient) Group(ctx context.Context, name string) ([]jasper.Process, error) {
	procs, err := c.client.Group(ctx, &internal.TagName{Value: name})
	if err != nil {
		return nil, errors.Wrap(err, "getting streaming client")
	}

	out := []jasper.Process{}
	for {
		info, err := procs.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "receiving process list")
		}

		out = append(out, &rpcProcess{
			client: c.client,
			info:   info,
		})
	}

	return out, nil
}

func (c *rpcClient) Get(ctx context.Context, name string) (jasper.Process, error) {
	info, err := c.client.Get(ctx, &internal.JasperProcessID{Value: name})
	if err != nil {
		return nil, errors.Wrap(err, "finding process")
	}

	return &rpcProcess{client: c.client, info: info}, nil
}

func (c *rpcClient) Clear(ctx context.Context) {
	_, _ = c.client.Clear(ctx, &emptypb.Empty{})
}

func (c *rpcClient) Close(ctx context.Context) error {
	resp, err := c.client.Close(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.WithStack(err)
	}
	if !resp.Success {
		return errors.New(resp.Text)
	}

	return nil
}

func (c *rpcClient) Status(ctx context.Context) (string, bool, error) {
	resp, err := c.client.Status(ctx, &emptypb.Empty{})
	if err != nil {
		return "", false, errors.WithStack(err)
	}
	return resp.HostId, resp.Active, nil
}

func (c *rpcClient) CloseConnection() error {
	return c.clientCloser()
}

func (c *rpcClient) ConfigureCache(ctx context.Context, opts options.Cache) error {
	resp, err := c.client.ConfigureCache(ctx, internal.ConvertCacheOptions(opts))
	if err != nil {
		return errors.WithStack(err)
	}
	if !resp.Success {
		return errors.New(resp.Text)
	}

	return nil
}

func (c *rpcClient) DownloadFile(ctx context.Context, opts options.Download) error {
	resp, err := c.client.DownloadFile(ctx, internal.ConvertDownloadOptions(opts))
	if err != nil {
		return errors.WithStack(err)
	}

	if !resp.Success {
		return errors.New(resp.Text)
	}

	return nil
}

func (c *rpcClient) DownloadMongoDB(ctx context.Context, opts options.MongoDBDownload) error {
	resp, err := c.client.DownloadMongoDB(ctx, internal.ConvertMongoDBDownloadOptions(opts))
	if err != nil {
		return errors.WithStack(err)
	}
	if !resp.Success {
		return errors.New(resp.Text)
	}

	return nil
}

func (c *rpcClient) GetLogStream(ctx context.Context, id string, count int) (jasper.LogStream, error) {
	stream, err := c.client.GetLogStream(ctx, &internal.LogRequest{
		Id:    &internal.JasperProcessID{Value: id},
		Count: int64(count),
	})
	if err != nil {
		return jasper.LogStream{}, errors.WithStack(err)
	}
	return stream.Export(), nil
}

func (c *rpcClient) GetBuildloggerURLs(ctx context.Context, id string) ([]string, error) {
	resp, err := c.client.GetBuildloggerURLs(ctx, &internal.JasperProcessID{Value: id})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp.Urls, nil
}

func (c *rpcClient) SignalEvent(ctx context.Context, name string) error {
	resp, err := c.client.SignalEvent(ctx, &internal.EventName{Value: name})
	if err != nil {
		return errors.WithStack(err)
	}
	if !resp.Success {
		return errors.New(resp.Text)
	}

	return nil
}

func (c *rpcClient) WriteFile(ctx context.Context, jopts options.WriteFile) error {
	stream, err := c.client.WriteFile(ctx)
	if err != nil {
		return errors.Wrap(err, "getting client stream")
	}

	sendOpts := func(jopts options.WriteFile) error {
		opts := internal.ConvertWriteFileOptions(jopts)
		return stream.Send(opts)
	}

	if err = jopts.WriteBufferedContent(sendOpts); err != nil {
		catcher := grip.NewBasicCatcher()
		catcher.Wrapf(err, "reading from content source")
		catcher.Wrapf(stream.CloseSend(), "closing send stream after error during read: %s", err.Error())
		return catcher.Resolve()
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return errors.WithStack(err)
	}

	if !resp.Success {
		return errors.New(resp.Text)
	}

	return nil
}

func (c *rpcClient) SendMessages(ctx context.Context, lp options.LoggingPayload) error {
	resp, err := c.client.SendMessages(ctx, internal.ConvertLoggingPayload(lp))
	if err != nil {
		return errors.WithStack(err)
	}

	if !resp.Success {
		return errors.New(resp.Text)
	}

	return nil
}

func (c *rpcClient) LoggingCache(ctx context.Context) jasper.LoggingCache {
	return &rpcLoggingCache{ctx: ctx, client: c.client}
}
