package mongowire

import (
	"context"
	"io"
	"syscall"

	"github.com/evergreen-ci/birch"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc/mongowire"
)

// Constants representing process commands.
const (
	ProcessIDCommand               = "processID"
	InfoCommand                    = "info"
	RunningCommand                 = "running"
	CompleteCommand                = "complete"
	WaitCommand                    = "wait"
	RespawnCommand                 = "respawn"
	SignalCommand                  = "signal"
	RegisterSignalTriggerIDCommand = "registerSignalTriggerID"
	GetTagsCommand                 = "getTags"
	TagCommand                     = "tag"
	ResetTagsCommand               = "resetTags"
)

func (s *Service) processID(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), ProcessIDCommand)
		return
	}
	id, ok := req.Lookup(ProcessIDCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), ProcessIDCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), ProcessIDCommand)
		return
	}

	resp := birch.NewDocument(birch.EC.String("id", proc.ID()))

	writeSuccessReply(w, resp, ProcessIDCommand)
}

func (s *Service) processInfo(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), InfoCommand)
		return
	}
	id, ok := req.Lookup(InfoCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), InfoCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), InfoCommand)
		return
	}

	info, err := procInfoToDocument(proc.Info(ctx))
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not convert process info to BSON document"), InfoCommand)
		return
	}

	resp := birch.NewDocument(birch.EC.SubDocument("info", info))

	writeSuccessReply(w, resp, InfoCommand)
}

func (s *Service) processRunning(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), RunningCommand)
		return
	}
	id, ok := req.Lookup(RunningCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), RunningCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), RunningCommand)
		return
	}

	resp := birch.NewDocument(birch.EC.Boolean("running", proc.Running(ctx)))

	writeSuccessReply(w, resp, RunningCommand)
}

func (s *Service) processComplete(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), CompleteCommand)
		return
	}
	id, ok := req.Lookup(CompleteCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), CompleteCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), CompleteCommand)
		return
	}

	resp := birch.NewDocument(birch.EC.Boolean("complete", proc.Complete(ctx)))

	writeSuccessReply(w, resp, CompleteCommand)
}

func (s *Service) processWait(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), WaitCommand)
		return
	}
	id, ok := req.Lookup(WaitCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), WaitCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), WaitCommand)
		return
	}

	exitCode, err := proc.Wait(ctx)
	resp := birch.NewDocument(birch.EC.Int("exitCode", exitCode))
	if err != nil {
		resp.Append(birch.EC.String("error", err.Error()))
	}

	writeSuccessReply(w, resp, WaitCommand)
}

func (s *Service) processRespawn(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), RespawnCommand)
		return
	}
	id, ok := req.Lookup(RespawnCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), RespawnCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), RespawnCommand)
		return
	}

	newProc, err := proc.Respawn(ctx)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not respawn process"), RespawnCommand)
		return
	}

	info, err := procInfoToDocument(newProc.Info(ctx))
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not convert process info to BSON document"), RespawnCommand)
		return
	}

	resp := birch.NewDocument(birch.EC.SubDocument("info", info))

	writeSuccessReply(w, resp, RespawnCommand)
}

func (s *Service) processSignal(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), SignalCommand)
		return
	}
	signalArgs, ok := req.Lookup(SignalCommand).MutableDocumentOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process signal arguments from request"), SignalCommand)
		return
	}
	id, ok := signalArgs.Lookup("id").StringValueOK()
	if !ok {
		writeErrorReply(w, errors.New("could not read process id from request"), SignalCommand)
		return
	}

	sigVal := signalArgs.Lookup("signal")
	sig, ok := sigVal.IntOK()
	if !ok {
		// The mongo shell treats number literals as doubles by default.
		sigDouble, ok := sigVal.DoubleOK()
		sig = int(sigDouble)
		if !ok {
			writeErrorReply(w, errors.New("could not read signal number from request"), SignalCommand)
			return
		}
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), SignalCommand)
		return
	}

	if err := proc.Signal(ctx, syscall.Signal(sig)); err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not signal process"), SignalCommand)
		return
	}

	writeSuccessReply(w, birch.NewDocument(), SignalCommand)
}

func (s *Service) processRegisterSignalTriggerID(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), RegisterSignalTriggerIDCommand)
		return
	}
	signalTriggerArgs, ok := req.Lookup(RegisterSignalTriggerIDCommand).MutableDocumentOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process signal trigger arguments from request"), RegisterSignalTriggerIDCommand)
		return
	}
	procID, ok := signalTriggerArgs.Lookup("id").StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), RegisterSignalTriggerIDCommand)
		return
	}
	sigID, ok := signalTriggerArgs.Lookup("signal").StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read signal trigger ID from request"), RegisterSignalTriggerIDCommand)
		return
	}

	makeTrigger, ok := jasper.GetSignalTriggerFactory(jasper.SignalTriggerID(sigID))
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not get signal trigger for ID"), RegisterSignalTriggerIDCommand)
		return
	}

	proc, err := s.manager.Get(ctx, procID)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), RegisterSignalTriggerIDCommand)
		return
	}

	if err := proc.RegisterSignalTrigger(ctx, makeTrigger()); err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not register signal trigger ID"), RegisterSignalTriggerIDCommand)
		return
	}

	writeSuccessReply(w, birch.NewDocument(), RegisterSignalTriggerIDCommand)
}

func (s *Service) processTag(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), TagCommand)
		return
	}
	tagArgs, ok := req.Lookup(TagCommand).MutableDocumentOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process signal trigger arguments from request"), RegisterSignalTriggerIDCommand)
		return
	}
	id, ok := tagArgs.Lookup("id").StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), RegisterSignalTriggerIDCommand)
		return
	}
	tag, ok := tagArgs.Lookup("tag").StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read signal trigger ID from request"), RegisterSignalTriggerIDCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), TagCommand)
		return
	}

	proc.Tag(tag)

	writeSuccessReply(w, birch.NewDocument(), TagCommand)
}

func (s *Service) processGetTags(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), GetTagsCommand)
		return
	}
	id, ok := req.Lookup(GetTagsCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), GetTagsCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), GetTagsCommand)
		return
	}

	tags, err := procTagsToArray(proc)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not convert process tags to BSON array"), GetTagsCommand)
		return
	}

	resp := birch.NewDocument(birch.EC.Array("tags", tags))

	writeSuccessReply(w, resp, GetTagsCommand)
}

func (s *Service) processResetTags(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := messageToDocument(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), GetTagsCommand)
		return
	}
	id, ok := req.Lookup(ResetTagsCommand).StringValueOK()
	if !ok {
		writeErrorReply(w, errors.Wrap(err, "could not read process id from request"), GetTagsCommand)
		return
	}

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), GetTagsCommand)
		return
	}

	proc.ResetTags()

	writeSuccessReply(w, birch.NewDocument(), GetTagsCommand)
}
