package mongowire

import (
	"context"
	"io"
	"time"

	"github.com/mongodb/ftdc/bsonx"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc"
	mongorpcBson "github.com/tychoish/mongorpc/bson"
	"github.com/tychoish/mongorpc/mongowire"
	"gopkg.in/mgo.v2/bson"
)

// Constants representing allowed OP_COMMAND commmands.
const (
	IsMasterCommand   = "isMaster"
	WhatsMyURICommand = "whatsmyuri"
	// Dumb, but the shell sends commands with different casing
	BuildInfoCommand               = "buildInfo"
	BuildinfoCommand               = "buildinfo"
	GetLogCommand                  = "getLog"
	GetFreeMonitoringStatusCommand = "getFreeMonitoringStatus"
	ReplSetGetStatusCommand        = "replSetGetStatus"
	ListCollectionsCommand         = "listCollections"

	CreateProcessCommand = "createProcess"
	GetCommand           = "get"
	ListCommand          = "list"
	GroupCommand         = "group"
	ClearCommand         = "clear"
	CloseCommand         = "close"
)

var (
	notOKResp = bsonx.EC.Int32("ok", 0)
	okResp    = bsonx.EC.Int32("ok", 1)
)

type Service struct {
	manager jasper.Manager
}

// NewManagerService wraps an existing Jasper manager in a mongo wire protocol
// service.
func NewManagerService(m jasper.Manager) *Service {
	return &Service{
		manager: m,
	}
}

func (s *Service) isMaster(ctx context.Context, w io.Writer, msg mongowire.Message) {
	writeSuccessReply(w, bsonx.NewDocument(), IsMasterCommand)
}

func (s *Service) whatsMyURI(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := bsonx.NewDocument(bsonx.EC.String("you", "localhost:12345"))
	writeReply(w, resp, WhatsMyURICommand)
}

func (s *Service) buildInfo(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := bsonx.NewDocument(bsonx.EC.String("version", "0.0.0"))
	writeReply(w, resp, BuildInfoCommand)
}

func (s *Service) getLog(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := bsonx.NewDocument(bsonx.EC.ArrayFromElements("log", bsonx.VC.ArrayFromValues(bsonx.VC.String("hello"))))
	writeSuccessReply(w, resp, GetLogCommand)
}

func (s *Service) getFreeMonitoringStatus(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := bsonx.NewDocument(notOKResp)
	writeReply(w, resp, GetFreeMonitoringStatusCommand)
}

func (s *Service) replSetGetStatus(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := bsonx.NewDocument(notOKResp)
	writeReply(w, resp, ReplSetGetStatusCommand)
}

func (s *Service) listCollections(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := bsonx.NewDocument(notOKResp)
	writeReply(w, resp, ListCollectionsCommand)
}

func (s *Service) createProcess(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.New("could not read request"), CreateProcessCommand)
		return
	}
	optsValue, ok := req.Lookup(CreateProcessCommand).MutableDocumentOK()
	if !ok {
		writeErrorReply(w, errors.New("could not read process create options from request"), CreateProcessCommand)
		return
	}
	optsBytes, err := optsValue.MarshalBSON()
	if err != nil {
		err = errors.Wrap(err, "could not convert process create options to BSON")
		writeErrorReply(w, err, CreateProcessCommand)
		return
	}
	opts := options.Create{}
	err = bson.Unmarshal(optsBytes, &opts)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not convert BSON to process create options"), CreateProcessCommand)
		return
	}

	// Spawn a new context so that the process' context is not potentially
	// canceled by the request's. See how rest_service.go's createProcess() does
	// this same thing.
	pctx, cancel := context.WithCancel(context.Background())

	proc, err := s.manager.CreateProcess(pctx, &opts)
	if err != nil {
		cancel()
		writeErrorReply(w, errors.Wrap(err, "could not create process"), CreateProcessCommand)
		return
	}

	if err := proc.RegisterTrigger(ctx, func(_ jasper.ProcessInfo) {
		cancel()
	}); err != nil {
		if info := getProcInfoNoHang(ctx, proc); !info.Complete {
			cancel()
		}
	}

	info, err := procInfoToDocument(getProcInfoNoHang(ctx, proc))
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not convert process info to document"), CreateProcessCommand)
		return
	}
	resp := bsonx.NewDocument(bsonx.EC.SubDocument("info", info))
	writeSuccessReply(w, resp, CreateProcessCommand)
}

func (s *Service) listProcesses(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.New("could not read request"), ListCommand)
		return
	}
	filter, ok := req.Lookup(ListCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.New("could not read process filter from request"), ListCommand)
		return
	}

	procs, err := s.manager.List(ctx, options.Filter(filter))
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not list processes"), ListCommand)
		return
	}

	infos, err := procInfosToBSONArray(ctx, procs)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not convert process information to BSON document"), ListCommand)
	}

	resp := bsonx.NewDocument(bsonx.EC.Array("infos", infos))

	writeSuccessReply(w, resp, ListCommand)
}

func (s *Service) listGroupMembers(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), GroupCommand)
		return
	}
	tag, ok := req.Lookup(GroupCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read group tag from request"), GroupCommand)
		return
	}

	procs, err := s.manager.Group(ctx, tag)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process group"), GroupCommand)
		return
	}

	procInfos, err := procInfosToBSONArray(ctx, procs)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "couuld not convert process information to BSON document"), GroupCommand)
		return
	}

	resp := bsonx.NewDocument(
		bsonx.EC.Array("infos", procInfos),
	)

	writeSuccessReply(w, resp, GroupCommand)
}

func (s *Service) getProcess(ctx context.Context, w io.Writer, msg mongowire.Message) {
	cmdMsg, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), GetCommand)
		return
	}
	id, ok := cmdMsg.Lookup(GetCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.New("could not read process id from request"), GetCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), GetCommand)
		return
	}

	info, err := procInfoToDocument(proc.Info(ctx))
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not convert process info to document"), GetCommand)
		return
	}

	resp := bsonx.NewDocument(bsonx.EC.SubDocument("info", info))

	writeSuccessReply(w, resp, GetCommand)
}

func (s *Service) clearManager(ctx context.Context, w io.Writer, msg mongowire.Message) {
	s.manager.Clear(ctx)
	writeSuccessReply(w, bsonx.NewDocument(), ClearCommand)
}

func (s *Service) closeManager(ctx context.Context, w io.Writer, msg mongowire.Message) {
	if err := s.manager.Close(ctx); err != nil {
		writeErrorReply(w, err, CloseCommand)
		return
	}
	writeSuccessReply(w, bsonx.NewDocument(), CloseCommand)
}

// func (s *Service) getProcessTags(ctx context.Context, w io.Writer, msg mongowire.Message) {
//     cmdMsg, ok := msg.(*mongowire.CommandMessage)
//     if !ok {
//         grip.Error(errors.New("received unexpected mongo message"))
//         return
//     }
//     reqDoc, err := bsonx.ReadDocument(cmdMsg.CommandArgs.BSON)
//     if err != nil {
//         grip.Error(errors.New("received unexpected mongo message"))
//         return
//     }
//     id, ok := reqDoc.Lookup("id").Int64ValueOK()
//     if !ok {
//         grip.Error(errors.Wrap(writeErrorReply(w, err), "error writing response after get id failure"))
//         return
//     }
//     proc, err := s.manager.Get(ctx, id)
//     if err != nil {
//         grip.Error(errors.WithStack(writeErrorReply(w, err), "error writing response after get process failure"))
//         return
//     }
// }

func getProcInfoNoHang(ctx context.Context, p jasper.Process) jasper.ProcessInfo {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return p.Info(ctx)
}

func messageToDocument(msg mongowire.Message) (*bsonx.Document, error) {
	cmdMsg, ok := msg.(*mongowire.CommandMessage)
	if !ok {
		// return nil, errors.New("message is not of type %s", mongowire.OP_COMMAND.String())
		return nil, errors.New("kim: TODO: MAKE-984")
	}
	return bsonx.ReadDocument(cmdMsg.CommandArgs.BSON)
}

func procInfosToBSONArray(ctx context.Context, procs []jasper.Process) (*bsonx.Array, error) {
	infos := bsonx.MakeArray(len(procs))
	for _, proc := range procs {
		info, err := procInfoToDocument(proc.Info(ctx))
		if err != nil {
			return infos, errors.Wrapf(err, "could not  convert process info to document for process %s", proc.ID())
		}
		infos.Append(bsonx.VC.Document(info))
	}
	return infos, nil
}

func procInfoToDocument(info jasper.ProcessInfo) (*bsonx.Document, error) {
	infoBytes, err := bson.Marshal(info)
	if err != nil {
		return nil, err
	}
	return bsonx.ReadDocument(infoBytes)
}

// kim: TODO: change op to mongowire.OpType
func writeErrorReply(w io.Writer, err error, op string) {
	errorDoc := bsonx.EC.String("error", err.Error())
	doc := bsonx.NewDocument(notOKResp, errorDoc)
	writeReply(w, doc, op)
}

func writeSuccessReply(w io.Writer, doc *bsonx.Document, op string) {
	doc.Prepend(okResp)
	writeReply(w, doc, op)
}

func writeReply(w io.Writer, doc *bsonx.Document, op string) {
	resp, err := doc.MarshalBSON()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message": "could not marshal BSON response",
			"op":      op,
		}))
		return
	}

	respDoc := mongorpcBson.Simple{BSON: resp, Size: int32(len(resp))}

	reply := mongowire.NewReply(int64(0), int32(0), int32(0), int32(1), []mongorpcBson.Simple{respDoc})
	_, err = w.Write(reply.Serialize())
	grip.Error(message.WrapError(err, message.Fields{
		"message": "could not write response",
		"op":      op,
	}))
}

func (s *Service) RegisterHandlers(host string, port int) (*mongorpc.Service, error) {
	srv := mongorpc.NewService(host, port)

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: IsMasterCommand,
	}, s.isMaster); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", IsMasterCommand)
	}

	// kim: TODO: would be better if this wasn't done registered both admin and
	// the current DB , or if we could just register an operation regardless of
	// DB context.

	// Register required handlers
	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: IsMasterCommand,
	}, s.isMaster); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", IsMasterCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: WhatsMyURICommand,
	}, s.whatsMyURI); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", WhatsMyURICommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: BuildinfoCommand,
	}, s.buildInfo); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", BuildinfoCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: BuildInfoCommand,
	}, s.buildInfo); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", BuildInfoCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: GetLogCommand,
	}, s.getLog); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", GetLogCommand)
	}

	// kim: TODO: would be better if these didn't have a dummy context DB
	// because they don't operate on databases anyways.
	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: GetLogCommand,
	}, s.getLog); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", GetLogCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: GetFreeMonitoringStatusCommand,
	}, s.getFreeMonitoringStatus); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", GetFreeMonitoringStatusCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: ReplSetGetStatusCommand,
	}, s.replSetGetStatus); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", ReplSetGetStatusCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: ListCollectionsCommand,
	}, s.listCollections); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", ListCollectionsCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: CreateProcessCommand,
	}, s.createProcess); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", CreateProcessCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: ListCommand,
	}, s.listProcesses); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", ListCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: GroupCommand,
	}, s.listGroupMembers); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", GroupCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: GetCommand,
	}, s.getProcess); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", GetCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: ClearCommand,
	}, s.clearManager); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", ClearCommand)
	}

	if err := srv.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: CloseCommand,
	}, s.closeManager); err != nil {
		return nil, errors.Wrapf(err, "could not register handler for %s", CloseCommand)
	}

	return srv, nil
}
