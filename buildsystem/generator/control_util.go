package generator

import (
	"github.com/pkg/errors"
	"github.com/yargevad/filepathx"
)

func withGlobMatches(patterns []string, op func(file string) error) error {
	for _, pattern := range patterns {
		matches, err := filepathx.Glob(pattern)
		if err != nil {
			return errors.Wrapf(err, "evaluating glob pattern '%s'", pattern)
		}

		for _, match := range matches {
			if err := op(match); err != nil {
				return errors.Wrapf(err, "file '%s'", match)
			}
		}
	}

	return nil
}
