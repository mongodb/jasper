package internal

import (
	"bytes"
	"encoding/json"
	"os"
	"syscall"
	"time"

	"github.com/evergreen-ci/bond"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/grip/send"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Export takes a protobuf RPC CreateOptions struct and returns the analogous
// Jasper CreateOptions struct. It is not safe to concurrently access the
// exported RPC CreateOptions and the returned Jasper CreateOptions.
func (opts *CreateOptions) Export() (*options.Create, error) {
	out := &options.Create{
		Args:               opts.Args,
		Environment:        opts.Environment,
		WorkingDirectory:   opts.WorkingDirectory,
		Timeout:            time.Duration(opts.TimeoutSeconds) * time.Second,
		TimeoutSecs:        int(opts.TimeoutSeconds),
		OverrideEnviron:    opts.OverrideEnviron,
		Tags:               opts.Tags,
		StandardInputBytes: opts.StandardInputBytes,
	}
	if len(opts.StandardInputBytes) != 0 {
		out.StandardInput = bytes.NewBuffer(opts.StandardInputBytes)
	}

	if opts.Output != nil {
		exportedOutput, err := opts.Output.Export()
		if err != nil {
			return nil, errors.Wrap(err, "exporting output")
		}
		out.Output = exportedOutput
	}

	for _, opt := range opts.OnSuccess {
		exportedOpt, err := opt.Export()
		if err != nil {
			return nil, errors.Wrap(err, "exporting create options")
		}
		out.OnSuccess = append(out.OnSuccess, exportedOpt)
	}
	for _, opt := range opts.OnFailure {
		exportedOpt, err := opt.Export()
		if err != nil {
			return nil, errors.Wrap(err, "exporting create options")
		}
		out.OnFailure = append(out.OnFailure, exportedOpt)
	}
	for _, opt := range opts.OnTimeout {
		exportedOpt, err := opt.Export()
		if err != nil {
			return nil, errors.Wrap(err, "exporting create options")
		}
		out.OnTimeout = append(out.OnTimeout, exportedOpt)
	}

	return out, nil
}

// ConvertCreateOptions takes a Jasper CreateOptions struct and returns an
// equivalent protobuf RPC *CreateOptions struct. ConvertCreateOptions is the
// inverse of (*CreateOptions) Export(). It is not safe to concurrently
// access the converted Jasper CreateOptions and the returned RPC
// CreateOptions.
func ConvertCreateOptions(opts *options.Create) (*CreateOptions, error) {
	if opts.TimeoutSecs == 0 && opts.Timeout != 0 {
		opts.TimeoutSecs = int(opts.Timeout.Seconds())
	}

	output, err := ConvertOutputOptions(opts.Output)
	if err != nil {
		return nil, errors.Wrap(err, "converting output options")
	}

	co := &CreateOptions{
		Args:               opts.Args,
		Environment:        opts.Environment,
		WorkingDirectory:   opts.WorkingDirectory,
		TimeoutSeconds:     int64(opts.TimeoutSecs),
		OverrideEnviron:    opts.OverrideEnviron,
		Tags:               opts.Tags,
		Output:             &output,
		StandardInputBytes: opts.StandardInputBytes,
	}

	for _, opt := range opts.OnSuccess {
		convertedOpts, err := ConvertCreateOptions(opt)
		if err != nil {
			return nil, errors.Wrap(err, "converting on success create options")
		}
		co.OnSuccess = append(co.OnSuccess, convertedOpts)
	}
	for _, opt := range opts.OnFailure {
		convertedOpts, err := ConvertCreateOptions(opt)
		if err != nil {
			return nil, errors.Wrap(err, "converting on failure create options")
		}
		co.OnFailure = append(co.OnFailure, convertedOpts)
	}
	for _, opt := range opts.OnTimeout {
		convertedOpts, err := ConvertCreateOptions(opt)
		if err != nil {
			return nil, errors.Wrap(err, "converting on timeout create options")
		}
		co.OnTimeout = append(co.OnTimeout, convertedOpts)
	}

	return co, nil
}

// Export takes a protobuf RPC ProcessInfo struct and returns the analogous
// Jasper ProcessInfo struct.
func (info *ProcessInfo) Export() (jasper.ProcessInfo, error) {
	opts, err := info.Options.Export()
	if err != nil {
		return jasper.ProcessInfo{}, errors.Wrap(err, "exporting create options")
	}
	return jasper.ProcessInfo{
		ID:         info.Id,
		PID:        int(info.Pid),
		IsRunning:  info.Running,
		Successful: info.Successful,
		Complete:   info.Complete,
		ExitCode:   int(info.ExitCode),
		Timeout:    info.Timedout,
		Options:    *opts,
		StartAt:    info.StartAt.AsTime(),
		EndAt:      info.EndAt.AsTime(),
	}, nil
}

// ConvertProcessInfo takes a Jasper ProcessInfo struct and returns an
// equivalent protobuf RPC *ProcessInfo struct. ConvertProcessInfo is the
// inverse of (*ProcessInfo) Export().
func ConvertProcessInfo(info jasper.ProcessInfo) (*ProcessInfo, error) {
	opts, err := ConvertCreateOptions(&info.Options)
	if err != nil {
		return nil, errors.Wrap(err, "converting create options")
	}
	return &ProcessInfo{
		Id:         info.ID,
		Pid:        int64(info.PID),
		ExitCode:   int32(info.ExitCode),
		Running:    info.IsRunning,
		Successful: info.Successful,
		Complete:   info.Complete,
		Timedout:   info.Timeout,
		StartAt:    timestamppb.New(info.StartAt),
		EndAt:      timestamppb.New(info.EndAt),
		Options:    opts,
	}, nil
}

// Export takes a protobuf RPC Signals struct and returns the analogous
// syscall.Signal.
func (s Signals) Export() syscall.Signal {
	switch s {
	case Signals_HANGUP:
		return syscall.SIGHUP
	case Signals_INIT:
		return syscall.SIGINT
	case Signals_TERMINATE:
		return syscall.SIGTERM
	case Signals_KILL:
		return syscall.SIGKILL
	case Signals_ABRT:
		return syscall.SIGABRT
	default:
		return syscall.Signal(0)
	}
}

// ConvertSignal takes a syscall.Signal and returns an
// equivalent protobuf RPC Signals struct. ConvertSignals is the
// inverse of (Signals) Export().
func ConvertSignal(s syscall.Signal) Signals {
	switch s {
	case syscall.SIGHUP:
		return Signals_HANGUP
	case syscall.SIGINT:
		return Signals_INIT
	case syscall.SIGTERM:
		return Signals_TERMINATE
	case syscall.SIGKILL:
		return Signals_KILL
	default:
		return Signals_UNKNOWN
	}
}

// ConvertFilter takes a Jasper Filter struct and returns an
// equivalent protobuf RPC *Filter struct.
func ConvertFilter(f options.Filter) *Filter {
	switch f {
	case options.All:
		return &Filter{Name: FilterSpecifications_ALL}
	case options.Running:
		return &Filter{Name: FilterSpecifications_RUNNING}
	case options.Terminated:
		return &Filter{Name: FilterSpecifications_TERMINATED}
	case options.Failed:
		return &Filter{Name: FilterSpecifications_FAILED}
	case options.Successful:
		return &Filter{Name: FilterSpecifications_SUCCESSFUL}
	default:
		return nil
	}
}

// Export takes a protobuf RPC OutputOptions struct and returns the analogous
// Jasper OutputOptions struct.
func (opts *OutputOptions) Export() (options.Output, error) {
	loggers := []*options.LoggerConfig{}
	for _, logger := range opts.Loggers {
		exportedLogger, err := logger.Export()
		if err != nil {
			return options.Output{}, errors.Wrap(err, "exporting logger config")
		}
		loggers = append(loggers, exportedLogger)
	}
	return options.Output{
		SuppressOutput:    opts.SuppressOutput,
		SuppressError:     opts.SuppressError,
		SendOutputToError: opts.RedirectOutputToError,
		SendErrorToOutput: opts.RedirectErrorToOutput,
		Loggers:           loggers,
	}, nil
}

// ConvertOutputOptions takes a Jasper OutputOptions struct and returns an
// equivalent protobuf RPC OutputOptions struct. ConvertOutputOptions is the
// inverse of (OutputOptions) Export().
func ConvertOutputOptions(opts options.Output) (OutputOptions, error) {
	loggers := []*LoggerConfig{}
	for _, logger := range opts.Loggers {
		convertedLoggerConfig, err := ConvertLoggerConfig(logger)
		if err != nil {
			return OutputOptions{}, errors.Wrap(err, "converting logger config")
		}
		loggers = append(loggers, convertedLoggerConfig)
	}
	return OutputOptions{
		SuppressOutput:        opts.SuppressOutput,
		SuppressError:         opts.SuppressError,
		RedirectOutputToError: opts.SendOutputToError,
		RedirectErrorToOutput: opts.SendErrorToOutput,
		Loggers:               loggers,
	}, nil
}

// Export takes a protobuf RPC Logger struct and returns the analogous
// Jasper Logger struct.
func (logger *LoggerConfig) Export() (*options.LoggerConfig, error) {
	var producer options.LoggerProducer
	switch {
	case logger.GetDefault() != nil:
		producer = logger.GetDefault().Export()
	case logger.GetFile() != nil:
		producer = logger.GetFile().Export()
	case logger.GetInherited() != nil:
		producer = logger.GetInherited().Export()
	case logger.GetInMemory() != nil:
		producer = logger.GetInMemory().Export()
	case logger.GetSplunk() != nil:
		producer = logger.GetSplunk().Export()
	case logger.GetBuildloggerv2() != nil:
		producer = logger.GetBuildloggerv2().Export()
	case logger.GetBuildloggerv3() != nil:
		return logger.GetBuildloggerv3().Export()
	case logger.GetRaw() != nil:
		return logger.GetRaw().Export()
	}
	if producer == nil {
		return nil, errors.New("invalid logger config options")
	}

	config := &options.LoggerConfig{}
	return config, config.Set(producer)
}

// ConvertLoggerConfig takes a Jasper options.LoggerConfig struct and returns
// an equivalent protobuf RPC LoggerConfig struct. ConvertLoggerConfig is the
// inverse of (LoggerConfig) Export().
func ConvertLoggerConfig(config *options.LoggerConfig) (*LoggerConfig, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling logger config to JSON")
	}

	return &LoggerConfig{
		Producer: &LoggerConfig_Raw{
			Raw: &RawLoggerConfig{
				Format:     ConvertRawLoggerConfigFormat(options.RawLoggerConfigFormatJSON),
				ConfigData: data,
			},
		},
	}, nil
}

// Export takes a protobuf RPC LogLevel struct and returns the analogous send
// LevelInfo struct.
func (l *LogLevel) Export() send.LevelInfo {
	return send.LevelInfo{Threshold: level.Priority(l.Threshold), Default: level.Priority(l.Default)}
}

// ConvertLogLevel takes a send LevelInfo struct and returns an equivalent
// protobuf RPC LogLevel struct. ConvertLogLevel is the inverse of
// (*LogLevel) Export().
func ConvertLogLevel(l send.LevelInfo) *LogLevel {
	return &LogLevel{Threshold: int32(l.Threshold), Default: int32(l.Default)}
}

// Export takes a protobuf RPC BufferOptions struct and returns the analogous
// Jasper BufferOptions struct.
func (opts *BufferOptions) Export() options.BufferOptions {
	return options.BufferOptions{
		Buffered: opts.Buffered,
		Duration: time.Duration(opts.Duration),
		MaxSize:  int(opts.MaxSize),
	}
}

// ConvertBufferOptions takes a Jasper BufferOptions struct and returns an
// equivalent protobuf RPC BufferOptions struct. ConvertBufferOptions is the
// inverse of (*BufferOptions) Export().
func ConvertBufferOptions(opts options.BufferOptions) *BufferOptions {
	return &BufferOptions{
		Buffered: opts.Buffered,
		Duration: int64(opts.Duration),
		MaxSize:  int64(opts.MaxSize),
	}
}

// Export takes a protobuf RPC LogFormat struct and returns the analogous
// Jasper LogFormat struct.
func (f LogFormat) Export() options.LogFormat {
	switch f {
	case LogFormat_LOGFORMATDEFAULT:
		return options.LogFormatDefault
	case LogFormat_LOGFORMATJSON:
		return options.LogFormatJSON
	case LogFormat_LOGFORMATPLAIN:
		return options.LogFormatPlain
	default:
		return options.LogFormatInvalid
	}
}

// ConvertLogFormat takes a Jasper LogFormat struct and returns an
// equivalent protobuf RPC LogFormat struct. ConvertLogFormat is the
// inverse of (LogFormat) Export().
func ConvertLogFormat(f options.LogFormat) LogFormat {
	switch f {
	case options.LogFormatDefault:
		return LogFormat_LOGFORMATDEFAULT
	case options.LogFormatJSON:
		return LogFormat_LOGFORMATJSON
	case options.LogFormatPlain:
		return LogFormat_LOGFORMATPLAIN
	default:
		return LogFormat_LOGFORMATUNKNOWN
	}
}

// Export takes a protobuf RPC BaseOptions struct and returns the analogous
// Jasper BaseOptions struct.
func (opts *BaseOptions) Export() options.BaseOptions {
	return options.BaseOptions{
		Level:  opts.Level.Export(),
		Buffer: opts.Buffer.Export(),
		Format: opts.Format.Export(),
	}
}

// Export takes a protobuf RPC DefaultLoggerOptions struct and returns the
// analogous Jasper options.LoggerProducer.
func (opts *DefaultLoggerOptions) Export() options.LoggerProducer {
	return &options.DefaultLoggerOptions{
		Prefix: opts.Prefix,
		Base:   opts.Base.Export(),
	}
}

// Export takes a protobuf RPC FileLoggerOptions struct and returns the
// analogous Jasper options.LoggerProducer.
func (opts *FileLoggerOptions) Export() options.LoggerProducer {
	return &options.FileLoggerOptions{
		Filename: opts.Filename,
		Base:     opts.Base.Export(),
	}
}

// Export takes a protobuf RPC InheritedLoggerOptions struct and returns the
// analogous Jasper options.LoggerProducer.
func (opts *InheritedLoggerOptions) Export() options.LoggerProducer {
	return &options.InheritedLoggerOptions{
		Base: opts.Base.Export(),
	}
}

// Export takes a protobuf RPC InMemoryLoggerOptions struct and returns the
// analogous Jasper options.LoggerProducer.
func (opts *InMemoryLoggerOptions) Export() options.LoggerProducer {
	return &options.InMemoryLoggerOptions{
		InMemoryCap: int(opts.InMemoryCap),
		Base:        opts.Base.Export(),
	}
}

// Export takes a protobuf RPC SplunkInfo struct and returns the analogous
// grip send.SplunkConnectionInfo struct.
func (opts *SplunkInfo) Export() send.SplunkConnectionInfo {
	return send.SplunkConnectionInfo{
		ServerURL: opts.Url,
		Token:     opts.Token,
		Channel:   opts.Channel,
	}
}

// ConvertSplunkInfo takes a grip send.SplunkConnectionInfo and returns the
// analogous protobuf RPC SplunkInfo struct. ConvertSplunkInfo is the inverse
// of (SplunkInfo) Export().
func ConvertSplunkInfo(opts send.SplunkConnectionInfo) *SplunkInfo {
	return &SplunkInfo{
		Url:     opts.ServerURL,
		Token:   opts.Token,
		Channel: opts.Channel,
	}
}

// Export takes a protobuf RPC SplunkLoggerOptions struct and returns the
// analogous Jasper options.LoggerProducer.
func (opts *SplunkLoggerOptions) Export() options.LoggerProducer {
	return &options.SplunkLoggerOptions{
		Splunk: opts.Splunk.Export(),
		Base:   opts.Base.Export(),
	}
}

// Export takes a protobuf RPC BuildloggerV2Info struct and returns the
// analogous grip send.BuildloggerConfig struct.
func (opts *BuildloggerV2Info) Export() send.BuildloggerConfig {
	return send.BuildloggerConfig{
		CreateTest: opts.CreateTest,
		URL:        opts.Url,
		Number:     int(opts.Number),
		Phase:      opts.Phase,
		Builder:    opts.Builder,
		Test:       opts.Test,
		Command:    opts.Command,
	}
}

// ConvertBuildloggerOptions takes a grip send.BuildloggerConfig and returns an
// equivalent protobuf RPC BuildloggerV2Info struct. ConvertBuildloggerOptions
// is the inverse of (BuildloggerV2Info) Export().
func ConvertBuildloggerOptions(opts send.BuildloggerConfig) *BuildloggerV2Info {
	return &BuildloggerV2Info{
		CreateTest: opts.CreateTest,
		Url:        opts.URL,
		Number:     int64(opts.Number),
		Phase:      opts.Phase,
		Builder:    opts.Builder,
		Test:       opts.Test,
		Command:    opts.Command,
	}
}

// Export takes a protobuf RPC BuildloggerV2Options struct and returns the
// analogous Jasper options.LoggerProducer.
func (opts *BuildloggerV2Options) Export() options.LoggerProducer {
	return &options.BuildloggerV2Options{
		Buildlogger: opts.Buildlogger.Export(),
		Base:        opts.Base.Export(),
	}
}

// Export takes the protobuf RPC BuildloggerV3Options struct and returns the
// analogous Jasper options.LoggerConfig.
func (opts *BuildloggerV3Options) Export() (*options.LoggerConfig, error) {
	data, err := json.Marshal(&opts)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling buildlogger v3 options to JSON")
	}

	return options.NewLoggerConfig("BuildloggerV3", options.RawLoggerConfigFormatJSON, data), nil
}

// Export takes a protobuf RPC RawLoggerConfigFormat enum and returns the
// analogous Jasper options.RawLoggerConfigFormat type.
func (f RawLoggerConfigFormat) Export() options.RawLoggerConfigFormat {
	switch f {
	case RawLoggerConfigFormat_RAWLOGGERCONFIGFORMATJSON:
		return options.RawLoggerConfigFormatJSON
	case RawLoggerConfigFormat_RAWLOGGERCONFIGFORMATBSON:
		return options.RawLoggerConfigFormatBSON
	default:
		return options.RawLoggerConfigFormatInvalid
	}
}

// ConvertRawLoggerConfigFormat takes a Jasper RawLoggerConfigFormat type and
// returns an equivalent protobuf RPC RawLoggerConfigFormat enum.
// ConvertLogFormat is the inverse of (RawLoggerConfigFormat) Export().
func ConvertRawLoggerConfigFormat(f options.RawLoggerConfigFormat) RawLoggerConfigFormat {
	switch f {
	case options.RawLoggerConfigFormatJSON:
		return RawLoggerConfigFormat_RAWLOGGERCONFIGFORMATJSON
	case options.RawLoggerConfigFormatBSON:
		return RawLoggerConfigFormat_RAWLOGGERCONFIGFORMATBSON
	default:
		return RawLoggerConfigFormat_RAWLOGGERCONFIGFORMATUNKNOWN
	}
}

// Export takes a protobuf RPC RawLoggerConfig struct and returns the
// analogous Jasper options.LoggerConfig
func (logger *RawLoggerConfig) Export() (*options.LoggerConfig, error) {
	config := &options.LoggerConfig{}
	if err := logger.Format.Export().Unmarshal(logger.ConfigData, config); err != nil {
		return nil, errors.Wrap(err, "unmarshalling raw config")
	}
	return config, nil
}

// Export takes a protobuf RPC BuildloggerURLs struct and returns the analogous
// []string.
func (u *BuildloggerURLs) Export() []string {
	return append([]string{}, u.Urls...)
}

// ConvertBuildloggerURLs takes a []string and returns the analogous protobuf
// RPC BuildloggerURLs struct. ConvertBuildloggerURLs is the
// inverse of (*BuildloggerURLs) Export().
func ConvertBuildloggerURLs(urls []string) *BuildloggerURLs {
	u := &BuildloggerURLs{Urls: []string{}}
	u.Urls = append(u.Urls, urls...)
	return u
}

// Export takes a protobuf RPC BuildOptions struct and returns the analogous
// bond.BuildOptions struct.
func (opts *BuildOptions) Export() bond.BuildOptions {
	return bond.BuildOptions{
		Target:  opts.Target,
		Arch:    bond.MongoDBArch(opts.Arch),
		Edition: bond.MongoDBEdition(opts.Edition),
		Debug:   opts.Debug,
	}
}

// ConvertBuildOptions takes a bond BuildOptions struct and returns an
// equivalent protobuf RPC BuildOptions struct. ConvertBuildOptions is the
// inverse of (*BuildOptions) Export().
func ConvertBuildOptions(opts bond.BuildOptions) *BuildOptions {
	return &BuildOptions{
		Target:  opts.Target,
		Arch:    string(opts.Arch),
		Edition: string(opts.Edition),
		Debug:   opts.Debug,
	}
}

// Export takes a protobuf RPC MongoDBDownloadOptions struct and returns the
// analogous Jasper MongoDBDownloadOptions struct.
func (opts *MongoDBDownloadOptions) Export() options.MongoDBDownload {
	jopts := options.MongoDBDownload{
		BuildOpts: opts.BuildOpts.Export(),
		Path:      opts.Path,
		Releases:  make([]string, 0, len(opts.Releases)),
	}
	jopts.Releases = append(jopts.Releases, opts.Releases...)
	return jopts
}

// ConvertMongoDBDownloadOptions takes a Jasper MongoDBDownloadOptions struct
// and returns an equivalent protobuf RPC MongoDBDownloadOptions struct.
// ConvertMongoDBDownloadOptions is the
// inverse of (*MongoDBDownloadOptions) Export().
func ConvertMongoDBDownloadOptions(jopts options.MongoDBDownload) *MongoDBDownloadOptions {
	opts := &MongoDBDownloadOptions{
		BuildOpts: ConvertBuildOptions(jopts.BuildOpts),
		Path:      jopts.Path,
		Releases:  make([]string, 0, len(jopts.Releases)),
	}
	opts.Releases = append(opts.Releases, jopts.Releases...)
	return opts
}

// Export takes a protobuf RPC CacheOptions struct and returns the analogous
// Jasper CacheOptions struct.
func (opts *CacheOptions) Export() options.Cache {
	return options.Cache{
		Disabled:   opts.Disabled,
		PruneDelay: time.Duration(opts.PruneDelaySeconds) * time.Second,
		MaxSize:    int(opts.MaxSize),
	}
}

// ConvertCacheOptions takes a Jasper CacheOptions struct and returns an
// equivalent protobuf RPC CacheOptions struct. ConvertCacheOptions is the
// inverse of (*CacheOptions) Export().
func ConvertCacheOptions(jopts options.Cache) *CacheOptions {
	return &CacheOptions{
		Disabled:          jopts.Disabled,
		PruneDelaySeconds: int64(jopts.PruneDelay / time.Second),
		MaxSize:           int64(jopts.MaxSize),
	}
}

// Export takes a protobuf RPC DownloadInfo struct and returns the analogous
// options.Download struct.
func (opts *DownloadInfo) Export() options.Download {
	return options.Download{
		Path:        opts.Path,
		URL:         opts.Url,
		ArchiveOpts: opts.ArchiveOpts.Export(),
	}
}

// ConvertDownloadOptions takes an options.Download struct and returns an
// equivalent protobuf RPC DownloadInfo struct. ConvertDownloadOptions is the
// inverse of (*DownloadInfo) Export().
func ConvertDownloadOptions(opts options.Download) *DownloadInfo {
	return &DownloadInfo{
		Path:        opts.Path,
		Url:         opts.URL,
		ArchiveOpts: ConvertArchiveOptions(opts.ArchiveOpts),
	}
}

// Export takes a protobuf RPC WriteFileInfo struct and returns the analogous
// options.WriteFile struct.
func (opts *WriteFileInfo) Export() options.WriteFile {
	return options.WriteFile{
		Path:    opts.Path,
		Content: opts.Content,
		Append:  opts.Append,
		Perm:    os.FileMode(opts.Perm),
	}
}

// ConvertWriteFileOptions takes an options.WriteFile struct and returns an
// equivalent protobuf RPC WriteFileInfo struct. ConvertWriteFileOptions is the
// inverse of (*WriteFileInfo) Export().
func ConvertWriteFileOptions(opts options.WriteFile) *WriteFileInfo {
	return &WriteFileInfo{
		Path:    opts.Path,
		Content: opts.Content,
		Append:  opts.Append,
		Perm:    uint32(opts.Perm),
	}
}

// Export takes a protobuf RPC ArchiveFormat struct and returns the analogous
// Jasper ArchiveFormat struct.
func (format ArchiveFormat) Export() options.ArchiveFormat {
	switch format {
	case ArchiveFormat_ARCHIVEAUTO:
		return options.ArchiveAuto
	case ArchiveFormat_ARCHIVETARGZ:
		return options.ArchiveTarGz
	case ArchiveFormat_ARCHIVEZIP:
		return options.ArchiveZip
	default:
		return options.ArchiveFormat("")
	}
}

// ConvertArchiveFormat takes a Jasper ArchiveFormat struct and returns an
// equivalent protobuf RPC ArchiveFormat struct. ConvertArchiveFormat is the
// inverse of (ArchiveFormat) Export().
func ConvertArchiveFormat(format options.ArchiveFormat) ArchiveFormat {
	switch format {
	case options.ArchiveAuto:
		return ArchiveFormat_ARCHIVEAUTO
	case options.ArchiveTarGz:
		return ArchiveFormat_ARCHIVETARGZ
	case options.ArchiveZip:
		return ArchiveFormat_ARCHIVEZIP
	default:
		return ArchiveFormat_ARCHIVEUNKNOWN
	}
}

// Export takes a protobuf RPC ArchiveOptions struct and returns the analogous
// Jasper ArchiveOptions struct.
func (opts *ArchiveOptions) Export() options.Archive {
	return options.Archive{
		ShouldExtract: opts.ShouldExtract,
		Format:        opts.Format.Export(),
		TargetPath:    opts.TargetPath,
	}
}

// ConvertArchiveOptions takes a Jasper ArchiveOptions struct and returns an
// equivalent protobuf RPC ArchiveOptions struct. ConvertArchiveOptions is the
// inverse of (ArchiveOptions) Export().
func ConvertArchiveOptions(opts options.Archive) *ArchiveOptions {
	return &ArchiveOptions{
		ShouldExtract: opts.ShouldExtract,
		Format:        ConvertArchiveFormat(opts.Format),
		TargetPath:    opts.TargetPath,
	}
}

// Export takes a protobuf RPC SignalTriggerParams struct and returns the analogous
// Jasper process ID and SignalTriggerID.
func (t *SignalTriggerParams) Export() (string, jasper.SignalTriggerID) {
	return t.ProcessID.Value, t.SignalTriggerID.Export()
}

// ConvertSignalTriggerParams takes a Jasper process ID and a SignalTriggerID
// and returns an equivalent protobuf RPC SignalTriggerParams struct.
// ConvertSignalTriggerParams is the inverse of (SignalTriggerParams) Export().
func ConvertSignalTriggerParams(jasperProcessID string, signalTriggerID jasper.SignalTriggerID) *SignalTriggerParams {
	return &SignalTriggerParams{
		ProcessID:       &JasperProcessID{Value: jasperProcessID},
		SignalTriggerID: ConvertSignalTriggerID(signalTriggerID),
	}
}

// Export takes a protobuf RPC SignalTriggerID and returns the analogous
// Jasper SignalTriggerID.
func (t SignalTriggerID) Export() jasper.SignalTriggerID {
	switch t {
	case SignalTriggerID_CLEANTERMINATION:
		return jasper.CleanTerminationSignalTrigger
	default:
		return jasper.SignalTriggerID("")
	}
}

// ConvertSignalTriggerID takes a Jasper SignalTriggerID and returns an
// equivalent protobuf RPC SignalTriggerID. ConvertSignalTrigger is the
// inverse of (SignalTriggerID) Export().
func ConvertSignalTriggerID(id jasper.SignalTriggerID) SignalTriggerID {
	switch id {
	case jasper.CleanTerminationSignalTrigger:
		return SignalTriggerID_CLEANTERMINATION
	default:
		return SignalTriggerID_NONE
	}
}

// Export takes a protobuf RPC LogStream and returns the analogous
// Jasper LogStream.
func (l *LogStream) Export() jasper.LogStream {
	return jasper.LogStream{
		Logs: l.Logs,
		Done: l.Done,
	}
}

// ConvertLogStream takes a Jasper LogStream and returns an
// equivalent protobuf RPC LogStream. ConvertLogStream is the
// inverse of (*LogStream) Export().
func ConvertLogStream(l jasper.LogStream) *LogStream {
	return &LogStream{
		Logs: l.Logs,
		Done: l.Done,
	}
}

// Export takes a protobuf RPC LoggingPayloadFormat and returns the
// analogous LoggingPayloadFormat.
func (lf LoggingPayloadFormat) Export() options.LoggingPayloadFormat {
	switch lf {
	case LoggingPayloadFormat_FORMATBSON:
		return options.LoggingPayloadFormatJSON
	case LoggingPayloadFormat_FORMATJSON:
		return options.LoggingPayloadFormatBSON
	case LoggingPayloadFormat_FORMATSTRING:
		return options.LoggingPayloadFormatString
	default:
		return ""
	}
}

// ConvertLoggingPayloadFormat takes LoggingPayloadFormat options and returns an
// equivalent protobuf RPC LoggingPayloadFormat. ConvertLoggingPayloadFormat is
// the inverse of (LoggingPayloadFormat) Export().
func ConvertLoggingPayloadFormat(in options.LoggingPayloadFormat) LoggingPayloadFormat {
	switch in {
	case options.LoggingPayloadFormatJSON:
		return LoggingPayloadFormat_FORMATJSON
	case options.LoggingPayloadFormatBSON:
		return LoggingPayloadFormat_FORMATBSON
	case options.LoggingPayloadFormatString:
		return LoggingPayloadFormat_FORMATSTRING
	default:
		return 0
	}
}

// Export takes a protobuf RPC LoggingPayload and returns the
// analogous LoggingPayload options.
func (lp *LoggingPayload) Export() *options.LoggingPayload {
	data := make([]interface{}, len(lp.Data))
	for idx := range lp.Data {
		switch val := lp.Data[idx].Data.(type) {
		case *LoggingPayloadData_Msg:
			data[idx] = val.Msg
		case *LoggingPayloadData_Raw:
			data[idx] = val.Raw
		}
	}

	return &options.LoggingPayload{
		Data:              data,
		LoggerID:          lp.LoggerID,
		IsMulti:           lp.IsMulti,
		PreferSendToError: lp.PreferSendToError,
		AddMetadata:       lp.AddMetadata,
		Priority:          level.Priority(lp.Priority),
		Format:            lp.Format.Export(),
	}
}

func convertMessage(format options.LoggingPayloadFormat, m interface{}) *LoggingPayloadData {
	out := &LoggingPayloadData{}

	switch m := m.(type) {
	case message.Composer:
		switch format {
		case options.LoggingPayloadFormatString:
			out.Data = &LoggingPayloadData_Msg{Msg: m.String()}
		case options.LoggingPayloadFormatBSON:
			payload, _ := bson.Marshal(m.Raw())
			out.Data = &LoggingPayloadData_Raw{Raw: payload}
		case options.LoggingPayloadFormatJSON:
			payload, _ := json.Marshal(m.Raw())
			out.Data = &LoggingPayloadData_Raw{Raw: payload}
		default:
			out.Data = &LoggingPayloadData_Raw{}
		}
	case string:
		switch format {
		case options.LoggingPayloadFormatJSON:
			out.Data = &LoggingPayloadData_Raw{Raw: []byte(m)}
		default:
			out.Data = &LoggingPayloadData_Msg{Msg: m}
		}
	case []byte:
		switch format {
		case options.LoggingPayloadFormatString:
			out.Data = &LoggingPayloadData_Msg{Msg: string(m)}
		default:
			out.Data = &LoggingPayloadData_Raw{Raw: m}
		}
	default:
		out.Data = &LoggingPayloadData_Raw{}
	}
	return out
}

// ConvertLoggingPayload takes LoggingPayload options and returns an
// equivalent protobuf RPC LoggingPayload. ConvertLoggingPayload is
// the inverse of (*LoggingPayload) Export().
func ConvertLoggingPayload(in options.LoggingPayload) *LoggingPayload {
	data := []*LoggingPayloadData{}
	switch val := in.Data.(type) {
	case []interface{}:
		for idx := range val {
			data = append(data, convertMessage(in.Format, val[idx]))
		}
	case []string:
		for idx := range val {
			data = append(data, convertMessage(in.Format, val[idx]))
		}
	case [][]byte:
		for idx := range val {
			data = append(data, convertMessage(in.Format, val[idx]))
		}
	case []message.Composer:
		for idx := range val {
			data = append(data, convertMessage(in.Format, val[idx]))
		}
	case *message.GroupComposer:
		msgs := val.Messages()
		for idx := range msgs {
			data = append(data, convertMessage(in.Format, msgs[idx]))
		}
	case string:
		data = append(data, convertMessage(in.Format, val))
	case []byte:
		data = append(data, convertMessage(in.Format, val))
	}

	return &LoggingPayload{
		LoggerID:          in.LoggerID,
		Priority:          int32(in.Priority),
		IsMulti:           in.IsMulti,
		PreferSendToError: in.PreferSendToError,
		AddMetadata:       in.AddMetadata,
		Format:            ConvertLoggingPayloadFormat(in.Format),
		Data:              data,
	}
}

// Export takes a protobuf RPC LoggingCacheInstance and returns the
// analogous CacheLogger options.
func (l *LoggingCacheInstance) Export() (*options.CachedLogger, error) {
	if !l.Outcome.Success {
		return nil, errors.New(l.Outcome.Text)
	}

	return &options.CachedLogger{
		Accessed:  l.Accessed.AsTime(),
		ID:        l.Id,
		ManagerID: l.ManagerID,
	}, nil
}

// ConvertCachedLogger takes CachedLogger options and returns an
// equivalent protobuf RPC LoggingCacheInstance. ConvertLoggingPayload is
// the inverse of (*LoggingCacheInstance) Export().
func ConvertCachedLogger(opts *options.CachedLogger) *LoggingCacheInstance {
	return &LoggingCacheInstance{
		Outcome: &OperationOutcome{
			Success: true,
		},
		Id:        opts.ID,
		ManagerID: opts.ManagerID,
		Accessed:  timestamppb.New(opts.Accessed),
	}
}

// ConvertLoggingCreateArgs takes the given ID and returns an equivalent
// protobuf RPC LoggingCacheCreateArgs.
func ConvertLoggingCreateArgs(id string, opts *options.Output) (*LoggingCacheCreateArgs, error) {
	o, err := ConvertOutputOptions(*opts)
	if err != nil {
		return nil, errors.Wrap(err, "converting output options")
	}
	return &LoggingCacheCreateArgs{
		Id:      id,
		Options: &o,
	}, nil
}
