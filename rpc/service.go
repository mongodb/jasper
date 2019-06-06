package rpc

import (
	"context"
	"net"

	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/rpc/internal"
	"github.com/pkg/errors"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// AttachService attaches the jasper GRPC server to the given manager. After
// this function successfully returns, calls to Manager functions will be sent
// over GRPC to the Jasper GRPC server.
func AttachService(manager jasper.Manager, s *grpc.Server) error {
	return errors.WithStack(internal.AttachService(manager, s))
}

// StartService starts and RPC server with the specified address. If path is
// non-empty, the credentials will be read from the file to start a secure TLS
// service requiring client and server credentials; otherwise, it will start an
// insecure service. The credentials file should contain the exported
// ServerCredentials. The caller is responsible for closing the connection using
// the return jasper.CloseFunc.
func StartService(ctx context.Context, manager jasper.Manager, address net.Addr, path string) (jasper.CloseFunc, error) {
	lis, err := net.Listen(address.Network(), address.String())
	if err != nil {
		return nil, errors.Wrapf(err, "error listening on %s", address.String())
	}

	opts := []grpc.ServerOption{}
	if path != "" {
		creds, err := NewServerCredentialsFromFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting server credentials from file '%s'", path)
		}
		tlsCreds, err := creds.Resolve()
		if err != nil {
			return nil, errors.Wrap(err, "error generating TLS config from server credentials")
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsCreds)))
	}

	service := grpc.NewServer(opts...)

	if err := AttachService(manager, service); err != nil {
		return nil, errors.Wrap(err, "could not attach manager to service")
	}
	go service.Serve(lis)

	return func() error { service.Stop(); return nil }, nil
}
