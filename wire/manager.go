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

// managerID replies with a document of the form:
//     {"id": string}
func (s *service) managerID(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp, err := makeIDResponse(s.manager.ID()).Message()
	if err != nil {
		writeErrorReply(ctx, w, errors.New("could not make response"), ManagerIDCommand)
		return
	}
	writeReply(ctx, w, resp, ManagerIDCommand)
}

// managerCreateProcess replies with a document of the form:
//     {"ok": bool, "info": document}
func (s *service) managerCreateProcess(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractCreateProcessRequest(msg)
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not read request"), CreateProcessCommand)
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
		writeErrorReply(ctx, w, errors.Wrap(err, "could not create process"), CreateProcessCommand)
		return
	}

	if err := proc.RegisterTrigger(ctx, func(_ jasper.ProcessInfo) {
		cancel()
	}); err != nil {
		if info := getProcInfoNoHang(ctx, proc); !info.Complete {
			cancel()
		}
	}

	resp, err := makeInfoResponse(getProcInfoNoHang(ctx, proc)).Message()
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not make response"), CreateProcessCommand)
		return
	}
	writeReply(ctx, w, resp, CreateProcessCommand)
}

// managerList replies with a document of the form:
//     {"ok": bool, "infos": [documents...]}
func (s *service) managerList(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractListRequest(msg)
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not read request"), ListCommand)
	}
	filter := req.Filter

	procs, err := s.manager.List(ctx, filter)
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not list processes"), ListCommand)
		return
	}

	infos := make([]jasper.ProcessInfo, 0, len(procs))
	for _, proc := range procs {
		infos = append(infos, proc.Info(ctx))
	}
	resp, err := makeInfosResponse(infos).Message()
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not make response"), ListCommand)
		return
	}
	writeReply(ctx, w, resp, ListCommand)
}

// managerGroup replies with a document of the form:
//     {"ok": bool, "infos": [documents...]}
func (s *service) managerGroup(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractGroupRequest(msg)
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not read request"), GroupCommand)
		return
	}
	tag := req.Tag

	procs, err := s.manager.Group(ctx, tag)
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not get process group"), GroupCommand)
		return
	}

	infos := make([]jasper.ProcessInfo, 0, len(procs))
	for _, proc := range procs {
		infos = append(infos, proc.Info(ctx))
	}

	resp, err := makeInfosResponse(infos).Message()
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not make response"), GroupCommand)
		return
	}
	writeReply(ctx, w, resp, GroupCommand)
}

// managerGet replies with a document of the form:
//     {"ok": bool, "info": document}
func (s *service) managerGet(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractGetRequest(msg)
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not read request"), GetCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not get process"), GetCommand)
		return
	}

	resp, err := makeInfoResponse(proc.Info(ctx)).Message()
	if err != nil {
		writeErrorReply(ctx, w, errors.Wrap(err, "could not make response"), GetCommand)
		return
	}
	writeReply(ctx, w, resp, GetCommand)
}

// managerClear replies with a document of the form:
//     {"ok": bool}
func (s *service) managerClear(ctx context.Context, w io.Writer, msg mongowire.Message) {
	s.manager.Clear(ctx)
	writeOKReply(ctx, w, ClearCommand)
}

// managerClose replies with a document of the form:
//     {"ok": bool}
func (s *service) managerClose(ctx context.Context, w io.Writer, msg mongowire.Message) {
	if err := s.manager.Close(ctx); err != nil {
		writeErrorReply(ctx, w, err, CloseCommand)
		return
	}
	writeOKReply(ctx, w, CloseCommand)
}
