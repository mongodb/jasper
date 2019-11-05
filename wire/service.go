package wire

import (
	"context"
	"io"
	"net"
	"strconv"

	"github.com/evergreen-ci/mrpc"
	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/mongodb/grip"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
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

// TODO: support jasper.RemoteClient functionality
type service struct {
	*mrpc.Service
	manager jasper.Manager
}

// StartService wraps an existing Jasper manager in a mongo wire protocol
// service and starts it. The  caller is responsible for closing the connection
// using the returned jasper.CloseFunc.
func StartService(ctx context.Context, m jasper.Manager, addr net.Addr) (jasper.CloseFunc, error) { //nolint: interfacer
	host, p, err := net.SplitHostPort(addr.String())
	if err != nil {
		return nil, errors.Wrap(err, "invalid address")
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return nil, errors.Wrap(err, "port is not a number")
	}

	svc := &service{
		Service: mrpc.NewService(host, port),
		manager: m,
	}
	if err := svc.registerHandlers(); err != nil {
		return nil, errors.Wrap(err, "error registering handlers")
	}

	cctx, ccancel := context.WithCancel(context.Background())
	go func() {
		grip.Notice(svc.Run(cctx))
	}()

	return func() error { ccancel(); return nil }, nil
}

func (s *service) isMaster(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp, err := makeErrorResponse(true, nil).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), IsMasterCommand)
		return
	}
	writeResponse(ctx, w, resp, IsMasterCommand)
}

func (s *service) whatsMyURI(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp, err := makeWhatsMyURIResponse(s.Address()).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), WhatsMyURICommand)
		return
	}
	writeResponse(ctx, w, resp, WhatsMyURICommand)
}

func (s *service) buildInfo(ctx context.Context, w io.Writer, msg mongowire.Message) {
	// resp := birch.NewDocument(birch.EC.String("version", "0.0.0"))
	// writeSuccessResponse(w, resp, BuildInfoCommand)
	resp, err := makeBuildInfoResponse("0.0.0").Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), BuildInfoCommand)
		return
	}
	writeResponse(ctx, w, resp, BuildInfoCommand)
}

func (s *service) getLog(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp, err := makeGetLogResponse([]string{}).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), GetLogCommand)
		return
	}
	writeResponse(ctx, w, resp, GetLogCommand)
}

func (s *service) getFreeMonitoringStatus(ctx context.Context, w io.Writer, msg mongowire.Message) {
	writeNotOKResponse(ctx, w, GetFreeMonitoringStatusCommand)
}

func (s *service) replSetGetStatus(ctx context.Context, w io.Writer, msg mongowire.Message) {
	writeNotOKResponse(ctx, w, ReplSetGetStatusCommand)
}

func (s *service) listCollections(ctx context.Context, w io.Writer, msg mongowire.Message) {
	writeNotOKResponse(ctx, w, ListCollectionsCommand)
}

func (s *service) registerHandlers() error {
	for name, handler := range map[string]mrpc.HandlerFunc{
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
