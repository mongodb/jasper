package remote

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/evergreen-ci/certdepot"
	"github.com/mongodb/grip"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/testutil"
	"github.com/pkg/errors"
)

func makeInsecureRPCServiceAndClient(ctx context.Context, mngr jasper.Manager) (Manager, error) {
	addr, err := tryStartRPCService(ctx, func(ctx context.Context, addr net.Addr) error {
		return startTestRPCService(ctx, mngr, addr, nil)
	})
	if err != nil {
		return nil, errors.Wrap(err, "starting RPC service")
	}
	return newTestRPCClient(ctx, addr, nil)
}

func tryStartRPCService(ctx context.Context, startService func(context.Context, net.Addr) error) (net.Addr, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", testutil.GetPortNumber()))
			if err != nil {
				continue
			}

			if err = startService(ctx, addr); err != nil {
				continue
			}

			return addr, err
		}
	}
}

func makeTLSRPCServiceAndClient(ctx context.Context, mngr jasper.Manager) (Manager, error) {
	caCertFile := filepath.Join("testdata", "ca.crt")

	serverCertFile := filepath.Join("testdata", "server.crt")
	serverKeyFile := filepath.Join("testdata", "server.key")

	clientCertFile := filepath.Join("testdata", "client.crt")
	clientKeyFile := filepath.Join("testdata", "client.key")

	// Make CA credentials
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading cert file")
	}

	// Make server credentials
	serverCert, err := os.ReadFile(serverCertFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading cert file")
	}
	serverKey, err := os.ReadFile(serverKeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading key file")
	}
	serverCreds, err := certdepot.NewCredentials(caCert, serverCert, serverKey)
	if err != nil {
		return nil, errors.Wrap(err, "initializing test server credentials")
	}

	addr, err := tryStartRPCService(ctx, func(ctx context.Context, addr net.Addr) error {
		return startTestRPCService(ctx, mngr, addr, serverCreds)
	})
	if err != nil {
		return nil, errors.Wrap(err, "starting RPC service")
	}

	clientCert, err := os.ReadFile(clientCertFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading cert file")
	}
	clientKey, err := os.ReadFile(clientKeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading key file")
	}
	clientCreds, err := certdepot.NewCredentials(caCert, clientCert, clientKey)
	if err != nil {
		return nil, errors.Wrap(err, "initializing test client credentials")
	}

	return newTestRPCClient(ctx, addr, clientCreds)
}

// startTestService creates a server for testing purposes that terminates when
// the context is done.
func startTestRPCService(ctx context.Context, mngr jasper.Manager, addr net.Addr, creds *certdepot.Credentials) error {
	closeService, err := StartRPCService(ctx, mngr, addr, creds)
	if err != nil {
		return errors.Wrap(err, "starting server")
	}

	go func() {
		<-ctx.Done()
		grip.Error(closeService())
	}()

	return nil
}

// newTestRPCClient establishes a client for testing purposes that closes when
// the context is done.
func newTestRPCClient(ctx context.Context, addr net.Addr, creds *certdepot.Credentials) (Manager, error) {
	client, err := NewRPCClient(ctx, addr, creds)
	if err != nil {
		return nil, errors.Wrap(err, "getting client")
	}

	go func() {
		<-ctx.Done()
		grip.Notice(client.CloseConnection())
	}()

	return client, nil
}
