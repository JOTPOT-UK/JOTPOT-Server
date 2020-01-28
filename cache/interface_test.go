package cache

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func readAll(r io.Reader, bufSize int) ([]byte, error) {
	out := make([]byte, 0, bufSize)
	buf := make([]byte, bufSize)
	for {
		n, err := r.Read(buf)
		out = append(out, buf[:n]...)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

var testErr = errors.New("test error")

type TestCloser struct {
	io.Reader
}

func (_ TestCloser) Close() error {
	return testErr
}

func getReader(n int64) (io.ReadCloser, []byte) {
	data := []byte("Hello, world!\r\nHmm...\nThis is interesting!\n")
	r := NewLimitReadCloser(TestCloser{bytes.NewReader(data)}, n)
	return r, data[:n]
}

func TestLimitReadCloser1(t *testing.T) {
	r, expected := getReader(5)
	got, err := readAll(r, 1)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(got, expected) {
		t.Error("LimitReadCloser did not read the correct data.")
	} else if r.Close() != testErr {
		t.Error("LimitReadCloser.Close() did not call the underlying close method!")
	}
}

func TestLimitReadCloser2(t *testing.T) {
	r, expected := getReader(5)
	got, err := readAll(r, 2)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(got, expected) {
		t.Error("LimitReadCloser did not read the correct data.")
	} else if r.Close() != testErr {
		t.Error("LimitReadCloser.Close() did not call the underlying close method!")
	}
}

func TestLimitReadCloser5(t *testing.T) {
	r, expected := getReader(5)
	got, err := readAll(r, 5)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(got, expected) {
		t.Error("LimitReadCloser did not read the correct data.")
	} else if r.Close() != testErr {
		t.Error("LimitReadCloser.Close() did not call the underlying close method!")
	}
}

func TestLimitReadCloser6(t *testing.T) {
	r, expected := getReader(5)
	got, err := readAll(r, 6)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(got, expected) {
		t.Error("LimitReadCloser did not read the correct data.")
	} else if r.Close() != testErr {
		t.Error("LimitReadCloser.Close() did not call the underlying close method!")
	}
}

func TestLimitReadCloser10(t *testing.T) {
	r, expected := getReader(5)
	got, err := readAll(r, 10)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(got, expected) {
		t.Error("LimitReadCloser did not read the correct data.")
	} else if r.Close() != testErr {
		t.Error("LimitReadCloser.Close() did not call the underlying close method!")
	}
}
