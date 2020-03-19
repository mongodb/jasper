package options

import (
	"encoding/json"
	"time"

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

type LoggingPayloadFormat string

const (
	LoggingPayloadFormatBSON   = "bson"
	LoggingPayloadFormatJSON   = "json"
	LoggingPayloadFormatSTRING = "string"
)

func (lp *LoggingPayload) Send(logger *CachedLogger) error {
	if logger == nil || (logger.Error == nil && logger.Output == nil) {
		return errors.New("no output configured")
	}

	var sender send.Sender

	if lp.ForceSendToError && logger.Error != nil {
		sender = logger.Error
	} else if logger.Output != nil {
		sender = logger.Output
	} else if logger.Error != nil {
		sender = logger.Error
	} else {
		return errors.New("could not configure output for message")
	}

	msg, err := lp.convertMessage(lp.Data)
	if err != nil {
		return errors.WithStack(err)
	}

	sender.Send(msg)

	return nil
}

func (lp *LoggingPayload) convertMessage(value interface{}) (message.Composer, error) {
	switch data := value.(type) {
	case string:
		return lp.produceMessage([]byte(data))
	case []byte:
		return lp.produceMessage(data)
	case []string:
		if lp.IsMulti {
			batch := []message.Composer{}
			for _, str := range data {
				elem, err := lp.produceMessage([]byte(str))
				if err != nil {
					return nil, errors.WithStack(err)
				}
				batch = append(batch, elem)
			}
			return message.NewGroupComposer(batch), nil
		}
		return message.ConvertToComposer(lp.Priority, data), nil
	case [][]byte:
		if lp.IsMulti {
			batch := []message.Composer{}
			for _, dt := range data {
				elem, err := lp.produceMessage(dt)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				batch = append(batch, elem)
			}
			return message.NewGroupComposer(batch), nil
		}

		return message.NewLineMessage(lp.Priority, byteSlicesToStringSlice(data)), nil
	case []interface{}:
		if lp.IsMulti {
			batch := []message.Composer{}
			for _, dt := range data {
				elem, err := lp.convertMessage(dt)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				batch = append(batch, elem)
			}
			return message.NewGroupComposer(batch), nil
		}
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
			return nil, errors.Wrap(err, "problem parsing json from message body")
		}

		if lp.AddMetadata {
			return message.NewFields(lp.Priority, payload), nil
		}

		return message.NewSimpleFields(lp.Priority, payload), nil

	default:
		if lp.AddMetadata {
			return message.NewSimpleBytesMessage(lp.Priority, data), nil
		}

		return message.NewBytesMessage(lp.Priority, data), nil
	}
}

func byteSlicesToStringSlice(in [][]byte) []string {
	out := make([]string, len(in))
	for idx := range in {
		out[idx] = string(in[idx])
	}
	return out
}
