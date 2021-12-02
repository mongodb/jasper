package options

import (
	"fmt"
	"net"
	"runtime"

	"github.com/docker/docker/client"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// Docker encapsulates options related to connecting to a Docker daemon.
type Docker struct {
	Host       string `bson:"host,omitempty" json:"host,omitempty" yaml:"host,omitempty"`
	Port       int    `bson:"port,omitempty" json:"port,omitempty" yaml:"port,omitempty"`
	APIVersion string `bson:"api_version,omitempty" json:"api_version,omitempty" yaml:"api_version,omitempty"`
	Image      string `bson:"image,omitempty" json:"image,omitempty" yaml:"image,omitempty"`
	// OS refers to the major operating system on which the Docker container
	// runs.
	OS string `bson:"os,omitempty" json:"os,omitempty" yaml:"os,omitempty"`
	// Arch is the CPU architecture of the machine on which the Docker container
	// runs.
	Arch string `bson:"arch,omitempty" json:"arch,omitempty" yaml:"arch,omitempty"`
}

// Validate checks whether all the required fields are set and sets defaults if
// none are specified.
func (opts *Docker) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(opts.Port < 0, "port must be positive value")
	catcher.NewWhen(opts.Image == "", "Docker image must be specified")
	if opts.OS == "" {
		if OSSupportsDocker(runtime.GOOS) {
			opts.OS = runtime.GOOS
		} else {
			catcher.Errorf("cannot set default platform to current runtime platform '%s' because it is unsupported", opts.OS)
		}
	} else if !OSSupportsDocker(opts.OS) {
		catcher.Errorf("unrecognized platform '%s'", opts.OS)
	}
	if opts.Arch == "" {
		opts.Arch = runtime.GOARCH
	}
	return catcher.Resolve()
}

// Copy returns a copy of the options for only the exported fields.
func (opts *Docker) Copy() *Docker {
	optsCopy := *opts
	return &optsCopy
}

// OSSupportsDocker returns whether or not the operating system is supported by
// Docker.
func OSSupportsDocker(platform string) bool {
	switch platform {
	case "darwin", "linux", "windows":
		return true
	default:
		return false
	}
}

// Resolve converts the Docker options into options to initialize a Docker
// client.
func (opts *Docker) Resolve() (*client.Client, error) {
	var clientOpts []client.Opt

	if opts.Host != "" && opts.Port > 0 {
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.Port))
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve Docker daemon address %s:%d", opts.Host, opts.Port)
		}
		clientOpts = append(clientOpts, client.WithHost(addr.String()))
	}

	if opts.APIVersion != "" {
		clientOpts = append(clientOpts, client.WithAPIVersionNegotiation())
	} else {
		clientOpts = append(clientOpts, client.WithVersion(opts.APIVersion))
	}

	client, err := client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "could not create Docker client")
	}
	return client, nil
}
