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
	"gopkg.in/mgo.v2/bson"
)

func getProcInfoNoHang(ctx context.Context, p jasper.Process) jasper.ProcessInfo {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return p.Info(ctx)
}

func messageToDocument(msg mongowire.Message) (*birch.Document, error) {
	cmdMsg, ok := msg.(*mongowire.CommandMessage)
	if !ok {
		return nil, errors.Errorf("message is not of type %s", mongowire.OP_COMMAND.String())
	}
	return cmdMsg.CommandArgs, nil
}

func procInfosToArray(ctx context.Context, procs []jasper.Process) (*birch.Array, error) {
	infos := birch.MakeArray(len(procs))
	for _, proc := range procs {
		info, err := procInfoToDocument(proc.Info(ctx))
		if err != nil {
			return infos, errors.Wrapf(err, "could not convert process info to document for process %s", proc.ID())
		}
		infos.Append(birch.VC.Document(info))
	}
	return infos, nil
}

func procInfoToDocument(info jasper.ProcessInfo) (*birch.Document, error) {
	infoBytes, err := bson.Marshal(info)
	if err != nil {
		return nil, err
	}
	return birch.ReadDocument(infoBytes)
}

func procTagsToArray(proc jasper.Process) (*birch.Array, error) {
	procTags := proc.GetTags()
	tags := birch.MakeArray(len(procTags))
	for _, tag := range procTags {
		tags.Append(birch.VC.String(tag))
	}
	return tags, nil
}

func writeErrorReply(w io.Writer, err error, op string) {
	errorDoc := birch.EC.String("error", err.Error())
	doc := birch.NewDocument(notOKResp, errorDoc)
	writeReply(w, doc, op)
}

func writeSuccessReply(w io.Writer, doc *birch.Document, op string) {
	doc.Prepend(okResp)
	writeReply(w, doc, op)
}

func writeReply(w io.Writer, doc *birch.Document, op string) {
	reply := mongowire.NewReply(int64(0), int32(0), int32(0), int32(1), []*birch.Document{doc})
	_, err := w.Write(reply.Serialize())
	grip.Error(message.WrapError(err, message.Fields{
		"message": "could not write response",
		"op":      op,
	}))
}
