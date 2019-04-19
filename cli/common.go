package cli

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/mongodb/grip"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/rpc"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	hostFlagName         = "flag"
	portFlagName         = "port"
	serviceFlagName      = "service"
	certFilePathFlagName = "cert_path"

	serviceREST = "rest"
	serviceRPC  = "rpc"
)

// remoteClient returns a remote client that connects to the service at the
// given host and port, with the optional SSL/TLS credentials file specified at
// the given location.
func remoteClient(ctx context.Context, service, host string, port int, certFilePath string) (jasper.RemoteClient, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve address")
	}

	if service == serviceREST {
		return jasper.NewRESTClient(addr), nil
	} else if service == serviceRPC {
		return rpc.NewClient(ctx, addr, certFilePath)
	}
	return nil, errors.Errorf("unrecognized service type '%s'", service)
}

// doPassthrough passes input from stdin to the input validator, validates the
// input, runs the request, and writes the response of the request to stdout.
func doPassthrough(c *cli.Context, input Validator, request func(context.Context, jasper.RemoteClient) interface{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := readInput(os.Stdin, input); err != nil {
		return errors.Wrap(err, "error reading from stdin")
	}
	if err := input.Validate(); err != nil {
		return errors.Wrap(err, "input is invalid")
	}

	return withConnection(ctx, c, func(client jasper.RemoteClient) error {
		return errors.Wrap(writeOutput(os.Stdout, request(ctx, client)), "error writing to stdout")
	})
}

// doPassthrough runs the request and writes the output of the request to
// stdout.
func doPassthroughNoInput(c *cli.Context, request func(context.Context, jasper.RemoteClient) interface{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return withConnection(ctx, c, func(client jasper.RemoteClient) error {
		return errors.Wrap(writeOutput(os.Stdout, request(ctx, client)), "error writing to stdout")
	})
}

// withConnection runs the operation within the scope of a remote client
// connection.
func withConnection(ctx context.Context, c *cli.Context, operation func(jasper.RemoteClient) error) error {
	host := c.Parent().String(hostFlagName)
	port := c.Parent().Int(portFlagName)
	service := c.Parent().String(serviceFlagName)
	certFilePath := c.Parent().String(certFilePathFlagName)

	client, err := remoteClient(ctx, service, host, port, certFilePath)
	if err != nil {
		return errors.Wrap(err, "error setting up remote client")
	}

	catcher := grip.NewBasicCatcher()
	catcher.Add(operation(client))
	catcher.Add(client.CloseConnection())

	return catcher.Resolve()
}
