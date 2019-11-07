package wire

import (
	"context"
	"io"

	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
)

// Constants representing manager commands.
const (
	ManagerIDCommand     = "id"
	CreateProcessCommand = "create_process"
	GetCommand           = "get"
	ListCommand          = "list"
	GroupCommand         = "group"
	ClearCommand         = "clear"
	CloseCommand         = "close"
)

func (s *service) managerID(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp, err := makeIDResponse(s.manager.ID()).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.New("could not make response"), ManagerIDCommand)
		return
	}
	writeResponse(ctx, w, resp, ManagerIDCommand)
}

func (s *service) managerCreateProcess(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractCreateProcessRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), CreateProcessCommand)
		return
	}

	opts := req.Options

	// Spawn a new context so that the process' context is not potentially
	// canceled by the request's. See how rest_service.go's createProcess() does
	// this same thing.
	pctx, cancel := context.WithCancel(context.Background())

	proc, err := s.manager.CreateProcess(pctx, &opts)
	if err != nil {
		cancel()
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not create process"), CreateProcessCommand)
		return
	}

	if err = proc.RegisterTrigger(ctx, func(_ jasper.ProcessInfo) {
		cancel()
	}); err != nil {
		info := getProcInfoNoHang(ctx, proc)
		cancel()
		// If we get an error registering a trigger, then we should make sure that
		// the reason for it isn't just because the process has exited already,
		// since that should not be considered an error.
		if !info.Complete {
			writeErrorResponse(ctx, w, errors.Wrap(err, "could not register trigger"), CreateProcessCommand)
			return
		}
	}

	resp, err := makeInfoResponse(getProcInfoNoHang(ctx, proc)).Message()
	if err != nil {
		cancel()
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), CreateProcessCommand)
		return
	}
	writeResponse(ctx, w, resp, CreateProcessCommand)
}

func (s *service) managerList(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractListRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), ListCommand)
	}
	filter := req.Filter

	procs, err := s.manager.List(ctx, filter)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not list processes"), ListCommand)
		return
	}

	infos := make([]jasper.ProcessInfo, 0, len(procs))
	for _, proc := range procs {
		infos = append(infos, proc.Info(ctx))
	}
	resp, err := makeInfosResponse(infos).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), ListCommand)
		return
	}
	writeResponse(ctx, w, resp, ListCommand)
}

func (s *service) managerGroup(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractGroupRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), GroupCommand)
		return
	}
	tag := req.Tag

	procs, err := s.manager.Group(ctx, tag)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process group"), GroupCommand)
		return
	}

	infos := make([]jasper.ProcessInfo, 0, len(procs))
	for _, proc := range procs {
		infos = append(infos, proc.Info(ctx))
	}

	resp, err := makeInfosResponse(infos).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), GroupCommand)
		return
	}
	writeResponse(ctx, w, resp, GroupCommand)
}

func (s *service) managerGet(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractGetRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), GetCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), GetCommand)
		return
	}

	resp, err := makeInfoResponse(proc.Info(ctx)).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), GetCommand)
		return
	}
	writeResponse(ctx, w, resp, GetCommand)
}

func (s *service) managerClear(ctx context.Context, w io.Writer, msg mongowire.Message) {
	s.manager.Clear(ctx)
	writeOKResponse(ctx, w, ClearCommand)
}

func (s *service) managerClose(ctx context.Context, w io.Writer, msg mongowire.Message) {
	if err := s.manager.Close(ctx); err != nil {
		writeErrorResponse(ctx, w, err, CloseCommand)
		return
	}
	writeOKResponse(ctx, w, CloseCommand)
}
