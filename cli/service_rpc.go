package cli

import (
	"context"

	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	keyFilePathFlagName = "key_path"

	envVarRPCHost  = "JASPER_RPC_HOST"
	envVarRPCPort  = "JASPER_RPC_PORT"
	defaultRPCPort = 2286
)

func serviceRPC() cli.Command {
	return cli.Command{
		Name:   rpcService,
		Usage:  "run an RPC service",
		Flags:  rpcServiceFlags(),
		Before: validatePort(portFlagName),
		Action: func(c *cli.Context) error {
			manager, err := jasper.NewLocalManager(false)
			if err != nil {
				return errors.Wrap(err, "failed to construct RPC manager")
			}

			daemon := makeRPCDaemon(
				c.String(hostFlagName),
				c.Int(portFlagName),
				c.String(certFilePathFlagName),
				c.String(keyFilePathFlagName),
				manager,
			)

			ctx, cancel := context.WithCancel(context.Background())
			go handleSignals(ctx, cancel)

			return daemon.run(ctx)
		},
	}
}
