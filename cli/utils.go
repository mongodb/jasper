package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"os"
	"strings"

	"github.com/mongodb/grip"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/rpc"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	restService = "rest"
	rpcService  = "rpc"
)

// mergeBeforeFuncs returns a cli.BeforeFunc that runs all funcs and accumulates
// the errors.
func mergeBeforeFuncs(funcs ...cli.BeforeFunc) cli.BeforeFunc {
	return func(c *cli.Context) error {
		catcher := grip.NewBasicCatcher()
		for _, f := range funcs {
			catcher.Add(f(c))
		}
		return catcher.Resolve()
	}
}

// joinFlagNames joins multiple CLI flag names.
func joinFlagNames(names ...string) string {
	return strings.Join(names, ", ")
}

// validatePort validates that the flag given by the name is a valid port value.
func validatePort(flagName string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		port := c.GlobalInt(flagName)
		if port <= 0 || port > math.MaxUint16 {
			return errors.New("port must be between 0-65536 exclusive")
		}
		return nil
	}
}

// readInput reads JSON from the input and decodes it to the output.
func readInput(input io.Reader, output interface{}) error {
	bytes, err := ioutil.ReadAll(input)
	if err != nil {
		return errors.Wrap(err, "error reading from input")
	}
	return errors.Wrap(json.Unmarshal(bytes, output), "error decoding to output")
}

// writeOutput encodes the output as JSON and writes it to w.
func writeOutput(output io.Writer, input interface{}) error {
	bytes, err := json.MarshalIndent(input, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error encoding input")
	}
	if _, err := output.Write(bytes); err != nil {
		return errors.Wrap(err, "error writing to output")
	}

	return nil
}

// makeRemoteClient returns a remote client that connects to the service at the
// given host and port, with the optional SSL/TLS credentials file specified at
// the given location.
func makeRemoteClient(ctx context.Context, service, host string, port int, certFilePath string) (jasper.RemoteClient, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve address")
	}

	if service == restService {
		return jasper.NewRESTClient(addr), nil
	} else if service == rpcService {
		return rpc.NewClient(ctx, addr, certFilePath)
	}
	return nil, errors.Errorf("unrecognized service type '%s'", service)
}

// doPassthroughInputOutput passes input from standard input to the input validator,
// validates the input, runs the request, and writes the response of the request
// to standard output.
func doPassthroughInputOutput(c *cli.Context, input Validator, request func(context.Context, jasper.RemoteClient) (response interface{})) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := readInput(os.Stdin, input); err != nil {
		return errors.Wrap(err, "error reading from standard input")
	}
	if err := input.Validate(); err != nil {
		return errors.Wrap(err, "input is invalid")
	}

	return withConnection(ctx, c, func(client jasper.RemoteClient) error {
		return errors.Wrap(writeOutput(os.Stdout, request(ctx, client)), "error writing to standard output")
	})
}

// doPassthroughOutput runs the request and writes the output of the request to
// standard output.
func doPassthroughOutput(c *cli.Context, request func(context.Context, jasper.RemoteClient) (response interface{})) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return withConnection(ctx, c, func(client jasper.RemoteClient) error {
		return errors.Wrap(writeOutput(os.Stdout, request(ctx, client)), "error writing to standard output")
	})
}

// withConnection runs the operation within the scope of a remote client
// connection.
func withConnection(ctx context.Context, c *cli.Context, operation func(jasper.RemoteClient) error) error {
	host := c.GlobalString(hostFlagName)
	port := c.GlobalInt(portFlagName)
	service := c.GlobalString(serviceFlagName)
	certFilePath := c.GlobalString(certFilePathFlagName)

	client, err := makeRemoteClient(ctx, service, host, port, certFilePath)
	if err != nil {
		return errors.Wrap(err, "error setting up remote client")
	}

	catcher := grip.NewBasicCatcher()
	catcher.Add(operation(client))
	catcher.Add(client.CloseConnection())

	return catcher.Resolve()
}
