package cli

import (
	"context"
	"fmt"
	"net"

	"github.com/kardianos/service"
	"github.com/mongodb/grip"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/rpc"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceInstallRPC() cli.Command {
	return cli.Command{
		Name:   rpcService,
		Usage:  "install an RPC service",
		Flags:  rpcServiceFlags(),
		Before: validatePort(portFlagName),
		Action: func(c *cli.Context) error {
			daemon := &rpcDaemon{
				Host: c.String(hostFlagName),
				Port: c.Int(portFlagName),
			}
			config := serviceConfig([]string{"jasper", "service", "rpc"})
			return install(daemon, config)
		},
	}
}

type rpcDaemon struct {
	Host         string
	Port         int
	KeyFilePath  string
	CertFilePath string

	manager jasper.Manager
	exit    chan struct{}
}

func makeRPCDaemon(host string, port int, certFilePath, keyFilePath string, manager jasper.Manager) *rpcDaemon {
	return &rpcDaemon{
		Host:         host,
		Port:         port,
		CertFilePath: certFilePath,
		KeyFilePath:  keyFilePath,
		manager:      manager,
	}
}

func (d *rpcDaemon) Start(s service.Service) error {
	d.exit = make(chan struct{})
	var err error
	if d.manager, err = jasper.NewLocalManager(false); err != nil {
		return errors.Wrap(err, "failed to construct RPC manager")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go handleDaemonSignals(ctx, cancel, d.exit)

	go func(ctx context.Context, d *rpcDaemon) {
		grip.Error(errors.Wrap(d.run(ctx), "error running RPC service"))
	}(ctx, d)

	return nil
}

func (d *rpcDaemon) Stop(s service.Service) error {
	close(d.exit)
	return nil
}

func (d *rpcDaemon) run(ctx context.Context) error {
	return errors.Wrap(runServices(ctx, d.makeService), "error running RPC service")
}

func (d *rpcDaemon) makeService(ctx context.Context) (jasper.CloseFunc, error) {
	if d.manager == nil {
		return nil, errors.New("manager is not set on RPC service")
	}
	grip.Infof("starting RPC service at '%s:%d'", d.Host, d.Port)
	return makeRPCService(ctx, d.Host, d.Port, d.manager, d.CertFilePath, d.KeyFilePath)
}

// makeRPCService creates an RPC service around the manager serving requests on
// the host and port.
func makeRPCService(ctx context.Context, host string, port int, manager jasper.Manager, certFilePath, keyFilePath string) (jasper.CloseFunc, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve RPC address")
	}

	closeService, err := rpc.StartService(ctx, manager, addr, certFilePath, keyFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "error starting RPC service")
	}

	return closeService, nil
}
