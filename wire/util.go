package wire

import (
	"context"
	"io"
	"time"

	"github.com/evergreen-ci/birch"
	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func getProcInfoNoHang(ctx context.Context, p jasper.Process) jasper.ProcessInfo {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return p.Info(ctx)
}

func messageToResponse(msg mongowire.Message, out interface{}) error {
	doc, err := responseMessageToDocument(msg)
	if err != nil {
		return errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, out); err != nil {
		return errors.Wrap(err, "could not convert BSON to response")
	}
	return nil
}

func messageToRequest(msg mongowire.Message, out interface{}) error {
	doc, err := requestMessageToDocument(msg)
	if err != nil {
		return errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, out); err != nil {
		return errors.Wrap(err, "could not convert BSON to response")
	}
	return nil
}

// responseToMessage converts a response into a wire protocol reply.
// TODO: support OP_MSG
func responseToMessage(resp interface{}) (mongowire.Message, error) {
	b, err := bson.Marshal(resp)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	return mongowire.NewReply(0, 0, 0, 1, []birch.Document{*doc}), nil
}

// requestToMessage converts a request into a wire protocol query.
// TODO: support OP_MSG
func requestToMessage(req interface{}) (mongowire.Message, error) {
	// <namespace.$cmd  format is required to indicate that the OP_QUERY should
	// be interpreted as an OP_COMMAND.
	const namespace = "jasper.$cmd"

	b, err := bson.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	return mongowire.NewQuery(namespace, 0, 0, 1, doc, birch.NewDocument()), nil
}

// requestMessageToDocument converts a wire protocol request message into a
// document.
// TODO: support OP_MSG
func requestMessageToDocument(msg mongowire.Message) (*birch.Document, error) {
	cmdMsg, ok := msg.(*mongowire.CommandMessage)
	if !ok {
		return nil, errors.Errorf("message is not of type %s", mongowire.OP_COMMAND.String())
	}
	return cmdMsg.CommandArgs, nil
}

// responseMessageToDocument converts a wire protocol response message into a
// document.
// TODO: support OP_MSG
func responseMessageToDocument(msg mongowire.Message) (*birch.Document, error) {
	if replyMsg, ok := msg.(*mongowire.ReplyMessage); ok {
		return &replyMsg.Docs[0], nil
	}
	if cmdReplyMsg, ok := msg.(*mongowire.CommandReplyMessage); ok {
		return cmdReplyMsg.CommandReply, nil
	}
	return nil, errors.Errorf("message is not of type %s nor %s", mongowire.OP_COMMAND_REPLY.String(), mongowire.OP_REPLY.String())
}

func writeOKResponse(ctx context.Context, w io.Writer, op string) {
	resp := makeErrorResponse(true, nil)
	msg, err := resp.message()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message": "could not write response",
			"op":      op,
		}))
		return
	}
	writeResponse(ctx, w, msg, op)
}

func writeNotOKResponse(ctx context.Context, w io.Writer, op string) {
	resp := makeErrorResponse(false, nil)
	msg, err := resp.message()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message": "could not write response",
			"op":      op,
		}))
		return
	}
	writeResponse(ctx, w, msg, op)
}

func writeErrorResponse(ctx context.Context, w io.Writer, err error, op string) {
	resp := makeErrorResponse(false, err)
	msg, err := resp.message()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message": "could not write response",
			"op":      op,
		}))
		return
	}
	writeResponse(ctx, w, msg, op)
}

func writeResponse(ctx context.Context, w io.Writer, msg mongowire.Message, op string) {
	grip.Error(message.WrapError(mongowire.SendMessage(ctx, msg, w), message.Fields{
		"message": "could not write response",
		"op":      op,
	}))
}
