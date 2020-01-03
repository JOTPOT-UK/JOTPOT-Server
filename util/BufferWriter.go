package util

import "io"

type BufferWriter struct {
	Buf []byte
}

func NewBufferWriter(cap int) BufferWriter {
	return BufferWriter{make([]byte, 0, cap)}
}

func (w *BufferWriter) Write(data []byte) (int, error) {
	w.Buf = append(w.Buf, data...)
	return len(data), nil
}

func (w *BufferWriter) WriteString(data string) (int, error) {
	w.Buf = append(w.Buf, data...)
	return len(data), nil
}

func (w *BufferWriter) WriteTo(dst io.Writer) error {
	n, err := dst.Write(w.Buf)
	if err != nil {
		return err
	}
	if n < len(w.Buf) {
		return io.ErrShortWrite
	}
	return nil
}
