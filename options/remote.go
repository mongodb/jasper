package options

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type config struct {
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
// kim: TODO: support proxy config
type Remote struct {
	config
	// kim: TODO: this should be renamed
	Proxy Proxy
}

type Proxy struct {
	config
}

const defaultSSHPort = 22

// Validate ensures that enough information is provided to connect to a remote
// host.
func (opts *Remote) Validate() error {
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

func (opts *Remote) String() string {
	if opts.User == "" {
		return opts.Host
	}

	return fmt.Sprintf("%s@%s", opts.User, opts.Host)
}

func (opts *Remote) Resolve() (*ssh.Client, *ssh.Session, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, errors.Wrap(err, "invalid remote options")
	}

	var auth []ssh.AuthMethod
	if opts.KeyFile != "" {
		pubkey, err := opts.publicKey()
		if err != nil {
			return nil, nil, errors.Wrap(err, "problem getting public key from file")
		}
		auth = append(auth, pubkey)
	}
	if opts.Password != "" {
		auth = append(auth, ssh.Password(opts.Password))
	}
	config := &ssh.ClientConfig{
		Timeout:         opts.Timeout,
		User:            opts.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.Port), config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "problem dialing host")
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

func (opts *Remote) publicKey() (ssh.AuthMethod, error) {
	key, err := ioutil.ReadFile(opts.KeyFile)
	if err != nil {
		return nil, err
	}

	var signer ssh.Signer
	if opts.KeyPassphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(opts.KeyPassphrase))
		if err != nil {
			return nil, err
		}
	} else {
		signer, err = ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
	}
	return ssh.PublicKeys(signer), nil
}
