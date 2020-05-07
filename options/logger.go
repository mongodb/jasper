package options

import (
	"encoding/json"
	"time"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/send"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// DefaultLogName is the default name for logs emitted by Jasper.
	DefaultLogName = "jasper"
)

// LogFormat specifies a certain format for logging by Jasper. See the
// documentation for grip/send for more information on the various LogFormat's.
type LogFormat string

const (
	LogFormatPlain   LogFormat = "plain"
	LogFormatDefault LogFormat = "default"
	LogFormatJSON    LogFormat = "json"
	LogFormatInvalid LogFormat = "invalid"
)

// Validate ensures that the LogFormat is valid.
func (f LogFormat) Validate() error {
	switch f {
	case LogFormatDefault, LogFormatJSON, LogFormatPlain:
		return nil
	case LogFormatInvalid:
		return errors.New("invalid log format")
	default:
		return errors.New("unknown log format")
	}
}

// MakeFormatter creates a grip/send.MessageFormatter for the specified
// LogFormat on which it is called.
func (f LogFormat) MakeFormatter() (send.MessageFormatter, error) {
	switch f {
	case LogFormatDefault:
		return send.MakeDefaultFormatter(), nil
	case LogFormatPlain:
		return send.MakePlainFormatter(), nil
	case LogFormatJSON:
		return send.MakeJSONFormatter(), nil
	case LogFormatInvalid:
		return nil, errors.New("cannot make log format for invalid format")
	default:
		return nil, errors.New("unknown log format")
	}
}

// LogConfig type aliases []byte in order to implement the bson.MarshalBSON and
// json.MarshalJSON interfaces. It is intended to hold the logger configuration
// data.
type LogConfig []byte

func (lc LogConfig) MarshalBSON() ([]byte, error) { return lc, nil }
func (lc LogConfig) MarshalJSON() ([]byte, error) { return lc, nil }

// LoggerConfig represents the necessary information to construct a new grip
// send.Sender. LoggerConfig implements the bson.MarshalBSON and
// json.MarshalJSON interfaces.
type LoggerConfig struct {
	Type     string          `json:"type" bson:"type"`
	Format   LogConfigFormat `json:"format" bson:"format"`
	Config   LogConfig       `json:"config" bson:"config"`
	Registry LoggerRegistry  `json:"-" bson:"-"`

	sender send.Sender
}

// Validate ensures the LoggerConfig is valid.
func (lc *LoggerConfig) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(lc.Type == "", "cannot have empty logger type")
	catcher.Add(lc.Format.Validate())

	if lc.Registry == nil {
		lc.Registry = globalLoggerRegistry
	}

	return catcher.Resolve()
}

// Resolve resolves the LoggerConfig and returns the results grip send.Sender.
func (lc *LoggerConfig) Resolve() (send.Sender, error) {
	if lc.sender == nil {
		if err := lc.Validate(); err != nil {
			return nil, errors.Wrap(err, "invalid config")
		}

		factory, ok := lc.Registry.Resolve(lc.Type)
		if !ok {
			return nil, errors.Errorf("unregistered logger type '%s'", lc.Type)
		}
		producer := factory()

		if len(lc.Config) > 0 {
			if err := lc.Format.unmarshal(lc.Config, producer); err != nil {
				return nil, errors.Wrap(err, "problem unmarshalling data")
			}
		}

		sender, err := producer.Configure()
		if err != nil {
			return nil, err
		}
		lc.sender = sender
	}

	return lc.sender, nil
}

func (lc *LoggerConfig) MarshalBSON() ([]byte, error) {
	if lc.Format == "" {
		lc.Format = LogConfigFormatBSON
	}
	if lc.Format != LogConfigFormatBSON {
		return nil, errors.New("cannot marshal misconfigured bson logger config")
	}

	return bson.Marshal(lc)
}

func (lc *LoggerConfig) MarshalJSON() ([]byte, error) {
	if lc.Format == "" {
		lc.Format = LogConfigFormatJSON
	}
	if lc.Format != LogConfigFormatJSON {
		return nil, errors.New("cannot marshal misconfigured bson logger config")
	}

	return json.Marshal(lc)
}

// BaseOptions are the base options necessary for setting up most loggers.
type BaseOptions struct {
	Level  send.LevelInfo
	Buffer BufferOptions
	Format LogFormat
}

// Validate ensures that BaseOptions is valid.
func (opts *BaseOptions) Validate() error {
	catcher := grip.NewBasicCatcher()

	if opts.Level.Threshold == 0 && opts.Level.Default == 0 {
		opts.Level = send.LevelInfo{Default: level.Trace, Threshold: level.Trace}
	}

	catcher.NewWhen(!opts.Level.Valid(), "invalid log level")
	catcher.Wrap(opts.Buffer.Validate(), "invalid buffering options")
	catcher.Add(opts.Format.Validate())
	return catcher.Resolve()
}

// BufferOptions packages options for whether or not a Logger should be
// buffered and the duration and size of the respective buffer in the case that
// it should be.
type BufferOptions struct {
	Buffered bool          `bson:"buffered" json:"buffered" yaml:"buffered"`
	Duration time.Duration `bson:"duration" json:"duration" yaml:"duration"`
	MaxSize  int           `bson:"max_size" json:"max_size" yaml:"max_size"`
}

// Validate ensures that BufferOptions is valid.
func (opts *BufferOptions) Validate() error {
	if opts.Buffered && opts.Duration < 0 || opts.MaxSize < 0 {
		return errors.New("cannot have negative buffer duration or size")
	}

	return nil
}

type safeSender struct {
	baseSender send.Sender
	send.Sender
}

// NewSafeSender returns a grip send.Sender with the given base options. It
// overwrites the underlying Close method in order to ensure that both the base
// sender and buffered sender are closed correctly.
func NewSafeSender(baseSender send.Sender, opts BaseOptions) (send.Sender, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid options")
	}

	sender := &safeSender{}
	if opts.Buffer.Buffered {
		sender.Sender = send.NewBufferedSender(baseSender, opts.Buffer.Duration, opts.Buffer.MaxSize)
		sender.baseSender = baseSender
	} else {
		sender.Sender = baseSender
	}

	formatter, err := opts.Format.MakeFormatter()
	if err != nil {
		return nil, err
	}
	if err := sender.SetFormatter(formatter); err != nil {
		return nil, errors.New("failed to set log format")
	}

	return sender, nil
}

func (s *safeSender) Close() error {
	catcher := grip.NewBasicCatcher()

	catcher.Wrap(s.Sender.Close(), "problem closing sender")
	if s.baseSender != nil {
		catcher.Wrap(s.baseSender.Close(), "problem closing base sender")
	}

	return catcher.Resolve()
}
