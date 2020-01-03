package jps

import (
	"bufio"
	"io"
)

type Stream interface {
	io.ReadWriteCloser
	ConnectionDetails

	Session() StreamSession
}

//WrappedStream can be used for returning a stream with the Read, Write and Close methods overwritten.
// (To not overwrite a specific method, set the iterface for that method to the stream.)
//The intended use of this is to provide a standard way of allowing reuse of bufio/other buffered IO types.
//For example, if you handle a stream, and create a bufio.Reader for it, if Hyjack is called on your session, you need to return a stream which reads from the bufio.Reader.
//This should be used, as it provides a standard way for someone using your Hyjack call to get and then reuse your bufio.Reader rather than buffering it twice.
//The Hyjack call could have returned an optional pointer to a bufio.Reader and a bufio.Writer, however that would then not allow the use of other buffered IO types.
//Please see GetBufioReader, BufioReader, GetBufioWriter and BufioWriter for standard functions to get the underlying buffered IO.
type WrappedStream struct {
	Reader io.Reader
	Writer io.Writer
	Closer io.Closer
	Stream
}

//Read is an alias for ws.Reader.Read
func (ws *WrappedStream) Read(p []byte) (int, error) {
	return ws.Reader.Read(p)
}

//Write is an alias for ws.Reader.Write
func (ws *WrappedStream) Write(p []byte) (int, error) {
	return ws.Writer.Write(p)
}

//Close is an alias for ws.Closer.Close
func (ws *WrappedStream) Close() error {
	return ws.Closer.Close()
}

//GetBufioReader checks if the given Stream is a WrappedStream and if its Reader is a *bufio.Reader;
// if it is, then it returns it; otherwise, it returns nil.
//Also see BufioReader, which always gives you a reader.
func GetBufioReader(s Stream) *bufio.Reader {
	if ws, ok := s.(*WrappedStream); ok {
		if br, ok := ws.Reader.(*bufio.Reader); ok {
			return br
		}
	}
	return nil
}

//GetBufioWriter checks if the given Stream is a WrappedStream and if its Writer is a *bufio.Writer;
// if it is, then it returns it; otherwise, it returns nil.
//Also see BufioWriter, which always gives you a writer.
func GetBufioWriter(s Stream) *bufio.Writer {
	if ws, ok := s.(*WrappedStream); ok {
		if bw, ok := ws.Writer.(*bufio.Writer); ok {
			return bw
		}
	}
	return nil
}

//BufioReader returns a *bufio.Reader which reads from the given stream.
//If the Stream is a WrappedStream and its Reader is in fact a bufio.Reader, it returns that Reader rather than creating a new one.
//If a new reader is created, is uses s.Session.BufioSource().NewReader
func BufioReader(s Stream) *bufio.Reader {
	r := GetBufioReader(s)
	if r == nil {
		r = s.Session().BufioSource().NewReader(s)
	}
	return r
}

//BufioWriter returns a *bufio.Writer which writes to the given stream.
//If the Stream is a WrappedStream and its Writer is in fact a bufio.Writer, it returns that Writer rather than creating a new one.
//If a new writer is created, is uses s.Session.BufioSource().NewWriter
func BufioWriter(s Stream) *bufio.Writer {
	w := GetBufioWriter(s)
	if w == nil {
		w = s.Session().BufioSource().NewWriter(s)
	}
	return w
}
