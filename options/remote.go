package options

import (
	"fmt"

	"github.com/mongodb/grip"
)

// RemoteConfig represents the arguments to connect to a remote host.
type RemoteConfig struct {
}

// Remote represents options to SSH into a remote machine.
type Remote struct {
	Host string `bson:"host" json:"host"`
	User string `bson:"user" json:"user"`

	// Additional args to the SSH binary.
	Args []string `bson:"args,omitempty" json:"args,omitempty"`
}

// Copy returns a copy of the options for only the exported fields.
func (opts *Remote) Copy() *Remote {
	optsCopy := *opts
	return &optsCopy
}

// Validate ensures that enough information is provided to connect to a remote
// host.
func (opts *Remote) Validate() error {
	catcher := grip.NewBasicCatcher()
	if opts.Host == "" {
		catcher.New("host cannot be empty")
	}

	return catcher.Resolve()
}

func (opts *Remote) String() string {
	if opts.User == "" {
		return opts.Host
	}

	return fmt.Sprintf("%s@%s", opts.User, opts.Host)
}
