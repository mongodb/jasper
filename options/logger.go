package options

import (
	"time"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/send"
	"github.com/pkg/errors"
)

var GlobalLoggerRegistry LoggerRegistry = &basicLoggerRegistry{
	factories: map[string]LoggerProducerFactory{
		DefaultLogger:   NewDefaultLoggerProducer,
		FileLogger:      NewFileLoggerProducer,
		InheritedLogger: NewInheritedLoggerProducer,
		SumoLogicLogger: NewSumoLogicLoggerProducer,
		InMemoryLogger:  NewInMemoryLoggerProducer,
		SplunkLogger:    NewSplunkLoggerProducer,
		BuildloggerV2:   NewBuildloggerV2LoggerProducer,
	},
}

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

// LoggerConfig represents the necessary information to construct a new grip
// send.Sender.
type LoggerConfig struct {
	Type     string          `json:"type" bson:"type"`
	Format   LogConfigFormat `json:"format" bson:"format"`
	Data     []byte          `json:"data" bson:"data"`
	Registry LoggerRegistry  `json:"-" bson:"-"`
}

// Validate ensures the LoggerConfig is valid.
func (lc *LoggerConfig) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(lc.Type == "", "cannot have empty logger type")
	catcher.Add(lc.Format.Validate())

	if lc.Registry == nil {
		lc.Registry = GlobalLoggerRegistry
	}

	return catcher.Resolve()
}

// Resolve resolves the LoggerConfig and returns the results grip send.Sender.
func (lc *LoggerConfig) Resolve() (send.Sender, error) {
	if err := lc.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	factory, ok := lc.Registry.Resolve(lc.Type)
	if !ok {
		return nil, errors.Errorf("unregistered logger type '%s'", lc.Type)
	}
	producer := factory()

	if len(lc.Data) > 0 {
		if err := lc.Format.unmarshal(lc.Data, producer); err != nil {
			return nil, errors.Wrap(err, "problem unmarshalling data")
		}
	}

	return producer.Configure()
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

// NewDefaultLoggerProducer returns a LoggerProducer backed by
// DefaultLoggerOptions.
func NewDefaultLoggerProducer() LoggerProducer { return &DefaultLoggerOptions{} }

// Validate ensures DefaultLoggerOptions is valid.
func (opts *DefaultLoggerOptions) Validate() error {
	if opts.Prefix == "" {
		opts.Prefix = DefaultLogName
	}

	return opts.Base.Validate()
}

func (*DefaultLoggerOptions) Type() string { return DefaultLogger }
func (opts *DefaultLoggerOptions) Configure() (send.Sender, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid options")
	}

	sender, err := send.NewNativeLogger(opts.Prefix, opts.Base.Level)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating base default logger")
	}

	sender, err = NewSafeSender(sender, opts.Base)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating safe default logger")
	}
	return sender, nil
}

///////////////////////////////////////////////////////////////////////////////
// File Logger
///////////////////////////////////////////////////////////////////////////////

// FileLogger is the type name for the file logger.
const FileLogger = "file"

// FileLoggerOptions packages the options for creating a file logger.
type FileLoggerOptions struct {
	FileName string      `json:"file_name " bson:"file_name"`
	Base     BaseOptions `json:"base" bson:"base"`
}

// NewFileLoggerProducer returns a LoggerProducer backed by FileLoggerOptions.
func NewFileLoggerProducer() LoggerProducer { return &FileLoggerOptions{} }

// Validate ensures FileLoggerOptions is valid.
func (opts *FileLoggerOptions) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(opts.FileName == "", "must specify a filename")
	catcher.Add(opts.Base.Validate())
	return catcher.Resolve()
}

func (*FileLoggerOptions) Type() string { return FileLogger }
func (opts *FileLoggerOptions) Configure() (send.Sender, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid options")
	}

	sender, err := send.NewPlainFileLogger(DefaultLogName, opts.FileName, opts.Base.Level)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating base file logger")
	}

	sender, err = NewSafeSender(sender, opts.Base)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating safe file logger")
	}
	return sender, nil
}

///////////////////////////////////////////////////////////////////////////////
// Inherited Logger
///////////////////////////////////////////////////////////////////////////////

// InheritedLogger is the type name for the inherited logger.
const InheritedLogger = ""

// InheritLoggerOptions packages the options for creating an inherited logger.
type InheritedLoggerOptions struct {
	Base BaseOptions `json:"base" bson:"base"`
}

// NewInheritedLoggerProducer returns a LoggerProducer backed by
// InheritedLoggerOptions.
func NewInheritedLoggerProducer() LoggerProducer { return &InheritedLoggerOptions{} }

func (*InheritedLoggerOptions) Type() string { return InheritedLogger }
func (opts *InheritedLoggerOptions) Configure() (send.Sender, error) {
	var (
		sender send.Sender
		err    error
	)

	if err = opts.Base.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid options")
	}

	sender = grip.GetSender()
	if err = sender.SetLevel(opts.Base.Level); err != nil {
		return nil, errors.Wrap(err, "problem creating base inherited logger")
	}

	sender, err = NewSafeSender(sender, opts.Base)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating safe inherited logger")
	}
	return sender, nil
}

///////////////////////////////////////////////////////////////////////////////
// In Memory Logger
///////////////////////////////////////////////////////////////////////////////

// InMemoryLogger is the type name for the in memory logger.
const InMemoryLogger = "in-memory"

// InMemoryLoggerOptions packages the options for creating an in memory logger.
type InMemoryLoggerOptions struct {
	InMemoryCap int         `json:"in_memory_cap" bson:"in_memory_cap"`
	Base        BaseOptions `json:"base" bson:"base"`
}

// NewInMemoryLoggerProducer returns a LoggerProducer backed by
// InMemoryLoggerOptions.
func NewInMemoryLoggerProducer() LoggerProducer { return &InMemoryLoggerOptions{} }

func (opts *InMemoryLoggerOptions) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(opts.InMemoryCap <= 0, "invalid in-memory capacity")
	catcher.Add(opts.Base.Validate())
	return catcher.Resolve()
}

func (*InMemoryLoggerOptions) Type() string { return InMemoryLogger }
func (opts *InMemoryLoggerOptions) Configure() (send.Sender, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	sender, err := send.NewInMemorySender(DefaultLogName, opts.Base.Level, opts.InMemoryCap)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating base in-memory logger")
	}

	sender, err = NewSafeSender(sender, opts.Base)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating safe in-memory logger")
	}
	return sender, nil
}

///////////////////////////////////////////////////////////////////////////////
// Sumo Logic Logger
///////////////////////////////////////////////////////////////////////////////

// SumoLogicLogger is the type name for the sumo logic logger.
const SumoLogicLogger = "sumo-logic"

// SumoLogicLoggerOptions packages the options for creating a sumo logic
// logger.
type SumoLogicLoggerOptions struct {
	SumoEndpoint string      `json:"sumo_endpoint" bson:"sumo_endpoint"`
	Base         BaseOptions `json:"base" bson:"base"`
}

// SumoLogicLoggerProducer returns a LoggerProducer backed by
// SumoLogicLoggerOptions.
func NewSumoLogicLoggerProducer() LoggerProducer { return &SumoLogicLoggerOptions{} }

func (opts *SumoLogicLoggerOptions) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(opts.SumoEndpoint == "", "must specify a sumo endpoint")
	catcher.Add(opts.Base.Validate())
	return catcher.Resolve()
}

func (*SumoLogicLoggerOptions) Type() string { return SumoLogicLogger }
func (opts *SumoLogicLoggerOptions) Configure() (send.Sender, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	sender, err := send.NewSumo(DefaultLogName, opts.SumoEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating base sumo logic logger")
	}
	if err = sender.SetLevel(opts.Base.Level); err != nil {
		return nil, errors.Wrap(err, "problem setting level for sumo logic logger")
	}

	sender, err = NewSafeSender(sender, opts.Base)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating safe sumo logic logger")
	}
	return sender, nil
}

///////////////////////////////////////////////////////////////////////////////
// Splunk Logger
///////////////////////////////////////////////////////////////////////////////

// SplunkLogger is the type name for the splunk logger.
const SplunkLogger = "splunk"

// SplunkLoggerOptions packages the options for creating a splunk logger.
type SplunkLoggerOptions struct {
	Splunk send.SplunkConnectionInfo `json:"splunk" bson:"splunk"`
	Base   BaseOptions               `json:"base" bson:"base"`
}

// SplunkLoggerProducer returns a LoggerProducer backed by SplunkLoggerOptions.
func NewSplunkLoggerProducer() LoggerProducer { return &SplunkLoggerOptions{} }

func (opts *SplunkLoggerOptions) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(opts.Splunk.Populated(), "missing connection info for output type splunk")
	catcher.Add(opts.Base.Validate())
	return catcher.Resolve()
}

func (*SplunkLoggerOptions) Type() string { return SplunkLogger }
func (opts *SplunkLoggerOptions) Configure() (send.Sender, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	sender, err := send.NewSplunkLogger(DefaultLogName, opts.Splunk, opts.Base.Level)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating base splunk logger")
	}

	sender, err = NewSafeSender(sender, opts.Base)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating safe splunk logger")
	}
	return sender, nil
}

///////////////////////////////////////////////////////////////////////////////
// BuildloggerV2 Logger
///////////////////////////////////////////////////////////////////////////////

// BuildloggerV2 is the type name for the buildlogger v2 logger.
const BuildloggerV2 = "buildloggerv2"

// DefaultLoggerOptions packages the options for creating a default logger.
type BuildloggerV2Options struct {
	Buildlogger send.BuildloggerConfig `json:"buildlogger" bson:"buildlogger"`
	Base        BaseOptions            `json:"base" bson:"base"`
}

// NewBuildloggerV2LoggerProducer returns a LoggerProducer backed by
// BuildloggerV2Options.
func NewBuildloggerV2LoggerProducer() LoggerProducer { return &BuildloggerV2Options{} }

func (*BuildloggerV2Options) Type() string { return BuildloggerV2 }
func (opts *BuildloggerV2Options) Configure() (send.Sender, error) {
	if opts.Buildlogger.Local == nil {
		opts.Buildlogger.Local = send.MakeNative()
	}
	if opts.Buildlogger.Local.Name() == "" {
		opts.Buildlogger.Local.SetName(DefaultLogName)
	}

	sender, err := send.NewBuildlogger(DefaultLogName, &opts.Buildlogger, opts.Base.Level)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating base buildlogger logger")
	}

	sender, err = NewSafeSender(sender, opts.Base)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating safe buildlogger logger")
	}
	return sender, nil
}
