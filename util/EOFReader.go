package util

import "io"

//EOFReader is a type, of which the Read method always returns (0, io.EOF)
type EOFReader struct{}

//EOFReadCloser is a type, of which the Read method always returns (0, io.EOF), and the Close method always returns nil.
type EOFReadCloser struct{}

func (_ EOFReader) Read(dst []byte) (int, error)     { return 0, io.EOF }
func (_ EOFReadCloser) Read(dst []byte) (int, error) { return 0, io.EOF }
func (_ EOFReadCloser) Close() error                 { return nil }
