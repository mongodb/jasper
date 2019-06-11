package cli

import (
	"context"
	"strings"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func RunCMD() cli.Command {
	const commandFlagName = "command"

	return cli.Command{
		Name:  "run",
		Usage: "Run a command with Jasper",
		Flags: append(clientFlags(),
			cli.StringFlag{
				Name:  joinFlagNames(commandFlagName, "c"),
				Usage: "specify a command to run on a remote jasper service",
			},
			cli.StringSliceFlag{
				Name:  "env",
				Usage: "specify environment variables, in '<key>=<val>' forms. may specify more than once",
			},
			cli.BoolFlag{
				Name:  "sudo",
				Usage: "use this flag to run the command with sudo",
			},
			cli.StringFlag{
				Name:  "sudoAs",
				Usage: "use this to run commands as another user as in 'sudo -u <user>'",
			},
			cli.StringFlag{
				Name:  "id",
				Usage: "specify an id for this process (optional)",
			},
			cli.BoolTFlag{
				Usage: "disableShell",
			},
		),
		Before: mergeBeforeFuncs(clientBefore(),
			func(c *cli.Context) error {
				if c.String(commandFlagName) == "" {
					if c.NArg() != 1 {
						return errors.New("must specify a command")
					}
					return errors.Wrap(c.Set(commandFlagName, strings.Join(c.Args().Tail(), " ")), "problem setting command")
				}
				return nil
			}),
		Action: func(c *cli.Context) error {
			envvars := c.StringSlice("env")
			cmdStr := c.String(commandFlagName)
			useSudo := c.Bool("sudo")
			sudoAs := c.String("sudoAs")
			cmdID := c.String("id")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			return withConnection(ctx, c, func(client jasper.RemoteClient) error {
				cmd := client.CreateCommand(ctx).Sudo(useSudo).ID(cmdID).Append(cmdStr)

				if sudoAs != "" {
					cmd.SudoAs(sudoAs)
				}

				for _, e := range envvars {
					parts := strings.SplitN(e, "=", 2)
					cmd.AddEnv(parts[0], parts[1])
				}

				return cmd.Run(ctx)
			})
		},
	}
}
