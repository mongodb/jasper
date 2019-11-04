package wire

import (
	"context"
	"io"
	"syscall"

	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
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
	req, err := ExtractInfoRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), InfoCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), InfoCommand)
		return
	}

	resp, err := makeInfoResponse(proc.Info(ctx)).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), InfoCommand)
		return
	}
	writeResponse(ctx, w, resp, InfoCommand)
}

func (s *service) processRunning(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractRunningRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), RunningCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), RunningCommand)
		return
	}

	resp, err := makeRunningResponse(proc.Running(ctx)).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), RunningCommand)
		return
	}
	writeResponse(ctx, w, resp, RunningCommand)
}

func (s *service) processComplete(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractCompleteRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), CompleteCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), CompleteCommand)
		return
	}

	resp, err := makeCompleteResponse(proc.Complete(ctx)).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), CompleteCommand)
		return
	}
	writeResponse(ctx, w, resp, CompleteCommand)
}

func (s *service) processWait(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractWaitRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), WaitCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), WaitCommand)
		return
	}

	exitCode, err := proc.Wait(ctx)
	resp, err := makeWaitResponse(exitCode, err).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), WaitCommand)
		return
	}
	writeResponse(ctx, w, resp, WaitCommand)
}

func (s *service) processRespawn(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractRespawnRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), RespawnCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), RespawnCommand)
		return
	}

	pctx, cancel := context.WithCancel(context.Background())
	newProc, err := proc.Respawn(pctx)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "failed to register respawned process"), RespawnCommand)
		cancel()
		return
	}
	if err = s.manager.Register(ctx, newProc); err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "problem registering respawned process"), RespawnCommand)
		cancel()
		return
	}

	if err = newProc.RegisterTrigger(ctx, func(jasper.ProcessInfo) {
		cancel()
	}); err != nil {
		newProcInfo := getProcInfoNoHang(ctx, newProc)
		if !newProcInfo.Complete {
			writeErrorResponse(ctx, w, errors.Wrap(err, "failed to register trigger on respawned process"), RespawnCommand)
			return
		}
	}

	resp, err := makeInfoResponse(newProc.Info(ctx)).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), RespawnCommand)
		return
	}
	writeResponse(ctx, w, resp, RespawnCommand)
}

func (s *service) processSignal(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractSignalRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), SignalCommand)
		return
	}
	id := req.Params.ID
	sig := int(req.Params.Signal)

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), SignalCommand)
		return
	}

	if err := proc.Signal(ctx, syscall.Signal(sig)); err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not signal process"), SignalCommand)
		return
	}

	writeOKResponse(ctx, w, SignalCommand)
}

func (s *service) processRegisterSignalTriggerID(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractRegisterSignalTriggerIDRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), RegisterSignalTriggerIDCommand)
		return
	}
	procID := req.Params.ID
	sigID := req.Params.SignalTriggerID

	makeTrigger, ok := jasper.GetSignalTriggerFactory(sigID)
	if !ok {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get signal trigger for ID"), RegisterSignalTriggerIDCommand)
		return
	}

	proc, err := s.manager.Get(ctx, procID)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), RegisterSignalTriggerIDCommand)
		return
	}

	if err := proc.RegisterSignalTrigger(ctx, makeTrigger()); err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not register signal trigger"), RegisterSignalTriggerIDCommand)
		return
	}

	writeOKResponse(ctx, w, RegisterSignalTriggerIDCommand)
}

func (s *service) processTag(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractTagRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), TagCommand)
		return
	}
	id := req.Params.ID
	tag := req.Params.Tag

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), TagCommand)
		return
	}

	proc.Tag(tag)

	writeOKResponse(ctx, w, TagCommand)
}

func (s *service) processGetTags(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractGetTagsRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), GetTagsCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), GetTagsCommand)
		return
	}

	resp, err := makeGetTagsResponse(proc.GetTags()).Message()
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not make response"), GetTagsCommand)
		return
	}
	writeResponse(ctx, w, resp, GetTagsCommand)
}

func (s *service) processResetTags(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req, err := ExtractResetTagsRequest(msg)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not read request"), ResetTagsCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeErrorResponse(ctx, w, errors.Wrap(err, "could not get process"), ResetTagsCommand)
		return
	}

	proc.ResetTags()

	writeOKResponse(ctx, w, ResetTagsCommand)
}
