package options

import (
	"fmt"
	"runtime"
	"time"

	"github.com/evergreen-ci/bond"
	"github.com/mongodb/jasper/options"
)

// YesCreateOpts creates the options to run the "yes" command for the given
// duration.
func YesCreateOpts(timeout time.Duration) *options.Create {
	return &options.Create{Args: []string{"yes"}, Timeout: timeout}
}

// TrueCreateOpts creates the options to run the "true" command.
func TrueCreateOpts() *options.Create {
	return &options.Create{
		Args: []string{"true"},
	}
}

// FalseCreateOpts creates the options to run the "false" command.
func FalseCreateOpts() *options.Create {
	return &options.Create{
		Args: []string{"false"},
	}
}

// SleepCreateOpts creates the options to run the "sleep" command for the give
// nnumber of seconds.
func SleepCreateOpts(num int) *options.Create {
	return &options.Create{
		Args: []string{"sleep", fmt.Sprint(num)},
	}
}

// ValidMongoDBDownloadOptions returns valid options for downloading a MongoDB
// archive file.
func ValidMongoDBDownloadOptions() options.MongoDBDownload {
	target := runtime.GOOS
	if target == "darwin" {
		target = "osx"
	}

	edition := "enterprise"
	if target == "linux" {
		edition = "base"
	}

	return options.MongoDBDownload{
		BuildOpts: bond.BuildOptions{
			Target:  target,
			Arch:    bond.MongoDBArch("x86_64"),
			Edition: bond.MongoDBEdition(edition),
			Debug:   false,
		},
		Releases: []string{"4.0-current"},
	}
}

// ModifyOpts is a function that returns arbitrarily-modified process creation
// options for tests.
type ModifyOpts func(*options.Create) *options.Create
