package mongowire

import (
	"context"
	"io"
	"net"
	"syscall"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc/mongowire"
)

type client struct {
	conn      net.Conn
	namespace string
}

const namespace = "jasper.$cmd"

// TODO: is implementing all RemoteClient functionality necessary?
func NewClient(addr net.Addr) (jasper.RemoteClient, error) {
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not dial address")
	}
	// <namespace>.$cmd format is required to indicate the OP_QUERY should be
	// converted to OP_COMMAND
	return &client{conn: conn, namespace: namespace}, nil
}

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
		return ""
	}
	if err := mongowire.SendMessage(req, c.conn); err != nil {
		return ""
	}
	msg, err := mongowire.ReadMessage(c.conn)
	if err != nil {
		return ""
	}
	resp, err := ExtractIDResponse(msg)
	if err != nil {
		return ""
	}
	return resp.ID
}

// kim: TODO: implement the remainder
func (c *client) CreateProcess(ctx context.Context, opts *options.Create) (jasper.Process, error) {
	req, err := makeCreateProcessRequest(*opts).Message()
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}
	msg, err := doRequest(c.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	resp, err := ExtractInfoResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "could not read response")
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
	msg, err := doRequest(c.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	resp, err := ExtractInfosResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "could not read response")
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
	msg, err := doRequest(c.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	resp, err := ExtractInfosResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "could not read response")
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
	msg, err := doRequest(c.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	resp, err := ExtractInfoResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "could not read response")
	}
	info := resp.Info
	return &process{info: info, conn: c.conn}, nil
}

func (c *client) Clear(ctx context.Context) {
	req, err := makeClearRequest().Message()
	if err != nil {
		grip.Warning(message.WrapError(err, "could not make request"))
		return
	}
	msg, err := doRequest(c.conn, req)
	if err != nil {
		grip.Warning(message.WrapError(err, "request failed"))
		return
	}
	resp, err := ExtractInfoResponse(msg)
	if err != nil {
		grip.Warning(message.WrapError(err, "could not read response"))
		return
	}
	grip.Warning(errors.New(resp.Error))
}

func (c *client) Close(ctx context.Context) error {
	req, err := makeCloseRequest().Message()
	if err != nil {
		return errors.Wrap(err, "could not make request")
	}
	msg, err := doRequest(c.conn, req)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	resp, err := ExtractErrorResponse(msg)
	if err != nil {
		return errors.Wrap(err, "could not read response")
	}
	return errors.New(resp.Error)
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
		return jasper.ProcessInfo{}
	}
	msg, err := doRequest(p.conn, req)
	if err != nil {
		return jasper.ProcessInfo{}
	}
	resp, err := ExtractInfoResponse(msg)
	if err != nil {
		return jasper.ProcessInfo{}
	}
	grip.Warning(errors.New(resp.Error))
	p.info = resp.Info
	return p.info
}

func (p *process) Running(ctx context.Context) bool {
	if p.info.Complete {
		return false
	}

	return p.Info(ctx).IsRunning
}

func (p *process) Complete(ctx context.Context) bool {
	if p.info.Complete {
		return true
	}

	return p.Info(ctx).Complete
}

func (p *process) Signal(ctx context.Context, sig syscall.Signal) error {
	req, err := makeSignalRequest(p.ID(), int(sig)).Message()
	if err != nil {
		return errors.Wrap(err, "could not make request")
	}
	msg, err := doRequest(p.conn, req)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	resp, err := ExtractErrorResponse(msg)
	return errors.New(resp.Error)
}

func (p *process) Wait(ctx context.Context) (int, error) {
	req, err := makeWaitRequest(p.ID()).Message()
	if err != nil {
		return -1, errors.Wrap(err, "could not make request")
	}
	msg, err := doRequest(p.conn, req)
	if err != nil {
		return -1, errors.Wrap(err, "request failed")
	}
	resp, err := ExtractWaitResponse(msg)
	if err != nil {
		return -1, errors.Wrap(err, "could not read response")
	}
	return resp.ExitCode, errors.New(resp.Error)
}

func (p *process) Respawn(ctx context.Context) (jasper.Process, error) {
	req, err := makeRespawnRequest(p.ID()).Message()
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}
	msg, err := doRequest(p.conn, req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	resp, err := ExtractInfoResponse(msg)
	if err != nil {
		return nil, errors.Wrap(err, "could not read response")
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
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
		return errors.Wrap(err, "could not make request")
	}
	msg, err := doRequest(p.conn, req)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	resp, err := ExtractErrorResponse(msg)
	if err != nil {
		return errors.Wrap(err, "could not read response")
	}
	return errors.New(resp.Error)
}

func (p *process) Tag(tag string) {
	req, err := makeTagRequest(p.ID(), tag).Message()
	if err != nil {
		grip.Warningf("failed to tag process %s with tag %s", p.ID(), tag)
		return
	}
	msg, err := doRequest(p.conn, req)
	if err != nil {
		grip.Warningf("failed to tag process %s with tag %s", p.ID(), tag)
		return
	}
	resp, err := ExtractErrorResponse(msg)
	if err != nil {
		grip.Warningf("failed to tag process %s with tag %s", p.ID(), tag)
		return
	}
	grip.Warning(errors.New(resp.Error))
}

func (p *process) GetTags() []string {
	req, err := makeGetTagsRequest(p.ID()).Message()
	if err != nil {
		return nil
	}
	msg, err := doRequest(p.conn, req)
	if err != nil {
		return nil
	}
	resp, err := ExtractGetTagsResponse(msg)
	if err != nil {
		return nil
	}
	grip.Warning(errors.New(resp.Error))
	return resp.Tags
}

func (p *process) ResetTags() {
	req, err := makeResetTagsRequest(p.ID()).Message()
	if err != nil {
		grip.Warningf("failed to reset tags for process %s", p.ID())
		return
	}
	msg, err := doRequest(p.conn, req)
	if err != nil {
		grip.Warningf("failed to reset tags for process %s", p.ID())
		return
	}
	resp, err := ExtractErrorResponse(msg)
	if err != nil {
		grip.Warningf("failed to reset tags for process %s", p.ID())
		return
	}
	grip.Warning(errors.New(resp.Error))
}

func doRequest(rw io.ReadWriter, req mongowire.Message) (mongowire.Message, error) {
	if err := mongowire.SendMessage(req, rw); err != nil {
		return nil, errors.Wrap(err, "could not send request")
	}
	msg, err := mongowire.ReadMessage(rw)
	if err != nil {
		return nil, errors.Wrap(err, "could not read response")
	}
	return msg, nil
}
