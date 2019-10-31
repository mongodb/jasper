package mongowire

import (
	"context"
	"io"
	"time"

	"github.com/evergreen-ci/birch"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc/mongowire"
)

func getProcInfoNoHang(ctx context.Context, p jasper.Process) jasper.ProcessInfo {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return p.Info(ctx)
}

func requestToDocument(msg mongowire.Message) (*birch.Document, error) {
	cmdMsg, ok := msg.(*mongowire.CommandMessage)
	if !ok {
		return nil, errors.Errorf("message is not of type %s", mongowire.OP_COMMAND.String())
	}
	return cmdMsg.CommandArgs, nil
}

func responseToDocument(msg mongowire.Message) (*birch.Document, error) {
	if replyMsg, ok := msg.(*mongowire.ReplyMessage); ok {
		return replyMsg.Docs[0], nil
	}
	if cmdReplyMsg, ok := msg.(*mongowire.CommandReplyMessage); ok {
		return cmdReplyMsg.CommandReply, nil
	}
	return nil, errors.Errorf("message is not of type %s nor %s", mongowire.OP_COMMAND_REPLY.String(), mongowire.OP_REPLY.String())
}

func writeOKReply(w io.Writer, op string) {
	resp := makeErrorResponse(true, nil)
	msg, err := resp.Message()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message": "could not write response",
			"op":      op,
		}))
		return
	}
	writeReply(w, msg, op)
}

func writeNotOKReply(w io.Writer, op string) {
	resp := makeErrorResponse(false, nil)
	msg, err := resp.Message()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message": "could not write response",
			"op":      op,
		}))
		return
	}
	writeReply(w, msg, op)
}

func writeErrorReply(w io.Writer, err error, op string) {
	resp := makeErrorResponse(false, err)
	msg, err := resp.Message()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message": "could not write response",
			"op":      op,
		}))
		return
	}
	writeReply(w, msg, op)
}

func writeReply(w io.Writer, msg mongowire.Message, op string) {
	grip.Error(message.WrapError(mongowire.SendMessage(msg, w), message.Fields{
		"message": "could not write response",
		"op":      op,
	}))
}
