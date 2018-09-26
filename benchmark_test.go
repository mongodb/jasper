package jasper

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/send"
	"github.com/stretchr/testify/require"
)

type procMaker func(context.Context, *CreateOptions) (Process, error)
type logMaker func() Logger
type procMap map[string]procMaker
type timeMap map[string]time.Duration

func yesCreateOpts(timeout time.Duration) CreateOptions {
	return CreateOptions{Args: []string{"yes"}, Timeout: timeout}
}

func testDurations() timeMap {
	return timeMap{
		"Send1Second":  1 * time.Second,
		"Send2Seconds": 2 * time.Second,
		"Send4Seconds": 4 * time.Second,
		"Send8Seconds": 8 * time.Second,
	}
}

func getProcMap() procMap {
	return procMap{
		"Blocking": newBlockingProcess,
		"Basic":    newBasicProcess,
	}
}

func runBenchmark(ctx context.Context, b *testing.B, makeProc procMaker, optsList []*CreateOptions) {
	for n := 0; n < b.N; n++ {
		proc, err := makeProc(ctx, optsList[n])
		require.NoError(b, err)
		_ = proc.Wait(ctx)
	}
}

func makeAllCreateOpts(iters int, timeout time.Duration, logger Logger) []*CreateOptions {
	optsList := make([]*CreateOptions, iters)
	for n := 0; n < iters; n++ {
		opts := yesCreateOpts(timeout)
		optsList[n] = &opts
		optsList[n].Output.Loggers = []Logger{logger}
	}
	return optsList
}

func BenchmarkFileLogger(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for procName, makeProc := range getProcMap() {
		b.Run(procName, func(b *testing.B) {
			for sendTime, timeout := range testDurations() {
				b.Run(sendTime, func(b *testing.B) {
					file, err := ioutil.TempFile("build", "bench_out.txt")
					require.NoError(b, err)
					defer os.Remove(file.Name())

					var logType LogType = LogFile
					logOptions := LogOptions{FileName: file.Name()}

					optsList := makeAllCreateOpts(b.N, timeout, Logger{Type: logType, Options: logOptions})
					b.ResetTimer()
					runBenchmark(ctx, b, makeProc, optsList)

					info, err := file.Stat()
					require.NoError(b, err)
					numBytes := info.Size()
					seconds := int64(b.N * int(timeout.Seconds()))
					grip.Info(fmt.Sprintf("%s: bytes/sec = %d\n", b.Name(), numBytes/seconds))
				})
			}
		})
	}
}

func BenchmarkInMemoryLogger(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var logType LogType = LogInMemory
	logOptions := LogOptions{InMemoryCap: 1000}

	for procName, makeProc := range getProcMap() {
		b.Run(procName, func(b *testing.B) {
			for sendTime, timeoutMillis := range testDurations() {
				b.Run(sendTime, func(b *testing.B) {
					size := make(chan int64)
					optsList := makeAllCreateOpts(b.N, timeoutMillis, Logger{Type: logType, Options: logOptions})
					for _, opts := range optsList {
						opts.closers = append(opts.closers, func() {
							require.NotNil(b, opts.Output.outputSender.Sender)
							logger, ok := opts.Output.outputSender.Sender.(*send.InMemorySender)
							require.True(b, ok)
							size <- logger.TotalBytesSent()
						})
					}

					b.ResetTimer()
					runBenchmark(ctx, b, makeProc, optsList)

					var numBytes int64
					for i := 0; i < b.N; i++ {
						numBytes += <-size
					}
					seconds := int64(b.N * int(timeoutMillis.Seconds()))
					grip.Info(fmt.Sprintf("%s: bytes/sec = %d\n", b.Name(), numBytes/seconds))
				})
			}
		})
	}
}
