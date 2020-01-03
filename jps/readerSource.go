package jps

import (
	"bufio"
	"io"
	"sync"
)

//BufioReaderSource provides an interface for recieving new and recycling old bufio readers.
type BufioReaderSource interface {
	NewReader(io.Reader) *bufio.Reader
	RecycleReader(*bufio.Reader)
}

//BufioWriterSource provides an interface for recieving new and recycling old bufio writers.
type BufioWriterSource interface {
	NewWriter(io.Writer) *bufio.Writer
	RecycleWriter(*bufio.Writer)
}

//BufioSource provides an interface for recieving new and recycling old bufio readers and writers.
type BufioSource interface {
	BufioReaderSource
	BufioWriterSource
}

//BasicBufioSource is a simple implementation of BufioSource.
//It does not recycle readers or writers, instead, it creates new ones and recycle is a noop.
//This is useful for testing or in situations where buffers are not used frequently.
type BasicBufioSource struct {
	ReaderBufSize, WriterBufSize int
}

//NewReader returns a new bufio.Reader with the size specified by the ReaderBufSize field.
func (s *BasicBufioSource) NewReader(r io.Reader) *bufio.Reader {
	return bufio.NewReaderSize(r, s.ReaderBufSize)
}

//NewWriter returns a new bufio.Writer with the size specified by the WriterBufSize field.
func (s *BasicBufioSource) NewWriter(r io.Writer) *bufio.Writer {
	return bufio.NewWriterSize(r, s.WriterBufSize)
}

//RecycleReader is a noop to implement the BufioReaderSource interface.
func (s *BasicBufioSource) RecycleReader(_ *bufio.Reader) {}

//RecycleWriter is a noop to implement the BufioWriterSource interface.
func (s *BasicBufioSource) RecycleWriter(_ *bufio.Writer) {}

type CachedBufioReaderSource struct {
	Slice   []*bufio.Reader
	BufSize int
	Lock    sync.Mutex
}

func NewCachedBufioReaderSource(cacheSize, bufSize int) BufioReaderSource {
	return &CachedBufioReaderSource{
		Slice:   make([]*bufio.Reader, 0, cacheSize),
		BufSize: bufSize,
	}
}

func NewFilledCachedBufioReaderSource(cacheSize, bufSize int) BufioReaderSource {
	rv := &CachedBufioReaderSource{
		Slice:   make([]*bufio.Reader, 0, cacheSize),
		BufSize: bufSize,
	}
	rv.Fill()
	return rv
}

func (cbrs *CachedBufioReaderSource) Fill() {
	cbrs.Lock.Lock()
	defer cbrs.Lock.Unlock()
	for len(cbrs.Slice) < cap(cbrs.Slice) {
		cbrs.Slice = append(cbrs.Slice, bufio.NewReaderSize(nil, cbrs.BufSize))
	}
}

func (cbrs *CachedBufioReaderSource) NewReader(r io.Reader) *bufio.Reader {
	cbrs.Lock.Lock()
	i := len(cbrs.Slice)
	if i == 0 {
		cbrs.Lock.Unlock()
		return bufio.NewReaderSize(r, cbrs.BufSize)
	}
	i--
	rv := cbrs.Slice[i]
	cbrs.Slice = cbrs.Slice[:i]
	cbrs.Lock.Unlock()
	rv.Reset(r)
	return rv
}

func (cbrs *CachedBufioReaderSource) RecycleReader(br *bufio.Reader) {
	cbrs.Lock.Lock()
	defer cbrs.Lock.Unlock()
	if len(cbrs.Slice) < cap(cbrs.Slice) && br.Size() == cbrs.BufSize {
		cbrs.Slice = append(cbrs.Slice, br)
	}
}

type CachedBufioWriterSource struct {
	Slice   []*bufio.Writer
	BufSize int
	Lock    sync.Mutex
}

func NewCachedBufioWriterSource(cacheSize, bufSize int) BufioWriterSource {
	return &CachedBufioWriterSource{
		Slice:   make([]*bufio.Writer, 0, cacheSize),
		BufSize: bufSize,
	}
}

func NewFilledCachedBufioWriterSource(cacheSize, bufSize int) BufioWriterSource {
	rv := &CachedBufioWriterSource{
		Slice:   make([]*bufio.Writer, 0, cacheSize),
		BufSize: bufSize,
	}
	rv.Fill()
	return rv
}

func (cbws *CachedBufioWriterSource) Fill() {
	cbws.Lock.Lock()
	defer cbws.Lock.Unlock()
	for len(cbws.Slice) < cap(cbws.Slice) {
		cbws.Slice = append(cbws.Slice, bufio.NewWriterSize(nil, cbws.BufSize))
	}
}

func (cbws *CachedBufioWriterSource) NewWriter(w io.Writer) *bufio.Writer {
	cbws.Lock.Lock()
	i := len(cbws.Slice)
	if i == 0 {
		cbws.Lock.Unlock()
		return bufio.NewWriterSize(w, cbws.BufSize)
	}
	i--
	rv := cbws.Slice[i]
	cbws.Slice = cbws.Slice[:i]
	cbws.Lock.Unlock()
	rv.Reset(w)
	return rv
}

func (cbws *CachedBufioWriterSource) RecycleWriter(bw *bufio.Writer) {
	cbws.Lock.Lock()
	defer cbws.Lock.Unlock()
	if len(cbws.Slice) < cap(cbws.Slice) && bw.Size() == cbws.BufSize {
		cbws.Slice = append(cbws.Slice, bw)
	}
}

type CachedBufioSource struct {
	CachedBufioReaderSource
	CachedBufioWriterSource
}

func (cbs *CachedBufioSource) Fill() {
	cbs.CachedBufioReaderSource.Fill()
	cbs.CachedBufioWriterSource.Fill()
}

func NewCachedBufioSource(rCacheSize, wCacheSize, rBufSize, wBufSize int) BufioSource {
	return &CachedBufioSource{
		CachedBufioReaderSource{
			Slice:   make([]*bufio.Reader, 0, rCacheSize),
			BufSize: rBufSize,
		},
		CachedBufioWriterSource{
			Slice:   make([]*bufio.Writer, 0, wCacheSize),
			BufSize: wBufSize,
		},
	}
}

func NewFilledCachedBufioSource(rCacheSize, wCacheSize, rBufSize, wBufSize int) BufioSource {
	rv := &CachedBufioSource{
		CachedBufioReaderSource{
			Slice:   make([]*bufio.Reader, 0, rCacheSize),
			BufSize: rBufSize,
		},
		CachedBufioWriterSource{
			Slice:   make([]*bufio.Writer, 0, wCacheSize),
			BufSize: wBufSize,
		},
	}
	rv.Fill()
	return rv
}
