package mongowire

import (
	"context"
	"io"

	"github.com/mongodb/ftdc/bsonx"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc"
	"github.com/tychoish/mongorpc/mongowire"
)

// Constants representing required commands.
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
)

var (
	notOKResp = bsonx.EC.Int32("ok", 0)
	okResp    = bsonx.EC.Int32("ok", 1)
)

type Service struct {
	*mongorpc.Service
	manager jasper.Manager
}

// NewManagerService wraps an existing Jasper manager in a mongo wire protocol
// service.
func NewManagerService(m jasper.Manager, host string, port int) (*Service, error) {
	service := &Service{
		Service: mongorpc.NewService(host, port),
		manager: m,
	}
	if err := service.registerHandlers(); err != nil {
		return nil, errors.Wrap(err, "error registering handlers")
	}
	return service, nil
}

func (s *Service) isMaster(ctx context.Context, w io.Writer, msg mongowire.Message) {
	writeSuccessReply(w, bsonx.NewDocument(), IsMasterCommand)
}

func (s *Service) whatsMyURI(ctx context.Context, w io.Writer, msg mongowire.Message) {
	// kim: TODO: replace with address
	resp := bsonx.NewDocument(bsonx.EC.String("you", "localhost:12345"))
	writeReply(w, resp, WhatsMyURICommand)
}

func (s *Service) buildInfo(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := bsonx.NewDocument(bsonx.EC.String("version", "0.0.0"))
	writeReply(w, resp, BuildInfoCommand)
}

func (s *Service) getLog(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := bsonx.NewDocument(bsonx.EC.Array("log", bsonx.NewArray()))
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

func (s *Service) registerHandlers() error {
	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: IsMasterCommand,
	}, s.isMaster); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", IsMasterCommand)
	}

	// kim: TODO: would be better if this wasn't done registered both admin and
	// the current DB , or if we could just register an operation regardless of
	// DB context.

	// Register required handlers
	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: IsMasterCommand,
	}, s.isMaster); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", IsMasterCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: WhatsMyURICommand,
	}, s.whatsMyURI); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", WhatsMyURICommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: BuildinfoCommand,
	}, s.buildInfo); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", BuildinfoCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: BuildInfoCommand,
	}, s.buildInfo); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", BuildInfoCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: GetLogCommand,
	}, s.getLog); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", GetLogCommand)
	}

	// kim: TODO: would be better if these didn't have a dummy context DB
	// because they don't operate on databases anyways.
	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: GetLogCommand,
	}, s.getLog); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", GetLogCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: GetFreeMonitoringStatusCommand,
	}, s.getFreeMonitoringStatus); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", GetFreeMonitoringStatusCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "admin",
		Command: ReplSetGetStatusCommand,
	}, s.replSetGetStatus); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", ReplSetGetStatusCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: ListCollectionsCommand,
	}, s.listCollections); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", ListCollectionsCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: CreateProcessCommand,
	}, s.managerCreateProcess); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", CreateProcessCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: ListCommand,
	}, s.managerList); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", ListCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: GroupCommand,
	}, s.managerGroup); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", GroupCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: GetCommand,
	}, s.managerGet); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", GetCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: ClearCommand,
	}, s.managerClear); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", ClearCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: CloseCommand,
	}, s.managerClose); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", CloseCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: ProcessIDCommand,
	}, s.processID); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", ProcessIDCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: InfoCommand,
	}, s.processInfo); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", InfoCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: RunningCommand,
	}, s.processRunning); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", RunningCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: CompleteCommand,
	}, s.processComplete); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", CompleteCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: WaitCommand,
	}, s.processWait); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", WaitCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: RespawnCommand,
	}, s.processRespawn); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", RespawnCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: SignalCommand,
	}, s.processSignal); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", SignalCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: RegisterSignalTriggerIDCommand,
	}, s.processRegisterSignalTriggerID); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", RegisterSignalTriggerIDCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: TagCommand,
	}, s.processTag); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", TagCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: GetTagsCommand,
	}, s.processGetTags); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", GetTagsCommand)
	}

	if err := s.RegisterOperation(&mongowire.OpScope{
		Type:    mongowire.OP_COMMAND,
		Context: "test",
		Command: ResetTagsCommand,
	}, s.processResetTags); err != nil {
		return errors.Wrapf(err, "could not register handler for %s", ResetTagsCommand)
	}

	return nil
}
