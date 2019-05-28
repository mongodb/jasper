package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kardianos/service"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/recovery"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	checkInterval = 10 * time.Second

	programFlagName = "program"
	argsFlagName    = "args"
)

// handleDaemonSignals shuts down the daemon by cancelling the context, either
// when the context is done, it receives a terminate signal, or when it
// receives a signal to exit the daemon.
func handleDaemonSignals(ctx context.Context, cancel context.CancelFunc, exit chan struct{}) {
	defer recovery.LogStackTraceAndContinue("graceful shutdown")
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)

	select {
	case <-sig:
		grip.Debug("received signal")
	case <-ctx.Done():
		grip.Debug("context canceled")
	case <-exit:
		grip.Debug("received daemon exit signal")
	}
}

// runServices starts the given services, waits until the context is done, and
// closes all the running services.
func runServices(ctx context.Context, makeServices ...func(context.Context) (jasper.CloseFunc, error)) error {
	closeServices := []jasper.CloseFunc{}
	closeAllServices := func(closeServices []jasper.CloseFunc) error {
		catcher := grip.NewBasicCatcher()
		for _, closeService := range closeServices {
			catcher.Add(errors.Wrap(closeService(), "error closing service"))
		}
		return catcher.Resolve()
	}

	for _, makeService := range makeServices {
		closeService, err := makeService(ctx)
		if err != nil {
			catcher := grip.NewBasicCatcher()
			catcher.Wrap(err, "failed to create service")
			catcher.Add(closeAllServices(closeServices))
			return catcher.Resolve()
		}
		closeServices = append(closeServices, closeService)
	}

	<-ctx.Done()
	return closeAllServices(closeServices)
}

// serviceOptions returns all the service options, including all options
// specific to particular service management systems.
func serviceOptions() service.KeyValue {
	return service.KeyValue{
		// sysv-specific options

		// upstart-specific options

		// systemd-specific options
		"Restart":         "always",
		"RestartSec":      30,
		"TimeoutStartSec": 10,
	}
}

// serviceDependencies returns the list of service dependencies.
func serviceDependencies() []string {
	return []string{
		"Requires=network.target",
		"After=network-online.target",
	}
}

// serviceConfig returns the daemon service configuration.
func serviceConfig(args []string) *service.Config {
	return &service.Config{
		Name:         "jasperd",
		DisplayName:  "Jasper service",
		Description:  "Jasper is a process management service",
		Executable:   "", // No executable refers to the current executable.
		Arguments:    args,
		Dependencies: serviceDependencies(),
		Option:       serviceOptions(),
	}
}

// serviceInstall returns a cli.Command to install a Jasper service as a daemon.
// This will typically require elevated privileges.
func serviceInstall() cli.Command {
	return cli.Command{
		Name:  "install",
		Usage: "install a daemon service",
		Subcommands: []cli.Command{
			serviceInstallRPC(),
			serviceInstallREST(),
			serviceInstallCombined(),
		},
	}
}

// install installs the service based on the platform it is running on.
func install(daemon service.Interface, config *service.Config) error {
	svc, err := service.New(daemon, config)
	if err != nil {
		return errors.Wrap(err, "error initializing service to install")
	}
	return svc.Install()
}
