package cli

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const version = "2019-04-18"

//
// func main() {
//     app := makeApp()
//     grip.EmergencyFatal(app.Run(os.Args))
// }
//
// // TODO: potentially make this a cli.Command to be used in Curator OR just
// write this code in Curator.
func makeApp() *cli.App {
	app := cli.NewApp()
	app.Name = "jasper"
	app.Usage = "Jasper CLI to interact with Jasper services"
	app.Version = version

	app.Commands = []cli.Command{
		Process(),
		Manager(),
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  joinFlagNames(hostFlagName, "h"),
			Usage: "the host that is running the Jasper service (default: \"localhost\")",
			Value: "localhost",
		},
		cli.IntFlag{
			Name:  joinFlagNames(portFlagName, "p"),
			Usage: "the port that is running the Jasper service",
		},
		cli.StringFlag{
			Name:  joinFlagNames(serviceFlagName, "s"),
			Usage: "the service to contact (either 'rest' or 'rpc')",
		},
	}
	app.Before = mergeBeforeFuncs(
		func(c *cli.Context) error {
			port := c.Int(portFlagName)
			minPort, maxPort := 0, 1<<16-1
			if port < minPort || port > maxPort {
				return errors.New("port must be within 0-65535 inclusive")
			}
			return nil
		},
		func(c *cli.Context) error {
			service := c.String(serviceFlagName)
			if service != serviceREST && service != serviceRPC {
				return errors.New("service must be 'rest' or 'rpc'")
			}
			return nil
		},
	)

	return app
}

// // unpackString returns the string value from input if it is a string.
// func unpackString(input map[string]interface{}, key string) (string, error) {
//     val, ok := input[key]
//     if !ok {
//         return "", errors.Errorf("could not find field '%s' in input", key)
//     }
//     str, ok := val.(string)
//     if !ok {
//         return "", errors.Errorf("'%s' is of type %T, but expected string", key)
//     }
//     return str, nil
// }
//
// func unpackInt(input map[string]interface{}, key string) (int, error) {
//     val, ok := input[key]
//     if !ok {
//         return "", errors.Errorf("could not find field '%s' in input", key)
//     }
//     i, ok := val.(int)
//     if !ok {
//         return "", errors.Errorf("'%s' is of type %T, but expected int", key)
//     }
//     return i, nil
// }

// func Jasper() cli.Command {
//     var manager jasper.Manager
//     fmt.Println(manager)
//
//     return cli.Command{
//         Name:   "jasper",
//         Flags:  []cli.Flag{},
//         Before: func(c *cli.Context) error { return nil },
//         Subcommands: []cli.Command{
//             {
//                 Name: "create-process",
//             },
//             {
//                 Name: "create-command",
//             },
//             {
//                 Name: "register-process",
//             },
//             {
//                 Name: "list-processes",
//             },
//             {
//                 Name: "list-group-process",
//             },
//             {
//                 Name: "get-process",
//             },
//             {
//                 Name: "clear-manager",
//             },
//             {
//                 Name: "close-manager",
//             },
//             {
//                 Name: "process",
//                 Subcommands: []cli.Command{
//                     {
//                         Name: "info",
//                     },
//                     {
//                         Name: "is-running",
//                     },
//                     {
//                         Name: "is-complete",
//                     },
//                     {
//                         Name: "is-signal",
//                     },
//                     {
//                         Name: "wait",
//                     },
//                     {
//                         Name: "tag",
//                     },
//                     {
//                         Name: "get-tags",
//                     },
//                     {
//                         Name: "reset-tags",
//                     },
//                     {
//                         Name: "respawn",
//                     },
//                     {
//                         Name: "register-signal-trigger",
//                     },
//                 },
//             },
//         },
//     }
// }
