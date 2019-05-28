package cli

import (
	"context"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	envVarRESTHost  = "JASPER_REST_HOST"
	envVarRESTPort  = "JASPER_REST_PORT"
	defaultRESTPort = 2287
)

func serviceREST() cli.Command {
	return cli.Command{
		Name:   restService,
		Usage:  "run a REST service",
		Flags:  restServiceFlags(),
		Before: validatePort(portFlagName),
		Action: func(c *cli.Context) error {
			manager, err := jasper.NewLocalManager(false)
			if err != nil {
				return errors.Wrap(err, "failed to construct REST manager")
			}

			daemon := makeRESTDaemon(
				c.String(hostFlagName),
				c.Int(portFlagName),
				manager,
			)

			ctx, cancel := context.WithCancel(context.Background())
			go handleSignals(ctx, cancel)

			return daemon.run(ctx)
		},
	}
}
