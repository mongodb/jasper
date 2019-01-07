package jrpc

import (
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/jrpc/internal"
	"github.com/pkg/errors"
	grpc "google.golang.org/grpc"
)

// AttachService attaches the given manager to the jasper GRPC server. After
// this function successfully returns, calls to Manager functions will be sent
// over GRPC to the Jasper GRPC server.
func AttachService(manager jasper.Manager, s *grpc.Server) error {
	return errors.WithStack(internal.AttachService(manager, s))
}
