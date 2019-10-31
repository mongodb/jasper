package mongowire

import (
	"context"
	"net"

	"github.com/evergreen-ci/birch"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc/mongowire"
)

type client struct {
	conn      net.Conn
	namespace string
}

// TODO: is implementing all RemoteClient functionality necessary?
func NewClient(addr net.Addr) (jasper.RemoteClient, error) {
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not dial address")
	}
	// <namespace>.$cmd format is required to indicate the OP_QUERY should be
	// converted to OP_COMMAND
	return &client{conn: conn, namespace: "jasper.$cmd"}, nil
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
	input := birch.NewDocument(birch.EC.Double(ManagerIDCommand, 1))
	req := mongowire.NewQuery(c.namespace, 0, 0, 1, input, birch.NewDocument())
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
	return nil, nil
}

func (c *client) CreateCommand(ctx context.Context) *jasper.Command {
	return jasper.NewCommand().ProcConstructor(c.CreateProcess)
}

func (c *client) Register(ctx context.Context, proc jasper.Process) error {
	return errors.New("cannot register extant processes on remote process managers")
}

func (c *client) List(ctx context.Context, f options.Filter) ([]jasper.Process, error) {
	return nil, nil
}

func (c *client) Group(ctx context.Context, name string) ([]jasper.Process, error) {
	return nil, nil
}

func (c *client) Get(ctx context.Context, id string) (jasper.Process, error) {
	return nil, nil
}

func (c *client) Clear(ctx context.Context) {

}

func (c *client) Close(ctx context.Context) error {
	return nil
}
