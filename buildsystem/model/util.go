package model

import (
	"path/filepath"

	"github.com/evergreen-ci/utility"
	"github.com/pkg/errors"
)

// withMatchingFiles performs the given operation op on all files that match the
// gitignore-style globbing patterns within the working directory workDir.
func withMatchingFiles(workDir string, patterns []string, op func(file string) error) error {
	files, err := utility.BuildFileList(workDir, patterns...)
	if err != nil {
		return errors.Wrap(err, "evaluating file patterns")
	}
	for _, file := range files {
		if err := op(filepath.Join(workDir, file)); err != nil {
			return errors.Wrapf(err, "file '%s'", file)
		}
	}

	return nil
}
