package cli

import (
	"context"

	"github.com/kardianos/service"
	"github.com/mongodb/grip"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceInstallREST() cli.Command {
	return cli.Command{
		Name:   restService,
		Usage:  "install a REST service",
		Flags:  restServiceFlags(),
		Before: validatePort(portFlagName),
		Action: func(c *cli.Context) error {
			daemon := &restDaemon{
				Host: c.String(hostFlagName),
				Port: c.Int(portFlagName),
			}
			config := serviceConfig([]string{"jasper", "service", "rest"})
			return install(daemon, config)
		},
	}
}

type restDaemon struct {
	Host string
	Port int

	manager jasper.Manager
	exit    chan struct{}
}

func makeRESTDaemon(host string, port int, manager jasper.Manager) *restDaemon {
	return &restDaemon{
		Host:    host,
		Port:    port,
		manager: manager,
	}
}

func (d *restDaemon) Start(s service.Service) error {
	d.exit = make(chan struct{})
	var err error
	if d.manager, err = jasper.NewLocalManager(false); err != nil {
		return errors.Wrap(err, "failed to construct REST manager")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go handleDaemonSignals(ctx, cancel, d.exit)

	go func(ctx context.Context, d *restDaemon) {
		grip.Error(errors.Wrap(d.run(ctx), "error running REST service"))
	}(ctx, d)

	return nil
}

func (d *restDaemon) Stop(s service.Service) error {
	close(d.exit)
	return nil
}

func (d *restDaemon) run(ctx context.Context) error {
	return errors.Wrap(runServices(ctx, d.makeService), "error running REST service")
}

func (d *restDaemon) makeService(ctx context.Context) (jasper.CloseFunc, error) {
	if d.manager == nil {
		return nil, errors.New("manager is not set on REST service")
	}
	grip.Infof("starting REST service at '%s:%d'", d.Host, d.Port)
	return makeRESTService(ctx, d.Host, d.Port, d.manager)
}

// makeRESTService creates a REST service around the manager serving requests on
// the host and port.
func makeRESTService(ctx context.Context, host string, port int, manager jasper.Manager) (jasper.CloseFunc, error) {
	service := jasper.NewManagerService(manager)
	app := service.App(ctx)
	app.SetPrefix("jasper")
	if err := app.SetHost(host); err != nil {
		return nil, errors.Wrap(err, "error setting REST host")
	}
	if err := app.SetPort(port); err != nil {
		return nil, errors.Wrap(err, "error setting REST port")
	}

	go func() {
		grip.Warning(errors.Wrap(app.Run(ctx), "error running REST app"))
	}()

	return func() error { return nil }, nil
}
