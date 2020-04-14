package remote

import (
	"context"
	"io"

	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/evergreen-ci/mrpc/shell"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

const (
	LoggingCacheSizeCommand   = "logging_cache_size"
	LoggingCacheCreateCommand = "create_logging_cache"
	LoggingCacheDeleteCommand = "delete_logging_cache"
	LoggingCacheGetCommand    = "get_logging_cache"
	LoggingSendMessageCommand = "send_message"
)

func (s *mdbService) loggingSize(ctx context.Context, w io.Writer, msg mongowire.Message) {
	lc := s.serviceLoggingCacheRequest(ctx, w, msg, nil, LoggingCacheSizeCommand)
	if lc == nil {
		return
	}

	s.serviceLoggingCacheResponse(ctx, w, &loggingCacheSizeResponse{Size: lc.Len()}, LoggingCacheSizeCommand)
}

func (s *mdbService) loggingCreate(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := &loggingCacheCreateRequest{}
	lc := s.serviceLoggingCacheRequest(ctx, w, msg, req, LoggingCacheCreateCommand)
	if lc == nil {
		return
	}

	cachedLogger, err := lc.Create(req.ID, &req.Options)
	if err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "could not create logger"), LoggingCacheCreateCommand)
		return
	}

	s.serviceLoggingCacheResponse(ctx, w, &loggingCacheCreateAndGetResponse{CachedLogger: *cachedLogger}, LoggingCacheCreateCommand)
}

func (s *mdbService) loggingGet(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := &loggingCacheGetAndDeleteRequest{}
	lc := s.serviceLoggingCacheRequest(ctx, w, msg, req, LoggingCacheGetCommand)
	if lc == nil {
		return
	}

	cachedLogger := lc.Get(req.ID)
	if cachedLogger == nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.New("named logger does not exist"), LoggingCacheGetCommand)
		return
	}

	s.serviceLoggingCacheResponse(ctx, w, &loggingCacheCreateAndGetResponse{CachedLogger: *cachedLogger}, LoggingCacheGetCommand)
}

func (s *mdbService) loggingDelete(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := &loggingCacheGetAndDeleteRequest{}
	lc := s.serviceLoggingCacheRequest(ctx, w, msg, req, LoggingCacheDeleteCommand)
	if lc == nil {
		return
	}

	lc.Remove(req.ID)

	s.serviceLoggingCacheResponse(ctx, w, nil, LoggingCacheDeleteCommand)
}

func (s *mdbService) loggingSendMessage(ctx context.Context, w io.Writer, msg mongowire.Message) {
	req := &options.LoggingPayload{}
	lc := s.serviceLoggingCacheRequest(ctx, w, msg, req, LoggingCacheDeleteCommand)
	if lc == nil {
		return
	}

	cachedLogger := lc.Get(req.LoggerID)
	if cachedLogger == nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.New("named logger does not exist"), LoggingSendMessageCommand)
		return
	}
	if err := cachedLogger.Send(req); err != nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "problem sending message"), LoggingSendMessageCommand)
	}

	s.serviceLoggingCacheResponse(ctx, w, nil, LoggingSendMessageCommand)
}

func (s *mdbService) serviceLoggingCacheRequest(ctx context.Context, w io.Writer, msg mongowire.Message, req interface{}, command string) jasper.LoggingCache {
	lc := s.manager.LoggingCache(ctx)
	if lc == nil {
		shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.New("logging cache not supported"), command)
		return nil
	}

	if req != nil {
		if err := shell.MessageToRequest(msg, req); err != nil {
			shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "could not read request"), command)
			return nil
		}
	}

	return lc
}

func (s *mdbService) serviceLoggingCacheResponse(ctx context.Context, w io.Writer, resp interface{}, command string) {
	if resp != nil {
		shellResp, err := shell.ResponseToMessage(mongowire.OP_REPLY, resp)
		if err != nil {
			shell.WriteErrorResponse(ctx, w, mongowire.OP_REPLY, errors.Wrap(err, "could not make response"), command)
			return
		}

		shell.WriteResponse(ctx, w, shellResp, command)
	} else {
		shell.WriteOKResponse(ctx, w, mongowire.OP_REPLY, command)
	}
}
