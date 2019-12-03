package jasper

import "bytes"

type writeCloser struct {
	bytes.Buffer
}

func (w *writeCloser) Close() error { return nil }
