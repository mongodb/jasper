package remote

import (
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/send"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
)

// LoggingPayload captures the arguements to the SendMessages operation.
type LoggingPayload struct {
	LoggerID string `bson:"logger_id" json:"logger_id" yaml:"logger_id"`

	Messages         []interface{}        `bson:"messages" json:"messages" yaml:"messages"`
	Priority         level.Priority       `bson:"priority" json:"priority" yaml:"priority"`
	ForceSendToError bool                 `bson:"force_send_to_error" json:"force_send_to_error" yaml:"force_send_to_error"`
	AddMetadata      bool                 `bson:"add_metadata" json:"add_metadata" yaml:"add_metadata"`
	Format           LoggingPayloadFormat `bson:"payload_format,omitempty" json:"payload_format,omitempty" yaml:"payload_format,omitempty"`
}

type LoggingPayloadFormat string

const (
	LoggingPayloadFormatString      = "string"
	LoggingPayloadFormatStringLines = "lines"
	LoggingPayloadFormatBSON        = "bson"
	LoggingPayloadFormatJSON        = "json"
)

func (lp *LoggingPayload) Send(logger *jasper.CachedLogger) error {
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

	switch lp.Format {

	}

}
