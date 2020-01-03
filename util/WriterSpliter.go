package util

import "io"

type WriterSpliter []io.Writer

func (ws WriterSpliter) Write(src []byte) (int, error) {
	if len(ws) < 1 {
		return len(src), nil
	}
	n, err := ws[0].Write(src)
	src = src[:n]
	for i := 1; i < len(ws); i++ {
		_, _ = ws[i].Write(src)
	}
	return n, err
}
