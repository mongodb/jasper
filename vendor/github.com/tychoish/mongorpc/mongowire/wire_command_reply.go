package mongowire

import (
	"github.com/evergreen-ci/birch"
	"github.com/pkg/errors"
)

func (m *commandReplyMessage) HasResponse() bool     { return false }
func (m *commandReplyMessage) Header() MessageHeader { return m.header }
func (m *commandReplyMessage) Scope() *OpScope       { return nil }

func (m *commandReplyMessage) Serialize() []byte {
	size := 16 /* header */

	size += getDocSize(m.CommandReply)
	size += getDocSize(m.Metadata)
	for _, d := range m.OutputDocs {
		size += getDocSize(&d)
	}
	m.header.Size = int32(size)

	buf := make([]byte, size)
	m.header.WriteInto(buf)

	loc := 16
	loc += writeDocAt(loc, m.CommandReply, buf)
	loc += writeDocAt(loc, m.Metadata, buf)

	for _, d := range m.OutputDocs {
		loc += writeDocAt(loc, &d, buf)
	}

	return buf
}

func (h *MessageHeader) parseCommandReplyMessage(buf []byte) (Message, error) {
	rm := &commandReplyMessage{
		header: *h,
	}

	var err error

	rm.CommandReply, err = birch.ReadDocument(buf)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	replySize := getDocSize(rm.CommandReply)
	if len(buf) < replySize {
		return nil, errors.New("invalid command message -- message length is too short")
	}
	buf = buf[replySize:]

	rm.Metadata, err = birch.ReadDocument(buf)
	if err != nil {
		return nil, err
	}
	metaSize := getDocSize(rm.Metadata)
	if len(buf) < metaSize {
		return nil, errors.New("invalid command message -- message length is too short")
	}
	buf = buf[metaSize:]

	for len(buf) > 0 {
		doc, err := birch.ReadDocument(buf)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		buf = buf[getDocSize(doc):]
		rm.OutputDocs = append(rm.OutputDocs, *doc.Copy())
	}

	return rm, nil
}
