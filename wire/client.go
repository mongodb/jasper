package wire

import (
	"context"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

type client struct {
	conn      net.Conn
	namespace string
}

const (
	namespace = "jasper.$cmd"
)

// NewClient returns a remote client for connection to a MongoDB wire protocol
// service.
func NewClient(ctx context.Context, addr net.Addr) (jasper.RemoteClient, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s service at address %s", addr.Network(), addr.String())
	}
	// <namespace>.$cmd format is required to indicate the OP_QUERY should be
	// converted to OP_COMMAND
	return &client{conn: conn, namespace: namespace}, nil
}

// CloseConnection closes the client connection. Callers are expected to call
// this when finished with the client.
func (c *client) CloseConnection() error {
	return c.conn.Close()
}

// TODO: implement
func (c *client) ConfigureCache(ctx context.Context, opts options.Cache) error {
	return nil
}

func (c *client) DownloadFile(ctx context.Context, info options.Download) error {
	return nil
}

func (c *client) DownloadMongoDB(ctx context.Context, opts options.MongoDBDownload) error {
	return nil
}

func (c *client) GetLogStream(ctx context.Context, id string, count int) (jasper.LogStream, error) {
	return jasper.LogStream{}, nil
}

func (c *client) GetBuildloggerURLs(ctx context.Context, id string) ([]string, error) {
	return []string{}, nil
}

func (c *client) SignalEvent(ctx context.Context, name string) error {
	return nil
}

func (c *client) WriteFile(ctx context.Context, info options.WriteFile) error {
	return nil
}

func (c *client) ID() string {
	req, err := makeIDRequest().Message()
	if err != nil {
		grip.Warning(message.WrapError(err, "could not create request"))
		return ""
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msg, err := doRequest(ctx, c.conn, req)
	if err != nil {
		grip.Warning(message.WrapError(err, "failed during request"))
		return ""
	}
	resp, err := ExtractIDResponse(msg)
	if err != nil {
		grip.Warning(message.WrapError(err, "problem with received response"))
		return ""
	}
	return resp.ID
}

func (c *client) CreateProcess(ctx context.Context, opts *options.Create) (jasper.Process, error) {
	req, err := makeCreateProcessRequest(*opts).Message()
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}
	msg, err := doRequest(ctx, c.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed during request")
	}
	resp, err := ExtractInfoResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "problem with received response")
	}
	return &process{info: resp.Info, conn: c.conn}, nil
}

func (c *client) CreateCommand(ctx context.Context) *jasper.Command {
	return jasper.NewCommand().ProcConstructor(c.CreateProcess)
}

func (c *client) Register(ctx context.Context, proc jasper.Process) error {
	return errors.New("cannot register extant processes on remote process managers")
}

func (c *client) List(ctx context.Context, f options.Filter) ([]jasper.Process, error) {
	req, err := makeListRequest(f).Message()
	if err != nil {
		return nil, errors.Wrap(err, "could  not make request")
	}
	msg, err := doRequest(ctx, c.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed during request")
	}
	resp, err := ExtractInfosResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "problem with received response")
	}
	infos := resp.Infos
	procs := make([]jasper.Process, 0, len(infos))
	for _, info := range infos {
		procs = append(procs, &process{info: info, conn: c.conn})
	}
	return procs, nil
}

func (c *client) Group(ctx context.Context, tag string) ([]jasper.Process, error) {
	req, err := makeGroupRequest(tag).Message()
	if err != nil {
		return nil, errors.Wrap(err, "could  not make request")
	}
	msg, err := doRequest(ctx, c.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed during request")
	}
	resp, err := ExtractInfosResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "problem with received response")
	}
	infos := resp.Infos
	procs := make([]jasper.Process, 0, len(infos))
	for _, info := range infos {
		procs = append(procs, &process{info: info, conn: c.conn})
	}
	return procs, nil
}

func (c *client) Get(ctx context.Context, id string) (jasper.Process, error) {
	req, err := makeGetRequest(id).Message()
	if err != nil {
		return nil, errors.Wrap(err, "could  not make request")
	}
	msg, err := doRequest(ctx, c.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed during request")
	}
	resp, err := ExtractInfoResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "problem with received response")
	}
	info := resp.Info
	return &process{info: info, conn: c.conn}, nil
}

func (c *client) Clear(ctx context.Context) {
	req, err := makeClearRequest().Message()
	if err != nil {
		grip.Warning(message.WrapError(err, "could not create request"))
		return
	}
	msg, err := doRequest(ctx, c.conn, req)
	if err != nil {
		grip.Warning(message.WrapError(err, "failed during request"))
		return
	}
	if _, err := ExtractInfoResponse(msg); err != nil {
		grip.Warning(message.WrapError(err, "problem with received response"))
		return
	}
}

func (c *client) Close(ctx context.Context) error {
	req, err := makeCloseRequest().Message()
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}
	msg, err := doRequest(ctx, c.conn, req)
	if err != nil {
		return errors.Wrap(err, "failed during request")
	}
	if _, err := ExtractErrorResponse(msg); err != nil {
		return errors.Wrap(err, "problem with received response")
	}
	return nil
}

type process struct {
	info jasper.ProcessInfo
	conn net.Conn
}

func (p *process) ID() string { return p.info.ID }

func (p *process) Info(ctx context.Context) jasper.ProcessInfo {
	if p.info.Complete {
		return p.info
	}

	req, err := makeInfoRequest(p.ID()).Message()
	if err != nil {
		grip.Warning(message.WrapErrorf(err, "failed to get process info for %s", p.ID()))
		return jasper.ProcessInfo{}
	}
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		grip.Warning(message.WrapErrorf(err, "failed to get process info for %s", p.ID()))
		return jasper.ProcessInfo{}
	}
	resp, err := ExtractInfoResponse(msg)
	if err != nil {
		grip.Warning(message.WrapErrorf(err, "failed to get process info for %s", p.ID()))
		return jasper.ProcessInfo{}
	}
	p.info = resp.Info
	return p.info
}

func (p *process) Running(ctx context.Context) bool {
	if p.info.Complete {
		return false
	}

	req, err := makeRunningRequest(p.ID()).Message()
	if err != nil {
		grip.Warning(message.WrapErrorf(err, "failed to get process info for %s", p.ID()))
		return false
	}
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		grip.Warning(message.WrapErrorf(err, "failed to get process info for %s", p.ID()))
		return false
	}
	resp, err := ExtractRunningResponse(msg)
	if err != nil {
		grip.Warning(message.WrapErrorf(err, "failed to get process info for %s", p.ID()))
		return false
	}
	return resp.Running
}

func (p *process) Complete(ctx context.Context) bool {
	if p.info.Complete {
		return true
	}

	req, err := makeCompleteRequest(p.ID()).Message()
	if err != nil {
		grip.Warning(message.WrapErrorf(err, "failed to get process info for %s", p.ID()))
		return false
	}
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		grip.Warning(message.WrapErrorf(err, "failed to get process info for %s", p.ID()))
		return false
	}
	resp, err := ExtractCompleteResponse(msg)
	if err != nil {
		grip.Warning(message.WrapErrorf(err, "failed to get process info for %s", p.ID()))
		return false
	}
	return resp.Complete
}

func (p *process) Signal(ctx context.Context, sig syscall.Signal) error {
	req, err := makeSignalRequest(p.ID(), int(sig)).Message()
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		return errors.Wrap(err, "failed during request")
	}
	if _, err := ExtractErrorResponse(msg); err != nil {
		return err
	}
	return nil
}

func (p *process) Wait(ctx context.Context) (int, error) {
	req, err := makeWaitRequest(p.ID()).Message()
	if err != nil {
		return -1, errors.Wrap(err, "could not create request")
	}
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		return -1, errors.Wrap(err, "failed during request")
	}
	resp, err := ExtractWaitResponse(msg)
	return resp.ExitCode, err
}

func (p *process) Respawn(ctx context.Context) (jasper.Process, error) {
	req, err := makeRespawnRequest(p.ID()).Message()
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed during request")
	}
	resp, err := ExtractInfoResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "problem with received response")
	}
	return &process{info: resp.Info, conn: p.conn}, nil
}

func (p *process) RegisterTrigger(ctx context.Context, t jasper.ProcessTrigger) error {
	return errors.New("cannot register triggers on remote processes")
}

func (p *process) RegisterSignalTrigger(ctx context.Context, t jasper.SignalTrigger) error {
	return errors.New("cannot register signal triggers on remote processes")
}

func (p *process) RegisterSignalTriggerID(ctx context.Context, sigID jasper.SignalTriggerID) error {
	req, err := makeRegisterSignalTriggerIDRequest(p.ID(), sigID).Message()
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		return errors.Wrap(err, "failed during request")
	}
	if _, err := ExtractErrorResponse(msg); err != nil {
		return errors.Wrap(err, "problem with received response")
	}
	return nil
}

func (p *process) Tag(tag string) {
	req, err := makeTagRequest(p.ID(), tag).Message()
	if err != nil {
		grip.Warningf("failed to tag process %s with tag %s", p.ID(), tag)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		grip.Warningf("failed to tag process %s with tag %s", p.ID(), tag)
		return
	}
	if _, err := ExtractErrorResponse(msg); err != nil {
		grip.Warningf("failed to tag process %s with tag %s", p.ID(), tag)
		return
	}
}

func (p *process) GetTags() []string {
	req, err := makeGetTagsRequest(p.ID()).Message()
	if err != nil {
		grip.Warningf("failed to get tags for %s", p.ID())
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		grip.Warningf("failed to get tags for %s", p.ID())
		return nil
	}
	resp, err := ExtractGetTagsResponse(msg)
	if err != nil {
		grip.Warningf("failed to get tags for %s", p.ID())
		return nil
	}
	return resp.Tags
}

func (p *process) ResetTags() {
	req, err := makeResetTagsRequest(p.ID()).Message()
	if err != nil {
		grip.Warningf("failed to reset tags for process %s", p.ID())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msg, err := doRequest(ctx, p.conn, req)
	if err != nil {
		grip.Warningf("failed to reset tags for process %s", p.ID())
		return
	}
	if _, err := ExtractErrorResponse(msg); err != nil {
		grip.Warningf("failed to reset tags for process %s", p.ID())
		return
	}
}

// doRequest sends the given request and reads the response.
func doRequest(ctx context.Context, rw io.ReadWriter, req mongowire.Message) (mongowire.Message, error) {
	const requestMaxTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(ctx, requestMaxTimeout)
	defer cancel()
	if err := mongowire.SendMessage(ctx, req, rw); err != nil {
		return nil, errors.Wrap(err, "problem sending request")
	}
	msg, err := mongowire.ReadMessage(ctx, rw)
	if err != nil {
		return nil, errors.Wrap(err, "problem receiving response")
	}
	return msg, nil
}
