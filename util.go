package jasper

import (
	"os"

	"github.com/pkg/errors"
)

func sliceContains(group []string, name string) bool {
	for _, g := range group {
		if name == g {
			return true
		}
	}

	return false
}

// MakeEnclosingDirectories recursively makes directories (if necessary) for the given path.
func MakeEnclosingDirectories(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModeDir|os.ModePerm); err != nil {
			return err
		}
	} else if !info.IsDir() {
		return errors.Errorf("'%s' already exists and is not a directory", path)
	}
	return nil
}
