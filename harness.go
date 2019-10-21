// +build none

package jasper

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/evergreen-ci/poplar"
	"github.com/mongodb/grip/send"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

type makeProcess func(context.Context, *options.Create) (Process, error)

func yesCreateOpts(timeout time.Duration) options.Create {
	return options.Create{Args: []string{"yes"}, Timeout: timeout}
}

func procMap() map[string]func(context.Context, *options.Create) (Process, error) {
	return map[string]func(context.Context, *options.Create) (Process, error){
		"Basic":    newBasicProcess,
		"Blocking": newBlockingProcess,
	}
}

func runIteration(ctx context.Context, makeProc makeProcess, opts *options.Create) error {
	proc, err := makeProc(ctx, opts)
	if err != nil {
		return err
	}
	exitCode, err := proc.Wait(ctx)
	if err != nil && !proc.Info(ctx).Timeout {
		return errors.Wrapf(err, "process with id '%s' exited unexpectedly with code %d", proc.ID(), exitCode)
	}
	return nil
}

func makeCreateOpts(timeout time.Duration, logger options.Logger) *options.Create {
	opts := yesCreateOpts(timeout)
	opts.Output.Loggers = []options.Logger{logger}
	return &opts
}

func getInMemoryBenchmark(makeProc makeProcess) poplar.Benchmark {
	var logType options.LogType = options.LogInMemory
	logOptions := options.Log{InMemoryCap: 1000, Format: options.LogFormatPlain}
	opts := makeCreateOpts(c.timeout, options.Logger{Type: logType, Options: logOptions})

	return func(ctx context.Context, r poplar.Recorder, _ int) error {
		startAt := time.Now()
		r.Begin()
		err := runIteration(ctx, makeProc, opts)
		if err != nil {
			return err
		}
		r.Inc(1)
		logger := opts.Output.Loggers[0].GetSender().(*send.InMemorySender)
		r.IncSize(logger.TotalBytesSent())
		r.End(time.Since(startAt))

		return nil
	}
}

func getFileLoggerBenchmark(makeProc makeProcess) poplar.Benchmark {
	return func(ctx context.Context, r poplar.Recorder, _ int) error {
		var logType options.LogType = options.LogFile
		file, err := ioutil.TempFile("build", "bench_out.txt")
		if err != nil {
			return err
		}
		defer os.Remove(file.Name())
		logOptions := options.Log{FileName: file.Name(), Format: options.LogFormatPlain}
		opts := makeCreateOpts(c.timeout, options.Logger{Type: logType, Options: logOptions})

		startAt := time.Now()
		r.Begin()
		err = runIteration(ctx, makeProc, opts)
		if err != nil {
			return err
		}
		r.Inc(1)
		info, err := file.Stat()
		if err != nil {
			return err
		}
		r.IncSize(info.Size())
		r.End(time.Since(startAt))

		return nil
	}
}

func logBenchmarks() map[string]func(context.Context, *caseDefinition) result {
	return map[string]func(makeProcess) poplar.Benchmark{
		"InMemoryLogger": getInMemoryLoggerBenchmark,
		"FileLogger":     getFileLoggerBenchmark,
	}
}

func getLogBenchmarkSuite() poplar.BenchmarkSuite {
	benchmarkSuite := poplar.BenchmarkSuite{}
	for procName, makeProc := range procMap() {
		for logName, logBench := range logBenchmarks() {
			cases = append(benchmarkSuite,
				// TODO: figure out reasonable setttings for these cases
				&poplar.BenchmarkCase{
					Name:  fmt.Sprintf("%s/%s/Send1Second", logName, procName),
					Bench: logBench(makeProc),
				},
				&poplar.BenchmarkCase{
					Name:  fmt.Sprintf("%s/%s/Send5Seconds", logName, procName),
					Bench: logBench(markProc),
				},
				&poplar.BenchmarkCase{
					Name:  fmt.Sprintf("%s/%s/Send30Seconds", logName, procName),
					Bench: logBench(markProc),
				},
			)
		}
	}

	return benchmarkSuite
}
