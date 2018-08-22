package internal

import (
	"syscall"
	"time"

	"github.com/mongodb/jasper"
)

func (opts *CreateOptions) Export() *jasper.CreateOptions {
	out := &jasper.CreateOptions{
		Args:             opts.Args,
		Environment:      opts.Environment,
		WorkingDirectory: opts.WorkingDirectory,
		Timeout:          time.Duration(opts.TimeoutSeconds) * time.Second,
		TimeoutSecs:      int(opts.TimeoutSeconds),
		OverrideEnviron:  opts.OverrideEnviron,
		Tags:             opts.Tags,
	}

	for _, opt := range opts.OnSuccess {
		out.OnSuccess = append(out.OnSuccess, opt.Export())
	}

	for _, opt := range opts.OnFailure {
		out.OnFailure = append(out.OnFailure, opt.Export())
	}
	for _, opt := range opts.OnTimeout {
		out.OnTimeout = append(out.OnTimeout, opt.Export())
	}

	return out
}

func ConvertCreateOptions(opts *jasper.CreateOptions) *CreateOptions {
	if opts.TimeoutSecs == 0 && opts.Timeout != 0 {
		opts.TimeoutSecs = int(opts.Timeout.Seconds())
	}

	co := &CreateOptions{
		Args:             opts.Args,
		Environment:      opts.Environment,
		WorkingDirectory: opts.WorkingDirectory,
		TimeoutSeconds:   int64(opts.TimeoutSecs),
		OverrideEnviron:  opts.OverrideEnviron,
		Tags:             opts.Tags,
	}

	for _, opt := range opts.OnSuccess {
		co.OnSuccess = append(co.OnSuccess, ConvertCreateOptions(opt))
	}

	for _, opt := range opts.OnFailure {
		co.OnFailure = append(co.OnFailure, ConvertCreateOptions(opt))
	}
	for _, opt := range opts.OnTimeout {
		co.OnTimeout = append(co.OnTimeout, ConvertCreateOptions(opt))
	}

	return co
}

func (info *ProcessInfo) Export() jasper.ProcessInfo {
	return jasper.ProcessInfo{
		ID:         info.Id,
		PID:        int(info.Pid),
		IsRunning:  info.Running,
		Successful: info.Successful,
		Complete:   info.Complete,
		Timeout:    info.Timedout,
		Options:    *info.Options.Export(),
	}
}

func ConvertProcessInfo(info jasper.ProcessInfo) *ProcessInfo {
	return &ProcessInfo{
		Id:         info.ID,
		Pid:        int64(info.PID),
		Running:    info.IsRunning,
		Successful: info.Successful,
		Complete:   info.Complete,
		Timedout:   info.Timeout,
		Options:    ConvertCreateOptions(&info.Options),
	}
}

func (s Signals) Export() syscall.Signal {
	switch s {
	case Signals_SIGHUP:
		return syscall.SIGHUP
	case Signals_SIGINT:
		return syscall.SIGINT
	case Signals_SIGTERM:
		return syscall.SIGTERM
	case Signals_SIGKILL:
		return syscall.SIGKILL
	case Signals_SIGUSR1:
		return syscall.SIGUSR1
	case Signals_SIGUSR2:
		return syscall.SIGUSR2
	default:
		return syscall.Signal(0)
	}
}

func ConvertSignal(s syscall.Signal) Signals {
	switch s {
	case syscall.SIGHUP:
		return Signals_SIGHUP
	case syscall.SIGINT:
		return Signals_SIGINT
	case syscall.SIGTERM:
		return Signals_SIGTERM
	case syscall.SIGKILL:
		return Signals_SIGKILL
	case syscall.SIGUSR1:
		return Signals_SIGUSR1
	case syscall.SIGUSR2:
		return Signals_SIGUSR2
	default:
		return Signals_UNKNOWN
	}
}

func ConvertFilter(f jasper.Filter) *Filter {
	switch f {
	case jasper.All:
		return &Filter{Name: FilterSpecifications_ALL}
	case jasper.Running:
		return &Filter{Name: FilterSpecifications_RUNNING}
	case jasper.Terminated:
		return &Filter{Name: FilterSpecifications_TERMINATED}
	case jasper.Failed:
		return &Filter{Name: FilterSpecifications_FAILED}
	case jasper.Successful:
		return &Filter{Name: FilterSpecifications_SUCCESSFUL}
	default:
		return nil
	}
}
