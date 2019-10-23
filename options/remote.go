package options

import (
	"fmt"
	"io/ioutil"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Remote represents options to SSH into a remote machine.
type Remote struct {
	Host string
	User string
	// TODO: will be ignored, remove
	Args []string

	Port int

	// Authentication
	PrivKeyFile       string
	PrivKeyPassphrase string
	Password          string
}

const defaultSSHPort = 22

// Validate checks that the host is set so that the remote host can be
// identified.
func (opts *Remote) Validate() error {
	if opts.Host == "" {
		return errors.New("host cannot be empty")
	}
	if opts.Port == 0 {
		opts.Port = defaultSSHPort
	}
	if opts.PrivKeyFile == "" && opts.Password == "" {
		return errors.New("must specify an authentication method")
	}
	if opts.PrivKeyFile != "" && opts.Password != "" {
		return errors.New("cannot specify more than one authentication method")
	}
	if opts.PrivKeyFile == "" && opts.PrivKeyPassphrase != "" {
		return errors.New("cannot set passphrase without private key file")
	}
	return nil
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
	if opts.PrivKeyFile != "" {
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
	key, err := ioutil.ReadFile(opts.PrivKeyFile)
	if err != nil {
		return nil, err
	}

	var signer ssh.Signer
	if opts.PrivKeyPassphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(opts.PrivKeyPassphrase))
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
