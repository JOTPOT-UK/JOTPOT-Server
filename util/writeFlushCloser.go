package util

import "io"

type Flusher interface {
	Flush() error
}

type WriteFlusher interface {
	io.Writer
	Flusher
}

type WriteFlushCloser interface {
	io.WriteCloser
	Flusher
}
