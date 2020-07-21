package cli

import (
	"context"

	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/remote"
	"github.com/urfave/cli"
)

// Constants representing the Jasper RemoteClient interface as CLI commands.
const (
	RemoteCommand             = "remote"
	ConfigureCacheCommand     = "configure-cache"
	DownloadFileCommand       = "download-file"
	DownloadMongoDBCommand    = "download-mongodb"
	GetBuildloggerURLsCommand = "get-buildlogger-urls"
	GetLogStreamCommand       = "get-log-stream"
	SignalEventCommand        = "signal-event"
	WriteFileCommand          = "write-file"
	SendMessagesCommand       = "send-messages"
)

// Remote creates a cli.Command that allows the remote-specific methods in the
// RemoteClient interface except for CloseClient, for which there is no CLI
// equivalent.
func Remote() cli.Command {
	return cli.Command{
		Name: RemoteCommand,
		Subcommands: []cli.Command{
			remoteConfigureCache(),
			remoteDownloadFile(),
			remoteDownloadMongoDB(),
			remoteGetLogStream(),
			remoteGetBuildloggerURLs(),
			remoteSignalEvent(),
			remoteWriteFile(),
		},
	}
}

func remoteConfigureCache() cli.Command {
	return cli.Command{
		Name:   ConfigureCacheCommand,
		Flags:  clientFlags(),
		Before: clientBefore(),
		Action: func(c *cli.Context) error {
			input := options.Cache{}
			return doPassthroughInputOutput(c, &input, func(ctx context.Context, client remote.Manager) interface{} {
				return makeOutcomeResponse(client.ConfigureCache(ctx, input))
			})
		},
	}
}

func remoteWriteFile() cli.Command {
	return cli.Command{
		Name:   WriteFileCommand,
		Flags:  clientFlags(),
		Before: clientBefore(),
		Action: func(c *cli.Context) error {
			input := options.WriteFile{}
			return doPassthroughInputOutput(c, &input, func(ctx context.Context, client remote.Manager) interface{} {
				return makeOutcomeResponse(client.WriteFile(ctx, input))
			})
		},
	}
}

func remoteDownloadFile() cli.Command {
	return cli.Command{
		Name:   DownloadFileCommand,
		Flags:  clientFlags(),
		Before: clientBefore(),
		Action: func(c *cli.Context) error {
			input := options.Download{}
			return doPassthroughInputOutput(c, &input, func(ctx context.Context, client remote.Manager) interface{} {
				return makeOutcomeResponse(client.DownloadFile(ctx, input))
			})
		},
	}
}

func remoteDownloadMongoDB() cli.Command {
	return cli.Command{
		Name:   DownloadMongoDBCommand,
		Flags:  clientFlags(),
		Before: clientBefore(),
		Action: func(c *cli.Context) error {
			input := options.MongoDBDownload{}
			return doPassthroughInputOutput(c, &input, func(ctx context.Context, client remote.Manager) interface{} {
				return makeOutcomeResponse(client.DownloadMongoDB(ctx, input))
			})
		},
	}
}

func remoteGetLogStream() cli.Command {
	return cli.Command{
		Name:   GetLogStreamCommand,
		Flags:  clientFlags(),
		Before: clientBefore(),
		Action: func(c *cli.Context) error {
			input := LogStreamInput{}
			return doPassthroughInputOutput(c, &input, func(ctx context.Context, client remote.Manager) interface{} {
				logs, err := client.GetLogStream(ctx, input.ID, input.Count)
				if err != nil {
					return &LogStreamResponse{OutcomeResponse: *makeOutcomeResponse(err)}
				}
				return &LogStreamResponse{LogStream: logs, OutcomeResponse: *makeOutcomeResponse(nil)}
			})
		},
	}
}

func remoteGetBuildloggerURLs() cli.Command {
	return cli.Command{
		Name:   GetBuildloggerURLsCommand,
		Flags:  clientFlags(),
		Before: clientBefore(),
		Action: func(c *cli.Context) error {
			input := IDInput{}
			return doPassthroughInputOutput(c, &input, func(ctx context.Context, client remote.Manager) interface{} {
				urls, err := client.GetBuildloggerURLs(ctx, input.ID)
				if err != nil {
					return &BuildloggerURLsResponse{OutcomeResponse: *makeOutcomeResponse(err)}
				}
				return &BuildloggerURLsResponse{URLs: urls, OutcomeResponse: *makeOutcomeResponse(nil)}
			})
		},
	}
}

func remoteSignalEvent() cli.Command {
	return cli.Command{
		Name:   SignalEventCommand,
		Flags:  clientFlags(),
		Before: clientBefore(),
		Action: func(c *cli.Context) error {
			input := EventInput{}
			return doPassthroughInputOutput(c, &input, func(ctx context.Context, client remote.Manager) interface{} {
				if err := client.SignalEvent(ctx, input.Name); err != nil {
					return makeOutcomeResponse(err)
				}
				return makeOutcomeResponse(nil)
			})
		},
	}
}
