package options

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/evergreen-ci/birch"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/grip/send"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// CachedLogger is the cached item representing a processes normal
// output. It captures information about the cached item, as well as
// go interfaces for sending log messages.
type CachedLogger struct {
	ID       string    `bson:"id" json:"id" yaml:"id"`
	Manager  string    `bson:"manager_id" json:"manager_id" yaml:"manager_id"`
	Accessed time.Time `bson:"accessed" json:"accessed" yaml:"accessed"`

	Error  send.Sender `bson:"-" json:"-" yaml:"-"`
	Output send.Sender `bson:"-" json:"-" yaml:"-"`
}

func (cl *CachedLogger) getSender(forceError bool) (send.Sender, error) {
	if forceError && cl.Error != nil {
		return cl.Error, nil
	} else if cl.Output != nil {
		return cl.Output, nil
	} else if cl.Error != nil {
		return cl.Error, nil
	}

	return nil, errors.New("no output configured")
}

// LoggingPayload captures the arguements to the SendMessages operation.
type LoggingPayload struct {
	LoggerID         string               `bson:"logger_id" json:"logger_id" yaml:"logger_id"`
	Data             interface{}          `bson:"data" json:"data" yaml:"data"`
	Priority         level.Priority       `bson:"priority" json:"priority" yaml:"priority"`
	IsMulti          bool                 `bson:"multi" json:"multi" yaml:"multi"`
	ForceSendToError bool                 `bson:"force_send_to_error" json:"force_send_to_error" yaml:"force_send_to_error"`
	AddMetadata      bool                 `bson:"add_metadata" json:"add_metadata" yaml:"add_metadata"`
	Format           LoggingPayloadFormat `bson:"payload_format,omitempty" json:"payload_format,omitempty" yaml:"payload_format,omitempty"`
}

// LoggingPayloadFormat is an set enumerated values describing the
// formating or encoding of the payload data.
type LoggingPayloadFormat string

// notlint
const (
	LoggingPayloadFormatBSON   = "bson"
	LoggingPayloadFormatJSON   = "json"
	LoggingPayloadFormatSTRING = "string"
)

// Send resolves a sender from the cached logger (either the error or
// output endpoint,) and then sends the message from the data
// payload. This method ultimately is responsible for converting the
// payload to a message format.
func (lp *LoggingPayload) Send(logger *CachedLogger) error {
	if logger == nil {
		return errors.New("cannot send to nil logger")
	}

	sender, err := logger.getSender(lp.ForceSendToError)
	if err != nil {
		return errors.WithStack(err)
	}

	msg, err := lp.convert()
	if err != nil {
		return errors.WithStack(err)
	}

	sender.Send(msg)

	return nil
}

func (lp *LoggingPayload) convert() (message.Composer, error) {
	if lp.IsMulti {
		return lp.convertMultiMessage(lp.Data)
	}
	return lp.convertMessage(lp.Data)
}

func (lp *LoggingPayload) convertMultiMessage(value interface{}) (message.Composer, error) {
	switch data := value.(type) {
	case string:
		return lp.convertMultiMessage(strings.Split(data, "\n"))
	case []byte:
		payload, err := splitBSONSlice(lp.Format == LoggingPayloadFormatBSON, data)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return lp.convertMultiMessage(payload)
	case []string:
		batch := []message.Composer{}
		for _, str := range data {
			elem, err := lp.produceMessage([]byte(str))
			if err != nil {
				return nil, errors.WithStack(err)
			}
			batch = append(batch, elem)
		}
		return message.NewGroupComposer(batch), nil
	case [][]byte:
		batch := []message.Composer{}
		for _, dt := range data {
			elem, err := lp.produceMessage(dt)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			batch = append(batch, elem)
		}
		return message.NewGroupComposer(batch), nil
	case []interface{}:
		batch := []message.Composer{}
		for _, dt := range data {
			elem, err := lp.convertMessage(dt)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			batch = append(batch, elem)
		}
		return message.NewGroupComposer(batch), nil
	default:
		return message.ConvertToComposer(lp.Priority, value), nil
	}
}

func (lp *LoggingPayload) convertMessage(value interface{}) (message.Composer, error) {
	switch data := value.(type) {
	case string:
		return lp.produceMessage([]byte(data))
	case []byte:
		return lp.produceMessage(data)
	case []string:
		return message.ConvertToComposer(lp.Priority, data), nil
	case [][]byte:
		return message.NewLineMessage(lp.Priority, byteSlicesToStringSlice(data)), nil
	case []interface{}:
		return message.NewLineMessage(lp.Priority, data...), nil
	default:
		return message.ConvertToComposer(lp.Priority, value), nil
	}
}

func (lp *LoggingPayload) produceMessage(data []byte) (message.Composer, error) {
	switch lp.Format {
	case LoggingPayloadFormatJSON:
		payload := message.Fields{}
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, errors.Wrap(err, "problem parsing json from message body")
		}

		if lp.AddMetadata {
			return message.NewFields(lp.Priority, payload), nil
		}

		return message.NewSimpleFields(lp.Priority, payload), nil
	case LoggingPayloadFormatBSON:
		payload := message.Fields{}
		if err := bson.Unmarshal(data, &payload); err != nil {
			return nil, errors.Wrap(err, "problem parsing bson from message body")
		}

		if lp.AddMetadata {
			return message.NewFields(lp.Priority, payload), nil
		}

		return message.NewSimpleFields(lp.Priority, payload), nil
	default: // includes string case.
		if lp.AddMetadata {
			return message.NewSimpleBytesMessage(lp.Priority, data), nil
		}

		return message.NewBytesMessage(lp.Priority, data), nil
	}
}

func splitBSONSlice(isBson bool, data []byte) (interface{}, error) {
	if !isBson {
		return bytes.Split(data, []byte("\x00")), nil
	}
	out := [][]byte{}
	buf := bytes.NewBuffer(data)
	for {
		doc := birch.DC.New()
		_, err := doc.ReadFrom(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "problem reading bson from message data")
		}

		payload, err := doc.MarshalBSON()
		if err != nil {
			return nil, errors.Wrap(err, "problem constructing bson form")
		}
		out = append(out, payload)
	}
	return out, nil
}

func byteSlicesToStringSlice(in [][]byte) []string {
	out := make([]string, len(in))
	for idx := range in {
		out[idx] = string(in[idx])
	}
	return out
}
