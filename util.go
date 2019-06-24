package jasper

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/mongodb/grip"
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

func makeEnclosingDirectories(path string) error {
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

func writeFile(reader io.Reader, path string) error {
	if err := makeEnclosingDirectories(filepath.Dir(path)); err != nil {
		return errors.Wrap(err, "problem making enclosing directories")
	}

	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "problem creating file")
	}

	catcher := grip.NewBasicCatcher()
	if _, err := io.Copy(file, reader); err != nil {
		catcher.Add(errors.Wrap(err, "problem writing file"))
	}

	catcher.Add(errors.Wrap(file.Close(), "problem closing file"))

	return catcher.Resolve()
}

// CappedWriter implements a buffer that stores up to MaxBytes bytes.
// Returns ErrBufferFull on overflowing writes.
type CappedWriter struct {
	Buffer   *bytes.Buffer
	MaxBytes int
}

// ErrBufferFull returns an error indicating that a CappedWriter's buffer has
// reached max capacity.
func ErrBufferFull() error {
	return errors.New("buffer is full")
}

// Write writes to the buffer. An error is returned if the buffer is full.
func (cw *CappedWriter) Write(in []byte) (int, error) {
	remaining := cw.MaxBytes - cw.Buffer.Len()
	if len(in) <= remaining {
		return cw.Buffer.Write(in)
	}
	// fill up the remaining buffer and return an error
	n, _ := cw.Buffer.Write(in[:remaining])
	return n, ErrBufferFull()
}

// IsFull indicates whether the buffer is full.
func (cw *CappedWriter) IsFull() bool {
	return cw.Buffer.Len() == cw.MaxBytes
}

// String return the contents of the buffer as a string.
func (cw *CappedWriter) String() string {
	return cw.Buffer.String()
}

func (cw *CappedWriter) Bytes() []byte {
	return cw.Buffer.Bytes()
}

// Close is a noop method so that you can use CappedWriter as an
// io.WriteCloser.
func (cw *CappedWriter) Close() error { return nil }
