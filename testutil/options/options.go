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
	return options.MongoDBDownload{
		BuildOpts: ValidMongoDBBuildOptions(),
		Releases:  []string{"7.0-current"},
	}
}

// ValidMongoDBBuildOptions returns valid options for a MongoDB build.
func ValidMongoDBBuildOptions() bond.BuildOptions {
	var edition string
	var target string
	platform := runtime.GOOS
	switch platform {
	case "darwin":
		edition = "enterprise"
		target = "macos"
	case "linux":
		edition = "targeted"
		target = "ubuntu2204"
	default:
		edition = "enterprise"
		target = platform
	}
	arch := "x86_64"

	return bond.BuildOptions{
		Target:  target,
		Arch:    bond.MongoDBArch(arch),
		Edition: bond.MongoDBEdition(edition),
		Debug:   false,
	}
}

// ModifyOpts is a function that returns arbitrarily-modified process creation
// options for tests.
type ModifyOpts func(*options.Create) *options.Create
