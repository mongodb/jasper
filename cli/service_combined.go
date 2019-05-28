package cli

import (
	"context"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	restHostFlagName = "rest_host"
	restPortFlagName = "rest_port"

	rpcHostFlagName         = "rpc_host"
	rpcPortFlagName         = "rpc_port"
	rpcKeyFilePathFlagName  = "rpc_key_path"
	rpcCertFilePathFlagName = "rpc_cert_path"
)

func serviceCombined() cli.Command {
	return cli.Command{
		Name:  combinedService,
		Usage: "run a combined multiprotocol service",
		Flags: combinedServiceFlags(),
		Before: mergeBeforeFuncs(
			validatePort(restPortFlagName),
			validatePort(rpcPortFlagName),
		),
		Action: func(c *cli.Context) error {
			manager, err := jasper.NewLocalManager(false)
			if err != nil {
				return errors.Wrap(err, "failed to construct combined manager")
			}

			daemon := makeCombinedDaemon(
				makeRESTDaemon(
					c.String(restHostFlagName),
					c.Int(restPortFlagName),
					manager,
				),
				makeRPCDaemon(
					c.String(rpcHostFlagName),
					c.Int(rpcPortFlagName),
					c.String(rpcCertFilePathFlagName),
					c.String(rpcKeyFilePathFlagName),
					manager,
				),
			)

			ctx, cancel := context.WithCancel(context.Background())
			go handleSignals(ctx, cancel)

			return daemon.run(ctx)
		},
	}
}
