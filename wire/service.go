package wire

import (
	"context"
	"net"
	"strconv"

	"github.com/evergreen-ci/mrpc"
	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/evergreen-ci/mrpc/shell"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/recovery"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
)

// TODO: support jasper.RemoteClient functionality
type service struct {
	mrpc.Service
	manager jasper.Manager
}

// StartService wraps an existing Jasper manager in a mongo wire protocol
// service and starts it. The caller is responsible for closing the connection
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

	baseSvc, err := shell.NewShellService(host, port)
	if err != nil {
		return nil, errors.Wrap(err, "could not create base service")
	}
	svc := &service{
		Service: baseSvc,
		manager: m,
	}
	if err := svc.registerHandlers(); err != nil {
		return nil, errors.Wrap(err, "error registering handlers")
	}

	cctx, ccancel := context.WithCancel(context.Background())
	go func() {
		defer func() {
			grip.Error(recovery.HandlePanicWithError(recover(), nil, "running wire service"))
		}()
		grip.Notice(svc.Run(cctx))
	}()

	return func() error { ccancel(); return nil }, nil
}

func (s *service) registerHandlers() error {
	for name, handler := range map[string]mrpc.HandlerFunc{
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
