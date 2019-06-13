package cli

import (
	"github.com/urfave/cli"
)

const (
	// JasperCommand represents the Jasper interface as a CLI command.
	JasperCommand = "jasper"

	hostFlagName          = "host"
	portFlagName          = "port"
	credsFilePathFlagName = "creds_path"

	defaultLocalHostName = "localhost"
)

// Jasper is the CLI interface to Jasper services. This is used by
// jasper.Manager implementations that manage processes remotely via
// commands via command executions. The interface is designed for
// machine interaction.
func Jasper() cli.Command {
	return cli.Command{
		Name:  JasperCommand,
		Usage: "Jasper CLI to interact with Jasper services",
		Subcommands: []cli.Command{
			Client(),
			Service(),
		},
	}
}

// JasperCMD is a user-facing set of commands for exploring and
// running operations with Japser on a local or remote system, and is
// designed to support human users.
func JasperCMD() cli.Command {
	return cli.Command{
		Name:  JasperCommand,
		Usage: "Jasper Command CLI to create and manage Jasper processes",
		Subcommands: []cli.Command{
			Service(),
			RunCMD(),
			ListCMD(),
			KillCMD(),
			KillAllCMD(),
		},
	}
}
