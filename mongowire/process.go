package mongowire

import (
	"context"
	"io"
	"syscall"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc/mongowire"
)

// Constants representing process commands.
const (
	ProcessIDCommand               = "process_id"
	InfoCommand                    = "info"
	RunningCommand                 = "running"
	CompleteCommand                = "complete"
	WaitCommand                    = "wait"
	RespawnCommand                 = "respawn"
	SignalCommand                  = "signal"
	RegisterSignalTriggerIDCommand = "register_signal_trigger_id"
	GetTagsCommand                 = "get_tags"
	TagCommand                     = "tag"
	ResetTagsCommand               = "reset_tags"
)

func (s *service) processInfo(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractProcessInfoRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), InfoCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), InfoCommand)
		return
	}

	resp, err := makeInfoResponse(proc.Info(ctx)).Message()
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not make response"), InfoCommand)
		return
	}
	writeReply(w, resp, InfoCommand)
}

func (s *service) processRunning(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractRunningRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), RunningCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), RunningCommand)
		return
	}

	resp, err := makeRunningResponse(proc.Running(ctx)).Message()
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not make response"), RunningCommand)
		return
	}
	writeReply(w, resp, RunningCommand)
}

func (s *service) processComplete(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractCompleteRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), CompleteCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), CompleteCommand)
		return
	}

	resp, err := makeCompleteResponse(proc.Complete(ctx)).Message()
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not make response"), CompleteCommand)
		return
	}
	writeReply(w, resp, CompleteCommand)
}

func (s *service) processWait(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractWaitRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), WaitCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), WaitCommand)
		return
	}

	exitCode, err := proc.Wait(ctx)
	resp, err := makeWaitResponse(exitCode, err).Message()
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not make response"), WaitCommand)
		return
	}
	writeReply(w, resp, WaitCommand)
}

func (s *service) processRespawn(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractRespawnRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), RespawnCommand)
		return
	}
	id := req.ID

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

	resp, err := makeInfoResponse(newProc.Info(ctx)).Message()
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not make response"), RespawnCommand)
		return
	}
	writeReply(w, resp, RespawnCommand)
}

func (s *service) processSignal(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractSignalRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), SignalCommand)
		return
	}
	id := req.Params.ID
	sig := int(req.Params.Signal)

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), SignalCommand)
		return
	}

	if err := proc.Signal(ctx, syscall.Signal(sig)); err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not signal process"), SignalCommand)
		return
	}

	writeOKReply(w, SignalCommand)
}

func (s *service) processRegisterSignalTriggerID(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractRegisterSignalTriggerIDRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), RegisterSignalTriggerIDCommand)
		return
	}
	procID := req.Params.ID
	sigID := int(req.Params.SignalTriggerID)

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
		writeErrorReply(w, errors.Wrap(err, "could not register signal trigger"), RegisterSignalTriggerIDCommand)
		return
	}

	writeOKReply(w, RegisterSignalTriggerIDCommand)
}

func (s *service) processTag(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractTagRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), TagCommand)
		return
	}
	id := req.Params.ID
	tag := req.Params.Tag

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), TagCommand)
		return
	}

	proc.Tag(tag)

	writeOKReply(w, TagCommand)
}

func (s *service) processGetTags(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractGetTagsRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), GetTagsCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), GetTagsCommand)
		return
	}

	resp, err := makeGetTagsResponse(proc.GetTags()).Message()
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not make response"), GetTagsCommand)
		return
	}
	writeReply(w, resp, GetTagsCommand)
}

func (s *service) processResetTags(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractResetTagsRequest(msg)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not read request"), ResetTagsCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorReply(w, errors.Wrap(err, "could not get process"), ResetTagsCommand)
		return
	}

	proc.ResetTags()

	writeOKReply(w, ResetTagsCommand)
}
