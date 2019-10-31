package mongorpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"

	"github.com/evergreen-ci/birch"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/grip/recovery"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc/mongowire"
)

type HandlerFunc func(context.Context, io.Writer, mongowire.Message)

type Service struct {
	addr     string
	registry *OperationRegistry
}

func NewService(listenAddr string, port int) *Service {
	return &Service{
		addr:     fmt.Sprintf("%s:%d", listenAddr, port),
		registry: &OperationRegistry{ops: make(map[mongowire.OpScope]HandlerFunc)},
	}
}

func (s *Service) Address() string { return s.addr }

func (s *Service) RegisterOperation(scope *mongowire.OpScope, h HandlerFunc) error {
	return errors.WithStack(s.registry.Add(*scope, h))
}

func (s *Service) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return errors.Wrapf(err, "problem listening on %s", s.addr)
	}
	defer l.Close()

	grip.Infof("listening for connections on %s", s.addr)

	for {
		if ctx.Err() != nil {
			return errors.New("service terminated by canceled context")
		}

		conn, err := l.Accept()
		if err != nil {
			grip.Warning(message.WrapError(err, "problem accepting connection"))
			continue
		}

		go s.dispatchRequest(ctx, conn)
	}
}

func writeErrorReply(w io.Writer, err error) error {
	responseNotOk := birch.EC.Int32("ok", 0)
	errorDoc := birch.EC.String("error", err.Error())
	doc := birch.NewDocument(responseNotOk, errorDoc)

	reply := mongowire.NewReply(int64(0), int32(0), int32(0), int32(1), []*birch.Document{doc})
	_, err = w.Write(reply.Serialize())
	return errors.Wrap(err, "could not write response")
}

func (s *Service) dispatchRequest(ctx context.Context, conn net.Conn) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	defer func() {
		err := recovery.HandlePanicWithError(recover(), nil, "request handling")
		if err != nil {
			grip.Error(message.WrapError(err, "error during request handling"))
			if writeErr := writeErrorReply(conn, err); writeErr != nil {
				grip.Error(message.WrapError(writeErr, "error writing reply after panic recovery"))
				return
			}
		}
		if err := conn.Close(); err != nil {
			grip.Error(message.WrapErrorf(err, "error closing connection from %s", conn.RemoteAddr()))
			return
		}
		grip.Infof("closed connection from %s", conn.RemoteAddr())
	}()

	if c, ok := conn.(*tls.Conn); ok {
		// we do this here so that we can get the SNI server name
		if err := c.Handshake(); err != nil {
			grip.Warning(message.WrapError(err, "error doing tls handshake"))
			return
		}
		grip.Debugf("ssl connection to %s", c.ConnectionState().ServerName)
	}

	for {
		m, err := mongowire.ReadMessage(conn)
		if err != nil {
			if errors.Cause(err) == io.EOF {
				return
			}
			grip.Error(message.WrapError(err, "problem reading message"))
			return
		}

		scope := m.Scope()

		handler, ok := s.registry.Get(scope)
		if !ok {
			grip.Warningf("undefined command scope: %+v", scope)
			return
		}

		handler(ctx, conn, m)
	}
}
