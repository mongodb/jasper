package testutil

import (
	"fmt"
	"runtime"
	"time"

	"github.com/mongodb/jasper/options"
	"github.com/tychoish/bond"
)

// YesCreateOpts creates the options to run the yes command.
func YesCreateOpts(timeout time.Duration) *options.Create {
	return &options.Create{Args: []string{"yes"}, Timeout: timeout}
}

// TrueCreateOpts creates the options to run the true command.
func TrueCreateOpts() *options.Create {
	return &options.Create{
		Args: []string{"true"},
	}
}

// FalseCreateOpts creates the options to run the false command.
func FalseCreateOpts() *options.Create {
	return &options.Create{
		Args: []string{"false"},
	}
}

// SleepCreateOpts creates the options to run the sleep command.
func SleepCreateOpts(num int) *options.Create {
	return &options.Create{
		Args: []string{"sleep", fmt.Sprint(num)},
	}
}

// ValidMongoDBDownloadOptions returns options for a valid MongoDB download.
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
