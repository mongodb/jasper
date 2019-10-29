package mongowire

import (
	"context"
	"io"
	"time"

	"github.com/mongodb/ftdc/bsonx"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	mongorpcBson "github.com/tychoish/mongorpc/bson"
	"github.com/tychoish/mongorpc/mongowire"
	"gopkg.in/mgo.v2/bson"
)

func getProcInfoNoHang(ctx context.Context, p jasper.Process) jasper.ProcessInfo {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return p.Info(ctx)
}

func messageToDocument(msg mongowire.Message) (*bsonx.Document, error) {
	cmdMsg, ok := msg.(*mongowire.CommandMessage)
	if !ok {
		// return nil, errors.New("message is not of type %s", mongowire.OP_COMMAND.String())
		return nil, errors.New("kim: TODO: MAKE-984")
	}
	return bsonx.ReadDocument(cmdMsg.CommandArgs.BSON)
}

func procInfosToArray(ctx context.Context, procs []jasper.Process) (*bsonx.Array, error) {
	infos := bsonx.MakeArray(len(procs))
	for _, proc := range procs {
		info, err := procInfoToDocument(proc.Info(ctx))
		if err != nil {
			return infos, errors.Wrapf(err, "could not convert process info to document for process %s", proc.ID())
		}
		infos.Append(bsonx.VC.Document(info))
	}
	return infos, nil
}

func procInfoToDocument(info jasper.ProcessInfo) (*bsonx.Document, error) {
	infoBytes, err := bson.Marshal(info)
	if err != nil {
		return nil, err
	}
	return bsonx.ReadDocument(infoBytes)
}

func procTagsToArray(proc jasper.Process) (*bsonx.Array, error) {
	procTags := proc.GetTags()
	tags := bsonx.MakeArray(len(procTags))
	for _, tag := range procTags {
		tags.Append(bsonx.VC.String(tag))
	}
	return tags, nil
}

// kim: TODO: change op to mongowire.OpType
func writeErrorReply(w io.Writer, err error, op string) {
	errorDoc := bsonx.EC.String("error", err.Error())
	doc := bsonx.NewDocument(notOKResp, errorDoc)
	writeReply(w, doc, op)
}

func writeSuccessReply(w io.Writer, doc *bsonx.Document, op string) {
	doc.Prepend(okResp)
	writeReply(w, doc, op)
}

func writeReply(w io.Writer, doc *bsonx.Document, op string) {
	resp, err := doc.MarshalBSON()
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message": "could not marshal BSON response",
			"op":      op,
		}))
		return
	}

	respDoc := mongorpcBson.Simple{BSON: resp, Size: int32(len(resp))}

	reply := mongowire.NewReply(int64(0), int32(0), int32(0), int32(1), []mongorpcBson.Simple{respDoc})
	_, err = w.Write(reply.Serialize())
	grip.Error(message.WrapError(err, message.Fields{
		"message": "could not write response",
		"op":      op,
	}))
}
