package jasper

import (
	"testing"

	"github.com/surullabs/lint/gometalinter"
)

func TestLint(t *testing.T) {
	// Run default linters
	metalinter := gometalinter.Check{
		Args: []string{
			// Arguments to gometalinter. Do not include the package names here.
		},
	}
	if err := metalinter.Check("./..."); err != nil {
		t.Fatalf("lint failures: %v", err)
	}
}
