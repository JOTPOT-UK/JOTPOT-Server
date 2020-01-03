package util

import "io"

type ReaderClosePassOn struct {
	io.Reader
	io.Closer
}

type WriteFlusherClosePassOn struct {
	WriteFlusher
	io.Closer
}
