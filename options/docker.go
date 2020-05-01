package options

import (
	"fmt"
	"net"
	"runtime"

	"github.com/docker/docker/client"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// Docker encapsulates options related to connecting to a Docker daemon.
type Docker struct {
	Host       string
	Port       uint
	APIVersion string
	Image      string
	// Platform refers to the major operating system on which the Docker
	// container runs.
	Platform string
}

// Validate checks whether all the required fields are set and sets defaults if
// none are specified.
func (d *Docker) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(d.Image == "", "Docker image must be specified")
	if d.Platform == "" {
		if utility.StringSliceContains(DockerPlatforms(), runtime.GOOS) {
			d.Platform = runtime.GOOS
		} else {
			catcher.Errorf("cannot set default platform to current runtime platform '%s' because it is unsupported", d.Platform)
		}
	} else if !utility.StringSliceContains(DockerPlatforms(), d.Platform) {
		catcher.Errorf("unrecognized platform '%s'", d.Platform)
	}
	return catcher.Resolve()
}

// DockerPlatforms returns all supported platforms that can run Docker
// processes.
func DockerPlatforms() []string {
	return []string{"windows", "darwin", "linux"}
}

// Resolve converts the Docker options into options to initialize a Docker
// client.
func (d *Docker) Resolve() ([]client.Opt, error) {
	var opts []client.Opt

	if d.Host != "" && d.Port != 0 {
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", d.Host, d.Port))
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve Docker daemon address %s:%d", d.Host, d.Port)
		}
		opts = append(opts, client.WithHost(addr.String()))
	}

	if d.APIVersion != "" {
		opts = append(opts, client.WithAPIVersionNegotiation())
	} else {
		opts = append(opts, client.WithVersion(d.APIVersion))
	}

	return opts, nil
}
