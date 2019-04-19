package cli

import (
	"context"

	"github.com/mongodb/grip/level"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Manager creates a cli.Command that interfaces with a Jasper manager.
func Manager() cli.Command {
	return cli.Command{
		Name:  "manager",
		Flags: []cli.Flag{},
		Subcommands: []cli.Command{
			createProcess(),
			createCommand(),
			get(),
			list(),
			group(),
			clear(),
			closeManager(),
		},
	}
}

func createProcess() cli.Command {
	return cli.Command{
		Name: "create-process",
		Action: func(c *cli.Context) error {
			opts := &jasper.CreateOptions{}
			return doPassthrough(c, opts, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.CreateProcess(ctx, opts)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error creating process")}
				}
				return InfoResponse{Info: proc.Info(ctx)}
			})
		},
	}
}

func createCommand() cli.Command {
	return cli.Command{
		Name: "create-command",
		Action: func(c *cli.Context) error {
			opts := &CommandInput{}
			return doPassthrough(c, opts, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				cmd := client.CreateCommand(ctx).Background(opts.Background)
				if level.IsValidPriority(opts.Priority) {
					cmd = cmd.Priority(opts.Priority)
				}
				cmd = cmd.Background(opts.Background).
					ContinueOnError(opts.ContinueOnError).
					IgnoreError(opts.IgnoreError).
					ApplyFromOpts(&opts.CreateOptions)
				return ErrorResponse{cmd.Run(ctx)}
			})
		},
	}
}

func get() cli.Command {
	return cli.Command{
		Name: "get",
		Action: func(c *cli.Context) error {
			input := &IDInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Errorf("error getting process with ID '%s'", input.ID)}
				}
				return InfoResponse{Info: proc.Info(ctx)}
			})
		},
	}
}

func list() cli.Command {
	return cli.Command{
		Name: "list",
		Action: func(c *cli.Context) error {
			input := &FilterInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				procs, err := client.List(ctx, input.Filter)
				if err != nil {
					return ErrorResponse{errors.Errorf("error listing processes with filter '%s'", input.Filter)}
				}
				infos := make([]jasper.ProcessInfo, 0, len(procs))
				for _, proc := range procs {
					infos = append(infos, proc.Info(ctx))
				}
				return InfosResponse{Infos: infos}
			})
		},
	}
}

func group() cli.Command {
	return cli.Command{
		Name: "group",
		Action: func(c *cli.Context) error {
			input := &TagInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				procs, err := client.Group(ctx, input.Tag)
				if err != nil {
					return ErrorResponse{errors.Errorf("error grouping processes with tag '%s'", input.Tag)}
				}
				infos := make([]jasper.ProcessInfo, 0, len(procs))
				for _, proc := range procs {
					infos = append(infos, proc.Info(ctx))
				}
				return InfosResponse{Infos: infos}
			})
		},
	}
}

func clear() cli.Command {
	return cli.Command{
		Name: "clear",
		Action: func(c *cli.Context) error {
			return doPassthroughNoInput(c, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				client.Clear(ctx)
				return ErrorResponse{nil}
			})
		},
	}
}

func closeManager() cli.Command {
	return cli.Command{
		Name: "close",
		Action: func(c *cli.Context) error {
			return doPassthroughNoInput(c, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				return ErrorResponse{client.Close(ctx)}
			})
		},
	}
}
