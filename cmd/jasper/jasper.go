package main

import (
	"os"

	"github.com/mongodb/grip"
	jcli "github.com/mongodb/jasper/cli"
	"github.com/urfave/cli"
)

func main() {
	app := newApp()
	grip.Error(app.Run(os.Args))
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "jasper"
	app.Usage = "The Jasper client."
	app.Commands = jcli.Jasper().Subcommands
	return app
}
