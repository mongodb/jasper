package cli

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Client encapsulates the client-side interface to a Jasper service. Operations
// read from standard input (if necessary) and write the result to standard
// output.
func Client() cli.Command {
	return cli.Command{
		Name:  "client",
		Usage: "tools for making requests to Jasper services",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  portFlagName,
				Usage: "the port running the Jasper service",
			},
			cli.StringFlag{
				Name:  hostFlagName,
				Usage: "the host running the Jasper service (default: localhost)",
				Value: defaultLocalHostName,
			},
		},
		Before: mergeBeforeFuncs(
			validatePort(portFlagName),
			func(c *cli.Context) error {
				service := c.String(serviceFlagName)
				if service != restService && service != rpcService {
					return errors.Errorf("service must be '%s' or '%s'", restService, rpcService)
				}
				return nil
			},
		),
		Subcommands: []cli.Command{
			Manager(),
			Process(),
		},
	}
}
