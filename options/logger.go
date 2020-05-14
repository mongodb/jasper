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
// documentation for grip/send for more information on the various LogFormats.
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
		return nil, errors.Errorf("unknown log format '%s'", f)
	}
}

// RawLoggerConfigFormat describes the format of the raw logger configuration.
type RawLoggerConfigFormat string

const (
	RawLoggerConfigFormatBSON    RawLoggerConfigFormat = "BSON"
	RawLoggerConfigFormatJSON    RawLoggerConfigFormat = "JSON"
	RawLoggerConfigFormatInvalid RawLoggerConfigFormat = "invalid"
)

// Validate ensures that RawLoggerConfigFormat is valid.
func (f RawLoggerConfigFormat) Validate() error {
	switch f {
	case RawLoggerConfigFormatBSON, RawLoggerConfigFormatJSON:
		return nil
	case RawLoggerConfigFormatInvalid:
		return errors.New("invalid log format")
	default:
		return errors.Errorf("unknown raw logger config format '%s'", f)
	}
}

func (f RawLoggerConfigFormat) marshal(in LoggerProducer) ([]byte, error) {
	switch f {
	case RawLoggerConfigFormatBSON:
		return bson.Marshal(in)
	case RawLoggerConfigFormatJSON:
		return json.Marshal(in)
	default:
		return nil, errors.Errorf("unsupported format '%s'", f)
	}
}

func (f RawLoggerConfigFormat) Unmarshal(data []byte, out interface{}) error {
	switch f {
	case RawLoggerConfigFormatBSON:
		if err := bson.Unmarshal(data, out); err != nil {
			return errors.Wrapf(err, "could not render '%s' input into '%s'", data, out)

		}
	case RawLoggerConfigFormatJSON:
		if err := json.Unmarshal(data, out); err != nil {
			return errors.Wrapf(err, "could not render '%s' input into '%s'", data, out)

		}
	default:
		return errors.Errorf("unsupported format '%s'", f)
	}

	return nil
}

type RawLoggerConfig []byte

func (lc *RawLoggerConfig) MarshalJSON() ([]byte, error) {
	if len(*lc) > 0 {
		return *lc, nil
	}

	return json.Marshal([]byte(*lc))
}

func (lc *RawLoggerConfig) UnmarshalJSON(b []byte) error {
	*lc = b
	return nil
}

func (lc RawLoggerConfig) MarshalBSON() ([]byte, error) {
	if len(lc) > 0 {
		return lc, nil
	}

	return bson.Marshal([]byte(lc))
}

func (lc *RawLoggerConfig) UnmarshalBSON(b []byte) error {
	if len(b) > 0 {
		*lc = b
	} else {
		*lc = []byte{}
	}
	return nil
}

// LoggerConfig represents the necessary information to construct a new grip
// send.Sender. LoggerConfig implements the bson.MarshalBSON and
// json.MarshalJSON interfaces.
type LoggerConfig struct {
	info     loggerConfigInfo
	registry LoggerRegistry

	producer LoggerProducer
	sender   send.Sender
}

type loggerConfigInfo struct {
	Type   string                `json:"type" bson:"type"`
	Format RawLoggerConfigFormat `json:"format" bson:"format"`
	Config RawLoggerConfig       `json:"config" bson:"config"`
}

// NewLoggerConfig is an entry point for creating a LoggerConfig with a raw
// config. Typically, LoggerConfigs are created directly and setup with a
// LoggerProducer using the Set method. This function does not validate the
// config.
func NewLoggerConfig(producerType string, format RawLoggerConfigFormat, config []byte) *LoggerConfig {
	return &LoggerConfig{
		info: loggerConfigInfo{
			Type:   producerType,
			Format: format,
			Config: config,
		},
	}
}

func (lc *LoggerConfig) validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(lc.info.Type == "", "cannot have empty logger type")
	if len(lc.info.Config) > 0 {
		catcher.Add(lc.info.Format.Validate())
	}

	if lc.registry == nil {
		lc.registry = globalLoggerRegistry
	}

	return catcher.Resolve()
}

// Set sets the logger producer and type for the logger config.
func (lc *LoggerConfig) Set(producer LoggerProducer) error {
	if lc.registry == nil {
		lc.registry = globalLoggerRegistry
	}

	if !lc.registry.Check(producer.Type()) {
		return errors.New("unregistered logger producer")
	}

	lc.info.Type = producer.Type()
	lc.producer = producer

	return nil
}

// Producer returns the underlying logger producer for this logger config,
// which may be nil.
func (lc *LoggerConfig) Producer() LoggerProducer { return lc.producer }

// Type returns the type string.
func (lc *LoggerConfig) Type() string { return lc.info.Type }

// Config returns the raw logger config.
func (lc *LoggerConfig) Config() RawLoggerConfig { return lc.info.Config }

// Resolve resolves the LoggerConfig and returns the resulting grip
// send.Sender.
func (lc *LoggerConfig) Resolve() (send.Sender, error) {
	if lc.sender == nil {
		if err := lc.resolveProducer(); err != nil {
			return nil, errors.Wrap(err, "problem resolve logger producer")
		}

		sender, err := lc.producer.Configure()
		if err != nil {
			return nil, err
		}
		lc.sender = sender
	}

	return lc.sender, nil
}

func (lc *LoggerConfig) MarshalBSON() ([]byte, error) {
	if err := lc.resolveProducer(); err != nil {
		return nil, errors.Wrap(err, "problem resolving logger producer")
	}

	data, err := bson.Marshal(lc.producer)
	if err != nil {
		return nil, errors.Wrap(err, "problem producing logger config")
	}

	return bson.Marshal(&loggerConfigInfo{
		Type:   lc.producer.Type(),
		Format: RawLoggerConfigFormatBSON,
		Config: data,
	})
}

func (lc *LoggerConfig) UnmarshalBSON(b []byte) error {
	info := loggerConfigInfo{}
	if err := bson.Unmarshal(b, &info); err != nil {
		return errors.Wrap(err, "problem unmarshalling config logger info")
	}

	lc.info = info
	return nil
}

func (lc *LoggerConfig) MarshalJSON() ([]byte, error) {
	if err := lc.resolveProducer(); err != nil {
		return nil, errors.Wrap(err, "problem resolving logger producer")
	}

	data, err := json.Marshal(lc.producer)
	if err != nil {
		return nil, errors.Wrap(err, "problem producing logger config")
	}

	return json.Marshal(&loggerConfigInfo{
		Type:   lc.producer.Type(),
		Format: RawLoggerConfigFormatJSON,
		Config: data,
	})
}

func (lc *LoggerConfig) UnmarshalJSON(b []byte) error {
	info := loggerConfigInfo{}
	if err := json.Unmarshal(b, &info); err != nil {
		return errors.Wrap(err, "problem unmarshalling config logger info")
	}

	lc.info = info
	return nil
}

func (lc *LoggerConfig) resolveProducer() error {
	if lc.producer == nil {
		if err := lc.validate(); err != nil {
			return errors.Wrap(err, "invalid logger config")
		}

		factory, ok := lc.registry.Resolve(lc.info.Type)
		if !ok {
			return errors.Errorf("unregistered logger type '%s'", lc.info.Type)
		}
		lc.producer = factory()

		if len(lc.info.Config) > 0 {
			if err := lc.info.Format.Unmarshal(lc.info.Config, lc.producer); err != nil {
				return errors.Wrap(err, "problem unmarshalling data")
			}
		}
	}

	return nil
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

type SafeSender struct {
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

	sender := &SafeSender{}
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

func (s *SafeSender) GetSender() send.Sender {
	if s.baseSender != nil {
		return s.baseSender
	}

	return s.Sender
}

func (s *SafeSender) Close() error {
	catcher := grip.NewBasicCatcher()

	catcher.Wrap(s.Sender.Close(), "problem closing sender")
	if s.baseSender != nil {
		catcher.Wrap(s.baseSender.Close(), "problem closing base sender")
	}

	return catcher.Resolve()
}
