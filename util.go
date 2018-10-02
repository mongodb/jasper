package jasper

import (
	"os"
	"path/filepath"
)

func sliceContains(group []string, name string) bool {
	for _, g := range group {
		if name == g {
			return true
		}
	}

	return false
}

// WriteFile writes the buffer to the file.
func WriteFile(buf []byte, path string) error {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if _, err = file.Write(buf); err != nil {
		return err
	}
	return nil
}
