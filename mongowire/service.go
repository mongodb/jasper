package mongowire

import (
	"context"
	"io"

	"github.com/evergreen-ci/birch"
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
	notOKResp = birch.EC.Int32("ok", 0)
	okResp    = birch.EC.Int32("ok", 1)
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
	writeSuccessReply(w, birch.NewDocument(), IsMasterCommand)
}

func (s *Service) whatsMyURI(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := birch.NewDocument(birch.EC.String("you", s.Address()))
	writeReply(w, resp, WhatsMyURICommand)
}

func (s *Service) buildInfo(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := birch.NewDocument(birch.EC.String("version", "0.0.0"))
	writeReply(w, resp, BuildInfoCommand)
}

func (s *Service) getLog(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := birch.NewDocument(birch.EC.Array("log", birch.NewArray()))
	writeSuccessReply(w, resp, GetLogCommand)
}

func (s *Service) getFreeMonitoringStatus(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := birch.NewDocument(notOKResp)
	writeReply(w, resp, GetFreeMonitoringStatusCommand)
}

func (s *Service) replSetGetStatus(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := birch.NewDocument(notOKResp)
	writeReply(w, resp, ReplSetGetStatusCommand)
}

func (s *Service) listCollections(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := birch.NewDocument(notOKResp)
	writeReply(w, resp, ListCollectionsCommand)
}

func (s *Service) registerHandlers() error {
	for name, handler := range map[string]mongorpc.HandlerFunc{
		// Required initialization commands
		IsMasterCommand:                s.isMaster,
		WhatsMyURICommand:              s.whatsMyURI,
		BuildinfoCommand:               s.buildInfo,
		BuildInfoCommand:               s.buildInfo,
		GetLogCommand:                  s.getLog,
		ReplSetGetStatusCommand:        s.replSetGetStatus,
		GetFreeMonitoringStatusCommand: s.getFreeMonitoringStatus,
		ListCollectionsCommand:         s.listCollections,

		// Manager commands
		ManagerIDCommand:     s.managerID,
		CreateProcessCommand: s.managerCreateProcess,
		ListCommand:          s.managerList,
		GroupCommand:         s.managerGroup,
		GetCommand:           s.managerGet,
		ClearCommand:         s.managerClear,
		CloseCommand:         s.managerClose,

		// Process commands
		ProcessIDCommand:               s.processID,
		InfoCommand:                    s.processInfo,
		RunningCommand:                 s.processRunning,
		CompleteCommand:                s.processComplete,
		WaitCommand:                    s.processWait,
		SignalCommand:                  s.processSignal,
		RegisterSignalTriggerIDCommand: s.processRegisterSignalTriggerID,
		RespawnCommand:                 s.processRespawn,
		TagCommand:                     s.processTag,
		GetTagsCommand:                 s.processGetTags,
		ResetTagsCommand:               s.processResetTags,
	} {
		if err := s.RegisterOperation(&mongowire.OpScope{
			Type:    mongowire.OP_COMMAND,
			Command: name,
		}, handler); err != nil {
			return errors.Wrapf(err, "could not register handler for %s", name)
		}
	}

	return nil
}
