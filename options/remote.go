package options

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type remoteConfig struct {
	Host          string
	User          string
	Port          int
	Key           string
	KeyFile       string
	KeyPassphrase string
	Password      string
	// Connection timeout
	Timeout time.Duration
}

// Remote represents options to SSH into a remote machine.
// kim: TODO: support proxy remoteConfig
type Remote struct {
	remoteConfig
	Proxy *Proxy
}

// Proxy represents the remote configuration to access a remote proxy machine.
type Proxy struct {
	remoteConfig
}

const defaultSSHPort = 22

func (opts *remoteConfig) validate() error {
	catcher := grip.NewBasicCatcher()
	if opts.Host == "" {
		catcher.New("host cannot be empty")
	}
	if opts.Port == 0 {
		opts.Port = defaultSSHPort
	}
	numAuthMethods := 0
	for _, authMethod := range []string{opts.Key, opts.KeyFile, opts.Password} {
		if authMethod != "" {
			numAuthMethods++
		}
	}
	if numAuthMethods != 1 {
		catcher.Errorf("must specify exactly one authentication method, found %d", numAuthMethods)
	}
	if opts.Key == "" && opts.KeyFile == "" && opts.KeyPassphrase != "" {
		catcher.New("cannot set passphrase without specifying key or key file")
	}
	return catcher.Resolve()
}

func (opts *remoteConfig) resolve() (*ssh.ClientConfig, error) {
	var auth []ssh.AuthMethod
	if opts.Key != "" || opts.KeyFile != "" {
		pubkey, err := opts.publicKeyAuth()
		if err != nil {
			return nil, errors.Wrap(err, "coudl not get public key")
		}
		auth = append(auth, pubkey)
	}
	if opts.Password != "" {
		auth = append(auth, ssh.Password(opts.Password))
	}
	return &ssh.ClientConfig{
		Timeout:         opts.Timeout,
		User:            opts.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (opts *remoteConfig) publicKeyAuth() (ssh.AuthMethod, error) {
	var key []byte
	if opts.KeyFile != "" {
		var err error
		key, err = ioutil.ReadFile(opts.KeyFile)
		if err != nil {
			return nil, errors.Wrap(err, "could not read key file")
		}
	} else {
		key = []byte(opts.Key)
	}

	var signer ssh.Signer
	var err error
	if opts.KeyPassphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(opts.KeyPassphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(key)
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not get signer")
	}
	return ssh.PublicKeys(signer), nil
}

// Validate ensures that enough information is provided to connect to a remote
// host.
func (opts *Remote) Validate() error {
	catcher := grip.NewBasicCatcher()

	if opts.Proxy != nil {
		catcher.Wrap(opts.Proxy.validate(), "invalid proxy config")
	}

	catcher.Add(opts.validate())
	return catcher.Resolve()
}

func (opts *Remote) String() string {
	if opts.User == "" {
		return opts.Host
	}

	return fmt.Sprintf("%s@%s", opts.User, opts.Host)
}

// Resolve returns the SSH client and session from the options.
func (opts *Remote) Resolve() (*ssh.Client, *ssh.Session, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, errors.Wrap(err, "invalid remote options")
	}

	var client *ssh.Client
	if opts.Proxy != nil {
		proxyConfig, err := opts.Proxy.resolve()
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not create proxy config")
		}
		proxyClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", opts.Proxy.Host, opts.Proxy.Port), proxyConfig)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not dial proxy")
		}

		targetConn, err := proxyClient.Dial("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.Port))
		if err != nil {
			catcher := grip.NewBasicCatcher()
			catcher.Wrap(proxyClient.Close(), "error closing connection to proxy")
			catcher.Wrap(err, "could not dial target host")
			return nil, nil, catcher.Resolve()
		}

		targetConfig, err := opts.resolve()
		if err != nil {
			catcher := grip.NewBasicCatcher()
			catcher.Wrap(proxyClient.Close(), "error closing connection to proxy")
			catcher.Wrap(err, "could not create target config")
			return nil, nil, catcher.Resolve()
		}
		gatewayConn, chans, reqs, err := ssh.NewClientConn(targetConn, fmt.Sprintf("%s:%d", opts.Host, opts.Port), targetConfig)
		if err != nil {
			catcher := grip.NewBasicCatcher()
			catcher.Wrap(targetConn.Close(), "error closing connection to target")
			catcher.Wrap(proxyClient.Close(), "error closing connection to proxy")
			catcher.Wrap(err, "could not establish connection to target via proxy")
			return nil, nil, catcher.Resolve()
		}
		client = ssh.NewClient(gatewayConn, chans, reqs)
	} else {
		var err error
		config, err := opts.resolve()
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not create config")
		}
		client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.Port), config)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not dial host")
		}
	}

	session, err := client.NewSession()
	if err != nil {
		catcher := grip.NewBasicCatcher()
		catcher.Add(client.Close())
		catcher.Add(err)
		return nil, nil, errors.Wrap(catcher.Resolve(), "could not establish session")
	}
	return client, session, nil
}
