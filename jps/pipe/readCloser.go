package pipe

import "io"

type ReadCloserReader struct {
	R io.Reader
}

func (r *ReadCloserReader) Read(dst []byte) (int, error) { return r.R.Read(dst) }
func (r *ReadCloserReader) Close() error                 { return nil }
