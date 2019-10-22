package options

import (
	"fmt"

	"github.com/pkg/errors"
)

// Remote represents options to SSH into a remote machine.
type Remote struct {
	Host string
	User string
	Args []string

	// kim: TODO: new fields
	Port int

	// Authentication
	PrivKeyFile       string
	PrivKeyPassphrase string
	Password          string
}

// Validate checks that the host is set so that the remote host can be
// identified.
func (opts *Remote) Validate() error {
	if opts.Host == "" {
		return errors.New("host cannot be empty")
	}
	return nil
}

func (opts *Remote) String() string {
	if opts.User == "" {
		return opts.Host
	}

	return fmt.Sprintf("%s@%s", opts.User, opts.Host)
}

// func (opts *Remote) signer() (ssh.Signer, error) {
//     if opts.PrivKeyFile != "" {
//         key, err := ioutil.ReadFile(opts.PrivKeyFile)
//         if err != nil {
//             return nil, errors.Wrap(err, "problem reading private key file")
//         }
//
//         var signer ssh.Signer
//         if opts.PrivKeyPassphrase != "" {
//             signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(opts.PrivKeyPassphrase))
//         } else {
//             signer, err = ssh.ParsePrivateKey(key)
//         }
//         if err != nil {
//             return nil, errors.Wrap(err, "could not parse private key")
//         }
//         return signer, nil
//     }
//
//     return nil, errors.New("no signature provided")
// }
//
// func (opts *Remote) client() (ssh.Client, error) {
//     // conn, newChannels, reqs, err := ssh.NewClientConn(nil, fmt.Sprintf("%s:%d", opts.Host, opts.Port, nil))
//     return ssh.Client{}, errors.New("TODO: implement")
// }
