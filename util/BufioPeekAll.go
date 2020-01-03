package util

import (
	"bufio"
	"io"
	"unsafe"
)

type bufioReader struct {
	buf          []byte
	rd           io.Reader // reader provided by the client
	r, w         int       // buf read and write positions
	err          error
	lastByte     int // last byte read for UnreadByte; -1 means invalid
	lastRuneSize int // size of last rune read for UnreadRune; -1 means invalid
}

func BufioPeekAll(rr *bufio.Reader) []byte {
	r := (*bufioReader)(unsafe.Pointer(rr))
	if r.w < r.r {
		return append(r.buf[r.r:], r.buf[:r.w]...)
	}
	return r.buf[r.r:r.w]
}
