package jasper

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

var httpClientPool *sync.Pool

func init() {
	httpClientPool = &sync.Pool{
		New: func() interface{} {
			return &http.Client{}
		},
	}
}

// GetHTTPClient gets an HTTP client from the client pool.
func GetHTTPClient() *http.Client {
	return httpClientPool.Get().(*http.Client)
}

// PutHTTPClient returns the given HTTP client back to the pool.
func PutHTTPClient(client *http.Client) {
	httpClientPool.Put(client)
}

// WriteFileInfo represents the information necessary to write to a file.
type WriteFileInfo struct {
	Path string `json:"path"`
	// File content can come from either Content or Reader, but not both.
	// TODO: rename
	Content []byte      `json:"content"`
	Reader  io.Reader   `json:"-"`
	Append  bool        `json:"append"`
	Perm    os.FileMode `json:"perm"`
}

// validateContent ensures that there is at most one source of content for
// the file.
func (info *WriteFileInfo) validateContent() error {
	if len(info.Content) > 0 && info.Reader != nil {
		return errors.New("cannot have both data and reader set as file content")
	}
	// If neither is set, ensure that Content is empty rather than nil to
	// prevent potential writes with a nil slice.
	if len(info.Content) == 0 && info.Reader == nil {
		info.Content = []byte{}
	}
	return nil
}

// Validate ensures that all the parameters to write to a file are valid and sets
// default permissions if necessary.
func (info *WriteFileInfo) Validate() error {
	catcher := grip.NewBasicCatcher()
	if info.Path == "" {
		catcher.New("path to file must be specified")
	}

	if info.Perm == 0 {
		info.Perm = 0666
	}

	catcher.Add(info.validateContent())

	return catcher.Resolve()
}

// DoWrite writes the data to the given path, creating the directory hierarchy as
// needed and the file if it does not exist yet.
func (info *WriteFileInfo) DoWrite() error {
	if err := makeEnclosingDirectories(filepath.Dir(info.Path)); err != nil {
		return errors.Wrap(err, "problem making enclosing directories")
	}

	openFlags := os.O_RDWR | os.O_CREATE
	if info.Append {
		openFlags |= os.O_APPEND
	} else {
		openFlags |= os.O_TRUNC
	}

	file, err := os.OpenFile(info.Path, openFlags, 0666)
	if err != nil {
		return errors.Wrapf(err, "error opening file %s", info.Path)
	}

	catcher := grip.NewBasicCatcher()

	reader, err := info.ContentReader()
	if err != nil {
		catcher.Wrap(file.Close(), "error closing file")
		catcher.Wrap(err, "error getting file content as bytes")
		return catcher.Resolve()
	}

	bufReader := bufio.NewReader(reader)
	if _, err = io.Copy(file, bufReader); err != nil {
		catcher.Wrap(file.Close(), "error closing file")
		catcher.Wrap(err, "error writing content to file")
		return catcher.Resolve()
	}

	return errors.Wrap(file.Close(), "error closing file")
}

// SetPerm sets the file permissions on the file. This should be called after
// DoWrite. If no file exists at (WriteFileInfo).Path, it will error.
func (info *WriteFileInfo) SetPerm() error {
	return errors.Wrap(os.Chmod(info.Path, info.Perm), "error setting permissions")
}

// ContentBytes returns the contents to be written to the file as a byte slice.
func (info *WriteFileInfo) ContentBytes() ([]byte, error) {
	if err := info.validateContent(); err != nil {
		return nil, errors.Wrap(err, "could not validate file content source")
	}

	if info.Reader != nil {
		content, err := ioutil.ReadAll(info.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "could not read from reader content source")
		}
		info.Content = content
		info.Reader = nil
	}

	return info.Content, nil
}

// ContentReader returns the contents to be written to the file as an io.Reader.
func (info *WriteFileInfo) ContentReader() (io.Reader, error) {
	if err := info.validateContent(); err != nil {
		return nil, errors.Wrap(err, "could not validate file content source")
	}

	if info.Reader != nil {
		return info.Reader, nil
	}

	info.Reader = bytes.NewBuffer(info.Content)
	info.Content = nil

	return info.Reader, nil
}

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
