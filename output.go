package jasper

import (
	"context"
	"io"

	"github.com/mongodb/grip/send"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

// NewInMemoryLogger is a basic constructor that constructs a logger
// configuration for plain formatted in-memory buffered logger. The
// logger will capture up to maxSize messages.
func NewInMemoryLogger(maxSize int) (*options.LoggerConfig, error) {
	loggerProducer := &options.InMemoryLoggerOptions{
		InMemoryCap: maxSize,
		Base: options.BaseOptions{
			Format: options.LogFormatPlain,
		},
	}
	config := &options.LoggerConfig{}
	if err := config.Set(loggerProducer); err != nil {
		return nil, errors.Wrap(err, "setting logger producer for logger config")
	}
	return config, nil
}

// LogStream represents the output of reading the in-memory log buffer as a
// stream, containing the logs (if any) and whether or not the stream is done
// reading.
type LogStream struct {
	Logs []string `bson:"logs,omitempty" json:"logs,omitempty"`
	Done bool     `bson:"done" json:"done"`
}

// GetInMemoryLogStream gets at most count logs from the in-memory output logs
// for the given Process proc. If the process has not been called with
// Process.Wait(), this is not guaranteed to produce all the logs. This function
// assumes that there is exactly one in-memory logger attached to this process's
// output. It returns io.EOF if the stream is done. For remote interfaces, this
// function will not work; use (remote.Manager).GetLogStream() instead.
func GetInMemoryLogStream(ctx context.Context, proc Process, count int) ([]string, error) {
	if proc == nil {
		return nil, errors.New("cannot get output logs from nil process")
	}
	for _, logger := range proc.Info(ctx).Options.Output.Loggers {
		if logger.Type() != options.LogInMemory {
			continue
		}

		// This is fine because logger.sender is already set.
		sender, err := logger.Resolve()
		if err != nil {
			continue
		}

		safeSender, ok := sender.(*options.SafeSender)
		if ok {
			sender = safeSender.GetSender()
		}
		inMemorySender, ok := sender.(*send.InMemorySender)
		if !ok {
			continue
		}

		msgs, _, err := inMemorySender.GetCount(count)
		if err != nil {
			if err != io.EOF {
				err = errors.Wrap(err, "getting logs from in-memory stream")
			}
			return nil, err
		}

		strs := make([]string, 0, len(msgs))
		for _, msg := range msgs {
			str, err := inMemorySender.Formatter()(msg)
			if err != nil {
				return nil, err
			}
			strs = append(strs, str)
		}

		return strs, nil
	}
	return nil, errors.New("could not find in-memory output logs")
}
