package cli

import (
	"context"

	"github.com/kardianos/service"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceInstallCombined() cli.Command {
	return cli.Command{
		Name:  rpcService,
		Usage: "install a combined service",
		Flags: combinedServiceFlags(),
		Before: mergeBeforeFuncs(
			validatePort(restPortFlagName),
			validatePort(rpcPortFlagName),
		),
		Action: func(c *cli.Context) error {
			daemon := &combinedDaemon{
				RPCDaemon: &rpcDaemon{
					Host:         c.String(rpcHostFlagName),
					Port:         c.Int(rpcPortFlagName),
					CertFilePath: c.String(rpcCertFilePathFlagName),
					KeyFilePath:  c.String(rpcKeyFilePathFlagName),
				},
				RESTDaemon: &restDaemon{
					Host: c.String(restHostFlagName),
					Port: c.Int(restPortFlagName),
				},
			}
			config := serviceConfig([]string{"jasper", "service", "combined"})
			return install(daemon, config)
		},
	}
}

type combinedDaemon struct {
	RESTDaemon *restDaemon
	RPCDaemon  *rpcDaemon
}

func makeCombinedDaemon(rest *restDaemon, rpc *rpcDaemon) *combinedDaemon {
	return &combinedDaemon{
		RESTDaemon: rest,
		RPCDaemon:  rpc,
	}
}

func (d *combinedDaemon) Start(s service.Service) error {
	catcher := grip.NewBasicCatcher()
	catcher.Add(errors.Wrap(d.RPCDaemon.Start(s), "error starting RPC service"))
	catcher.Add(errors.Wrap(d.RESTDaemon.Start(s), "error starting REST service"))
	return catcher.Resolve()
}

func (d *combinedDaemon) Stop(s service.Service) error {
	catcher := grip.NewBasicCatcher()
	catcher.Add(errors.Wrap(d.RPCDaemon.Stop(s), "error stopping RPC service"))
	catcher.Add(errors.Wrap(d.RESTDaemon.Stop(s), "error stopping REST service"))
	return catcher.Resolve()
}

func (d *combinedDaemon) run(ctx context.Context) error {
	return errors.Wrap(runServices(ctx, d.RESTDaemon.makeService, d.RPCDaemon.makeService), "error running combined services")
}
