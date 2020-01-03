package util

import "io"

type NoopFlusher struct {
	io.WriteCloser
}

func (_ NoopFlusher) Flush() error {
	return nil
}
