package remote

import (
	"context"
	"io"
	"syscall"

	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/evergreen-ci/mrpc/shell"
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
	TagCommand                     = "add_tag"
	ResetTagsCommand               = "reset_tags"
)

func (s *mdbService) processInfo(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := infoRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), InfoCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), InfoCommand)
		return
	}

	resp, err := shell.ResponseToMessage(mongowire.OP_REPLY, makeInfoResponse(proc.Info(ctx)))
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting response to wire message"), InfoCommand)
		return
	}
	shell.WriteResponse(ctx, w, resp, InfoCommand)
}

func (s *mdbService) processRunning(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := runningRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), RunningCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), RunningCommand)
		return
	}

	resp, err := shell.ResponseToMessage(mongowire.OP_REPLY, makeRunningResponse(proc.Running(ctx)))
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting response to wire message"), RunningCommand)
		return
	}
	shell.WriteResponse(ctx, w, resp, RunningCommand)
}

func (s *mdbService) processComplete(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := &completeRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), CompleteCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), CompleteCommand)
		return
	}

	resp, err := shell.ResponseToMessage(mongowire.OP_REPLY, makeCompleteResponse(proc.Complete(ctx)))
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting response to wire message"), CompleteCommand)
		return
	}
	shell.WriteResponse(ctx, w, resp, CompleteCommand)
}

func (s *mdbService) processWait(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := waitRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), WaitCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), WaitCommand)
		return
	}

	exitCode, err := proc.Wait(ctx)
	resp, err := shell.ResponseToMessage(mongowire.OP_REPLY, makeWaitResponse(exitCode, err))
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting response to wire message"), WaitCommand)
		return
	}
	shell.WriteResponse(ctx, w, resp, WaitCommand)
}

func (s *mdbService) processRespawn(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := respawnRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), RespawnCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), RespawnCommand)
		return
	}

	pctx, cancel := context.WithCancel(context.Background())
	newProc, err := proc.Respawn(pctx)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "respawning process"), RespawnCommand)
		cancel()
		return
	}
	if err = s.manager.Register(ctx, newProc); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "registering respawned process"), RespawnCommand)
		cancel()
		return
	}

	if err = newProc.RegisterTrigger(ctx, func(jasper.ProcessInfo) {
		cancel()
	}); err != nil {
		newProcInfo := getProcInfoNoHang(ctx, newProc)
		cancel()
		if !newProcInfo.Complete {
			shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "registering trigger on respawned process"), RespawnCommand)
			return
		}
	}

	resp, err := shell.ResponseToMessage(mongowire.OP_REPLY, makeInfoResponse(newProc.Info(ctx)))
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting response to wire message"), RespawnCommand)
		return
	}
	shell.WriteResponse(ctx, w, resp, RespawnCommand)
}

func (s *mdbService) processSignal(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := signalRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), SignalCommand)
		return
	}
	id := req.Params.ID
	sig := req.Params.Signal

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), SignalCommand)
		return
	}

	if err := proc.Signal(ctx, syscall.Signal(sig)); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "signalling process"), SignalCommand)
		return
	}

	shell.WriteOKResponse(ctx, w, mongowire.OP_REPLY, SignalCommand)
}

func (s *mdbService) processRegisterSignalTriggerID(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := registerSignalTriggerIDRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), RegisterSignalTriggerIDCommand)
		return
	}
	procID := req.Params.ID
	sigID := req.Params.SignalTriggerID

	makeTrigger, ok := jasper.GetSignalTriggerFactory(sigID)
	if !ok {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Errorf("could not get signal trigger ID %s", sigID), RegisterSignalTriggerIDCommand)
		return
	}

	proc, err := s.manager.Get(ctx, procID)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), RegisterSignalTriggerIDCommand)
		return
	}

	if err := proc.RegisterSignalTrigger(ctx, makeTrigger()); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "registering signal trigger"), RegisterSignalTriggerIDCommand)
		return
	}

	shell.WriteOKResponse(ctx, w, mongowire.OP_REPLY, RegisterSignalTriggerIDCommand)
}

func (s *mdbService) processTag(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := tagRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), TagCommand)
		return
	}
	id := req.Params.ID
	tag := req.Params.Tag

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), TagCommand)
		return
	}

	proc.Tag(tag)

	shell.WriteOKResponse(ctx, w, mongowire.OP_REPLY, TagCommand)
}

func (s *mdbService) processGetTags(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := &getTagsRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), GetTagsCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), GetTagsCommand)
		return
	}

	resp, err := shell.ResponseToMessage(mongowire.OP_REPLY, makeGetTagsResponse(proc.GetTags()))
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting response to wire message"), GetTagsCommand)
		return
	}
	shell.WriteResponse(ctx, w, resp, GetTagsCommand)
}

func (s *mdbService) processResetTags(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := resetTagsRequest{}
	if err := shell.MessageToRequest(msg, &req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "converting wire message to request"), ResetTagsCommand)
		return
	}
	id := req.ID

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "getting process"), ResetTagsCommand)
		return
	}

	proc.ResetTags()

	shell.WriteOKResponse(ctx, w, mongowire.OP_REPLY, ResetTagsCommand)
}
