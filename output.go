package jasper

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/send"
	"github.com/pkg/errors"
)

// OutputOptions provides a common way to define and represent the
// output behavior of a evergreen/subprocess.Command operation.
type OutputOptions struct {
	Output            io.Writer `json:"-"`
	Error             io.Writer `json:"-"`
	SuppressOutput    bool      `json:"suppress_output"`
	SuppressError     bool      `json:"suppress_error"`
	SendOutputToError bool      `json:"redirect_output_to_error"`
	SendErrorToOutput bool      `json:"redirect_error_to_output"`
	Loggers           []Logger  `json:"loggers"`
	outputSender      *send.WriterSender
	errorSender       *send.WriterSender
	outputMulti       io.Writer
	errorMulti        io.Writer
}

type LogType string

const (
	LogBuildloggerV2 = "buildloggerv2"
	LogBuildloggerV3 = "buildloggerv3"
	LogDefault       = "default"
	LogFile          = "file"
	LogInherit       = "inherit"
	LogSplunk        = "splunk"
	LogSumologic     = "sumologic"
	LogInMemory      = "inmemory"
)

const (
	DefaultLogName = "jasper"
)

// By default, logger reads from both standard output and standard error.
type LogOptions struct {
	BuildloggerOptions send.BuildloggerConfig    `json:"buildlogger_options"`
	DefaultPrefix      string                    `json:"default_prefix"`
	FileName           string                    `json:"file_name"`
	SplunkOptions      send.SplunkConnectionInfo `json:"splunk_options"`
	SumoEndpoint       string                    `json:"sumo_endpoint"`
	InMemoryCap        int                       `json:"inmemory_cap"`
}

type Logger struct {
	Type    LogType    `json:"log_type"`
	Options LogOptions `json:"log_options"`
}

func (o OutputOptions) outputIsNull() bool {
	if o.Output == nil {
		return true
	}

	if o.Output == ioutil.Discard {
		return true
	}

	return false
}

func (o OutputOptions) outputLogging() bool {
	return len(o.Loggers) > 0 && !o.SuppressOutput
}

func (o OutputOptions) errorLogging() bool {
	return len(o.Loggers) > 0 && !o.SuppressError
}

func (o OutputOptions) errorIsNull() bool {
	if o.Error == nil {
		return true
	}

	if o.Error == ioutil.Discard {
		return true
	}

	return false
}

func (l LogType) Validate() error {
	switch l {
	case LogBuildloggerV2, LogBuildloggerV3, LogDefault, LogFile, LogInherit, LogSplunk, LogSumologic, LogInMemory:
		return nil
	default:
		return errors.New("unknown log type")
	}
}

func makeDevNullSender() (send.Sender, error) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return send.NewStreamLogger("", devNull, send.LevelInfo{Default: level.Trace, Threshold: level.Trace})
}

func (l LogType) Configure(opts LogOptions) (*send.Sender, error) {
	var sender *send.Sender

	switch l {
	case LogBuildloggerV2, LogBuildloggerV3:
		if opts.BuildloggerOptions.Local == nil {
			return nil, errors.New("must specify buildlogger local sender")
		}
		if opts.BuildloggerOptions.Local.Name() == "" {
			return nil, errors.New("must call SetName() on buildlogger local sender")
		}
		s, err := send.MakeBuildlogger(DefaultLogName, &opts.BuildloggerOptions)
		if err != nil {
			return nil, err
		}
		sender = &s
	case LogDefault:
		s, err := send.NewNativeLogger(DefaultLogName, send.LevelInfo{Default: level.Trace, Threshold: level.Trace})
		if err != nil {
			return nil, err
		}
		sender = &s
	case LogFile:
		s, err := send.MakePlainFileLogger(opts.FileName)
		if err != nil {
			return nil, err
		}
		s.SetName(DefaultLogName)
		s.SetFormatter(send.MakePlainFormatter())
		sender = &s
	case LogInherit:
		s := grip.GetSender()
		sender = &s
	case LogSplunk:
		if !opts.SplunkOptions.Populated() {
			return nil, errors.New("missing connection info for output type splunk")
		}
		s, err := send.NewSplunkLogger("", opts.SplunkOptions, send.LevelInfo{Default: level.Trace, Threshold: level.Trace})
		if err != nil {
			return nil, err
		}
		sender = &s
	case LogSumologic:
		if opts.SumoEndpoint == "" {
			return nil, errors.New("missing endpoint for output type sumologic")
		}
		s, err := send.NewSumo(DefaultLogName, opts.SumoEndpoint)
		if err != nil {
			return nil, err
		}
		sender = &s
	case LogInMemory:
		if opts.InMemoryCap <= 0 {
			return nil, errors.New("invalid inmemory capacity")
		}
		s, err := send.NewInMemorySender("jasper", send.LevelInfo{Default: level.Trace, Threshold: level.Trace}, opts.InMemoryCap)
		if err != nil {
			return nil, err
		}
		sender = &s
	default:
		return nil, errors.New("unknown log type")
	}

	return sender, nil
}

func (o OutputOptions) Validate() error {
	catcher := grip.NewBasicCatcher()

	if o.SuppressOutput && (!o.outputIsNull() || o.outputLogging()) {
		catcher.Add(errors.New("cannot suppress output if output is defined"))
	}

	if o.SuppressError && (!o.errorIsNull() || o.errorLogging()) {
		catcher.Add(errors.New("cannot suppress error if error is defined"))
	}

	if o.Error == o.Output && !o.errorIsNull() {
		catcher.Add(errors.New("cannot specify the same value for error and output"))
	}

	if o.SuppressOutput && o.SendOutputToError {
		catcher.Add(errors.New("cannot suppress output and redirect it to error"))
	}

	if o.SuppressError && o.SendErrorToOutput {
		catcher.Add(errors.New("cannot suppress error and redirect it to output"))
	}

	if o.SendOutputToError && o.Error == nil && !o.errorLogging() {
		catcher.Add(errors.New("cannot redirect output to error without a defined error writer"))
	}

	if o.SendErrorToOutput && o.Output == nil && !o.outputLogging() {
		catcher.Add(errors.New("cannot redirect error to output without a defined output writer"))
	}

	if o.SendOutputToError && o.SendErrorToOutput {
		catcher.Add(errors.New("cannot create redirect cycle between output and error"))
	}

	for _, logger := range o.Loggers {
		if err := logger.Type.Validate(); err != nil {
			catcher.Add(err)
		}
	}

	return catcher.Resolve()
}

func (o *OutputOptions) GetOutput() (io.Writer, error) {
	if o.SendOutputToError {
		return o.GetError()
	}

	if o.outputIsNull() && !o.outputLogging() {
		return ioutil.Discard, nil
	}

	if o.outputMulti != nil {
		return o.outputMulti, nil
	}

	if o.outputLogging() {
		outSenders := []send.Sender{}

		for _, logger := range o.Loggers {
			sender, err := logger.Type.Configure(logger.Options)
			if err != nil {
				return ioutil.Discard, err
			}
			outSenders = append(outSenders, *sender)
		}

		var outMulti send.Sender
		if len(outSenders) == 1 {
			outMulti = outSenders[0]
		} else {
			var err error
			outMulti, err = send.NewMultiSender(DefaultLogName, send.LevelInfo{Default: level.Info, Threshold: level.Trace}, outSenders)
			if err != nil {
				return ioutil.Discard, err
			}
		}
		o.outputSender = send.NewWriterSender(outMulti)
	}

	if !o.outputIsNull() && o.outputLogging() {
		o.outputMulti = io.MultiWriter(o.Output, o.outputSender)
	} else if !o.outputIsNull() {
		o.outputMulti = o.Output
	} else {
		o.outputMulti = o.outputSender
	}

	return o.outputMulti, nil
}

func (o *OutputOptions) GetError() (io.Writer, error) {
	if o.SendErrorToOutput {
		return o.GetOutput()
	}

	if o.errorIsNull() && !o.errorLogging() {
		return ioutil.Discard, nil
	}

	if o.errorMulti != nil {
		return o.errorMulti, nil
	}

	if o.errorLogging() {
		errSenders := []send.Sender{}

		for _, logger := range o.Loggers {
			sender, err := logger.Type.Configure(logger.Options)
			if err != nil {
				return ioutil.Discard, err
			}
			errSenders = append(errSenders, *sender)
		}

		errMulti, err := send.NewMultiSender(DefaultLogName, send.LevelInfo{Default: level.Error, Threshold: level.Trace}, errSenders)
		if err != nil {
			return ioutil.Discard, err
		}
		o.errorSender = send.NewWriterSender(errMulti)
	}

	if !o.errorIsNull() && o.errorLogging() {
		o.errorMulti = io.MultiWriter(o.Error, o.errorSender)
	} else if !o.errorIsNull() {
		o.errorMulti = o.Error
	} else {
		o.errorMulti = o.errorSender
	}

	return o.errorMulti, nil
}
