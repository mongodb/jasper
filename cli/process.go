package cli

import (
	"context"
	"syscall"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Process creates a cli.Command that interfaces with a Jasper process.
func Process() cli.Command {
	return cli.Command{
		Name:  "process",
		Flags: []cli.Flag{},
		Subcommands: []cli.Command{
			info(),
			running(),
			complete(),
			signal(),
			wait(),
			respawn(),
			registerSignalTriggerID(),
			tag(),
			getTags(),
			resetTags(),
		},
	}
}

func info() cli.Command {
	return cli.Command{
		Name: "info",
		Action: func(c *cli.Context) error {
			input := &IDInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				return InfoResponse{proc.Info(ctx)}
			})
		},
	}
}

func running() cli.Command {
	return cli.Command{
		Name: "running",
		Action: func(c *cli.Context) error {
			input := &IDInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				return RunningResponse{proc.Running(ctx)}
			})
		},
	}
}

func complete() cli.Command {
	return cli.Command{
		Name: "complete",
		Action: func(c *cli.Context) error {
			input := &IDInput{}
			return doPassthrough(c, &IDInput{}, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				return CompleteResponse{proc.Complete(ctx)}
			})
		},
	}
}

func signal() cli.Command {
	return cli.Command{
		Name: "signal",
		Action: func(c *cli.Context) error {
			input := &SignalInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				return ErrorResponse{proc.Signal(ctx, syscall.Signal(input.Signal))}
			})
		},
	}
}

func wait() cli.Command {
	return cli.Command{
		Name: "wait",
		Action: func(c *cli.Context) error {
			input := &IDInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				exitCode, err := proc.Wait(ctx)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error waiting on process with id '%s'", input.ID)}
				}
				return WaitResponse{ExitCode: exitCode}
			})
		},
	}
}

func respawn() cli.Command {
	return cli.Command{
		Name: "respawn",
		Action: func(c *cli.Context) error {
			input := &IDInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				newProc, err := proc.Respawn(ctx)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error respawning process with id '%s'", input.ID)}
				}
				return InfoResponse{newProc.Info(ctx)}
			})
		},
	}
}

func registerSignalTriggerID() cli.Command {
	return cli.Command{
		Name: "register-signal-trigger-id",
		Action: func(c *cli.Context) error {
			input := &SignalTriggerIDInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				return ErrorResponse{errors.Wrapf(proc.RegisterSignalTriggerID(ctx, input.SignalTriggerID), "couldn't register signal trigger with id '%s' on process with id '%s'", input.SignalTriggerID, input.ID)}
			})
		},
	}
}

func tag() cli.Command {
	return cli.Command{
		Name: "tag",
		Action: func(c *cli.Context) error {
			input := &TagIDInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				proc.Tag(input.Tag)
				return ErrorResponse{nil}
			})
		},
	}
}

func getTags() cli.Command {
	return cli.Command{
		Name: "get-tags",
		Action: func(c *cli.Context) error {
			input := &IDInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				return TagsResponse{Tags: proc.GetTags()}
			})
		},
	}
}

func resetTags() cli.Command {
	return cli.Command{
		Name: "get-tags",
		Action: func(c *cli.Context) error {
			input := &IDInput{}
			return doPassthrough(c, input, func(ctx context.Context, client jasper.RemoteClient) interface{} {
				proc, err := client.Get(ctx, input.ID)
				if err != nil {
					return ErrorResponse{errors.Wrapf(err, "error finding process with id '%s'", input.ID)}
				}
				return TagsResponse{Tags: proc.GetTags()}
			})
		},
	}
}
