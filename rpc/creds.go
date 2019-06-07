package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// ServerCredentials represents TLS credentials to run the RPC server.
type ServerCredentials struct {
	// CACert is the PEM-encoded client CA certificate.
	CACert []byte `json:"ca_cert"`
	// ServerCert is the PEM-encoded server certificate.
	ServerCert []byte `json:"server_cert"`
	// ServerKey is the PEM-encoded server key.
	ServerKey []byte `json:"server_key"`
}

// Validate checks that the ServerCredentials are all set to non-empty values.
func (c *ServerCredentials) Validate() error {
	catcher := grip.NewBasicCatcher()
	if len(c.CACert) == 0 {
		catcher.New("CA certificate should not be empty")
	}
	if len(c.ServerCert) == 0 {
		catcher.New("server certificate should not be empty")
	}
	if len(c.ServerKey) == 0 {
		catcher.New("server key should not be empty")
	}
	return catcher.Resolve()
}

// Resolve converts the
func (c *ServerCredentials) Resolve() (*tls.Config, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid server credentials")
	}

	clientCACerts := x509.NewCertPool()
	if ok := clientCACerts.AppendCertsFromPEM(c.CACert); !ok {
		return nil, errors.New("failed to append client CA certificate")
	}

	cert, err := tls.X509KeyPair(c.ServerCert, c.ServerKey)
	if err != nil {
		return nil, errors.Wrap(err, "problem loading server key pair")
	}

	return &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    clientCACerts,
	}, nil
}

// Export exports the ServerCredentials into JSON-encoded bytes.
func (c *ServerCredentials) Export() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid server credentials")
	}

	b, err := json.Marshal(c)
	if err != nil {
		return nil, errors.Wrap(err, "error exporting server credentials")
	}

	return b, nil
}

// NewServerCredentialsFromFile parses the PEM-encoded credentials in JSON
// format in the file at path into a ServerCredentials struct.
func NewServerCredentialsFromFile(path string) (*ServerCredentials, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "error opening server credentials file")
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "error reading server credentials file")
	}

	creds := ServerCredentials{}
	if err := json.Unmarshal(contents, &creds); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling contents of credentials file")
	}

	if err := creds.Validate(); err != nil {
		return nil, errors.Wrap(err, "read invalid credentials from file")
	}

	return &creds, nil
}
