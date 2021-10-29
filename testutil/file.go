package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// AddFileToDirectory adds an archive file given by fileName with the given
// fileContents to the directory.
func AddFileToDirectory(dir, fileName, fileContents string) error {
	archived, err := tryAddFileToArchive(dir, fileName, fileContents)
	if err != nil {
		return errors.Wrap(err, "adding file to archive")
	}
	if archived {
		return nil
	}

	file, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		return err
	}
	if _, err := file.Write([]byte(fileContents)); err != nil {
		catcher := grip.NewBasicCatcher()
		catcher.Add(err)
		catcher.Add(file.Close())
		return catcher.Resolve()
	}
	return file.Close()
}

func tryAddFileToArchive(dir, fileName, fileContents string) (bool, error) {
	format, err := archiver.ByExtension(fileName)
	if err != nil {
		return false, nil
	}
	archiveFormat, ok := format.(archiver.Archiver)
	if !ok {
		return false, nil
	}

	tmpFile, err := ioutil.TempFile(dir, "tmp.txt")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(fileContents)); err != nil {
		catcher := grip.NewBasicCatcher()
		catcher.Add(err)
		catcher.Add(tmpFile.Close())
		return false, catcher.Resolve()
	}
	if err := tmpFile.Close(); err != nil {
		return false, err
	}

	if err := archiveFormat.Archive([]string{tmpFile.Name()}, filepath.Join(dir, fileName)); err != nil {
		return false, err
	}

	return true, nil
}

// BuildDirectory is the project-level directory where all build artifacts are
// put.
func BuildDirectory() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filepath.Dir(file)), "build")
}
