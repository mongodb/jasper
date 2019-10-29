package mongowire

import (
	"context"
	"io"

	"github.com/mongodb/ftdc/bsonx"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc/mongowire"
	"gopkg.in/mgo.v2/bson"
)

// Constants representing manager commands.
const (
	ManagerIDCommand     = "managerID"
	CreateProcessCommand = "createProcess"
	GetCommand           = "get"
	ListCommand          = "list"
	GroupCommand         = "group"
	ClearCommand         = "clear"
	CloseCommand         = "close"
)

func (s *Service) managerID(ctx context.Context, w io.Writer, msg mongowire.Message) {
	resp := bsonx.NewDocument(bsonx.EC.String("id", s.manager.ID()))
	writeSuccessReply(w, resp, ManagerIDCommand)
}

func (s *Service) managerCreateProcess(ctx context.Context, w io.Writer, msg mongowire.Message) {
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

func (s *Service) managerList(ctx context.Context, w io.Writer, msg mongowire.Message) {
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

	infos, err := procInfosToArray(ctx, procs)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not convert process information to BSON array"), ListCommand)
	}

	resp := bsonx.NewDocument(bsonx.EC.Array("infos", infos))

	writeSuccessReply(w, resp, ListCommand)
}

func (s *Service) managerGroup(ctx context.Context, w io.Writer, msg mongowire.Message) {
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

	procInfos, err := procInfosToArray(ctx, procs)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "couuld not convert process information to BSON array"), GroupCommand)
		return
	}

	resp := bsonx.NewDocument(
		bsonx.EC.Array("infos", procInfos),
	)

	writeSuccessReply(w, resp, GroupCommand)
}

func (s *Service) managerGet(ctx context.Context, w io.Writer, msg mongowire.Message) {
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

func (s *Service) managerClear(ctx context.Context, w io.Writer, msg mongowire.Message) {
	s.manager.Clear(ctx)
	writeSuccessReply(w, bsonx.NewDocument(), ClearCommand)
}

func (s *Service) managerClose(ctx context.Context, w io.Writer, msg mongowire.Message) {
	if err := s.manager.Close(ctx); err != nil {
		writeErrorReply(w, err, CloseCommand)
		return
	}
	writeSuccessReply(w, bsonx.NewDocument(), CloseCommand)
}
