package util

import "io"

type PrintyReader struct {
	Reader io.Reader
	CC     io.Writer
}

func (pr PrintyReader) Read(dst []byte) (int, error) {
	n, err := pr.Reader.Read(dst)
	pr.CC.Write(dst[:n])
	return n, err
}
