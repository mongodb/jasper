package options

import (
	"time"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/send"
	"github.com/pkg/errors"
)

const (
	// DefaultLogName is the default name for logs emitted by Jasper.
	DefaultLogName = "jasper"
)

// LogFormat specifies a certain format for logging by Jasper.
// See the documentation for grip/send for more information on the various
// LogFormat's.
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

type LoggerConfig struct {
	Type     string          `json:"type" bson:"type"`
	Format   LogConfigFormat `json:"format" bson:"format"`
	Data     []byte          `json:"data" bson:"data"`
	Registry LoggerRegistry  `json:"-" bson:"-"`
}

func (lc *LoggerConfig) Validate() error {
	catcher := grip.NewBasicCatcher()

	// TODO: maybe check for nil or empty Data
	if lc.Type == "" {
		catcher.New("cannot have empty logger type")
	}
	catcher.Add(lc.Format.Validate())

	if lc.Registry == nil {
		lc.Registry = GlobalLoggerRegistry
	}

	return catcher.Resolve()
}

func (lc *LoggerConfig) Resolve() (send.Sender, error) {
	if err := lc.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	factory, ok := lc.Registry.Resolve(lc.Type)
	if !ok {
		return nil, errors.Errorf("unregistered logger type '%s'", lc.Type)
	}

	if err := lc.Format.Unmarshal(lc.Data, factory); err != nil {
		return nil, errors.Wrap(err, "problem unmarshalling data")
	}

	return factory.Configure()
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
	if !opts.Level.Valid() {
		catcher.New("invalid log level")
	}

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

type bufferedSender struct {
	baseSender send.Sender

	send.Sender
}

// NewBufferedSender returns a grip send.Sender buffered with the given
// options. It overwrites the underlying Close method in order to ensure that
// both the base sender and the buffered sender are closed correctly.
func NewBufferedSender(baseSender send.Sender, opts BaseOptions) (send.Sender, err) {
	sender := send.NewBufferedSender(baseSender, opts.Base.Buffer.Duration, opts.Base.Buffer.MaxSize)
	formatter, err := opts.Base.Format.MakeFormatter()
	if err != nil {
		return nil, err
	}
	if err := sender.SetFormatter(formatter); err != nil {
		return nil, errors.New("failed to set log format")
	}

	return &bufferedSender{
		baseSender: baseSender,
		Sender:     sender,
	}
}

func (*bufferedSender) Close() error {
	catcher := grip.NewBasicCatcher()

	catcher.Wrap(l.Sender.Close(), "problem closing sender")
	catcher.Wrap(l.baseSender.Close(), "problem closing base sender")

	return catcher.Resolve()
}

///////////////////////////////////////////////////////////////////////////////
// Default Logger
///////////////////////////////////////////////////////////////////////////////

// DefaultLogger is the type name for the default logger.
const DefaultLogger = "default"

// DefaultLoggerOptions packages the options for creating a default logger.
type DefaultLoggerOptions struct {
	Prefix string      `json:"prefix" bson:"prefix"`
	Base   BaseOptions `json:"base" bson:"base"`
}

// Validate ensures DefaultLoggerOptions is valid.
func (opts *DefaultLoggerOptions) Validate() error {
	if opts.Prefix == "" {
		opts.Prefix = DefaultLogName
	}

	return opts.Base.Validate()
}

func (opts *DefaultLoggerOptions) Configure() (send.Sender, error) {
	baseSender, err := send.NewNativeLogger(opts.Prefix, opts.Base.Level)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating default logger")
	}

	bufferedSender, err := NewBufferedSender(baseSender, opts.Base)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating buffered default logger")
	}

	return bufferedSender, nil
}
