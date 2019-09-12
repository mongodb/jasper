package testutil

import (
	"fmt"
	"time"

	"github.com/mongodb/jasper/options"
)

func YesCreateOpts(timeout time.Duration) options.Create {
	return options.Create{Args: []string{"yes"}, Timeout: timeout}
}

func TrueCreateOpts() *options.Create {
	return &options.Create{
		Args: []string{"true"},
	}
}

func FalseCreateOpts() *options.Create {
	return &options.Create{
		Args: []string{"false"},
	}
}

func SleepCreateOpts(num int) *options.Create {
	return &options.Create{
		Args: []string{"sleep", fmt.Sprint(num)},
	}
}
