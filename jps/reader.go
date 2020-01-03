package jps

import (
	"bufio"
	"io"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe"
)

//UnderlyingReader is the interface required by Reader.
type UnderlyingReader interface {
	Session() Session
	BodyLength() (int64, error)
	Body() (io.ReadCloser, error)
}

//Reader is a standard reader type that provides optional buffering and piping (to an UnderlyingReader).
type Reader struct {
	//Pipes is a slice of pointers to pipe generators, which, when the body is read, the data will be piped through after all the Transfer-Encoding and Content-Encoding pipes.
	// Any changes made to this slice after GetBody has been called do not have an effect.
	Pipes []*pipe.ReaderPipeGenerator

	request UnderlyingReader

	rawBody    io.ReadCloser
	pipedBody  io.ReadCloser
	buf        *bufio.Reader
	bufferBody bool
	closed     bool
}

//NewReader returns a new, unbuffered and unpiped, Reader that reads from the given UnderlyingReader.
func NewReader(r UnderlyingReader) Reader {
	return Reader{
		request: r,
	}
}

//IsBuffered returns true if the Reader has a buffer, or false if not.
func (r *Reader) IsBuffered() bool {
	return r.bufferBody
}

//SetBuffered enables or disables the buffer for this Reader.
//The return value will be true if, after this operation, the reader will be buffered, or false otherwise.
//SetBuffered(true) will enable the buffer and always return true.
//SetBuffered(false) will only disable the buffer if it has not already been used.
func (r *Reader) SetBuffered(buffer bool) bool {
	if r.buf != nil {
		return true
	}
	r.bufferBody = buffer
	return buffer
}

//Session returns the Session asociated with this Reader.
func (r *Reader) Session() Session {
	return r.request.Session()
}

//Length returns the length of the body - ie the total amount of data that will be readable.
//Length returns a negative value if this is unknown.
func (r *Reader) Length() (int64, error) {
	return r.request.BodyLength()
}

func (r *Reader) getRawBody() (io.ReadCloser, error) {
	var err error
	if r.rawBody == nil {
		if r.closed {
			panic("Reader: body closed")
		}
		r.rawBody, err = r.request.Body()
	}
	return r.rawBody, err
}

func (r *Reader) getPipedBody() (io.ReadCloser, error) {
	var err error
	if r.pipedBody == nil {
		var raw io.ReadCloser
		raw, err = r.getRawBody()
		if err == nil {
			r.pipedBody, err = pipe.From(raw, r.Pipes)
		}
	}
	return r.pipedBody, err
}

func (r *Reader) getBufioReader() (*bufio.Reader, error) {
	if r.buf != nil {
		piped, err := r.getPipedBody()
		if err != nil {
			return nil, err
		}
		r.buf = r.Session().BufioSource().NewReader(piped)
	}
	return r.buf, nil
}

func (r *Reader) getReader() (io.Reader, error) {
	if r.IsBuffered() {
		r.getBufioReader()
	}
	return r.getPipedBody()
}

func (r *Reader) Read(dst []byte) (int, error) {
	reader, err := r.getReader()
	if err != nil {
		return 0, err
	}
	return reader.Read(dst)
}

//GetBufioReader returns the *bufio.Reader which reads from the underlying reader and is used for the buffering of the Body.
//If the Body is not set to be buffered already, then it enables buffering and returns the buffer.
func (r *Reader) GetBufioReader() (*bufio.Reader, error) {
	if r.buf == nil {
		r.bufferBody = true
		return r.getBufioReader()
	}
	return r.buf, nil
}

//Close closes the pipes and then the UnderlyingReader.
func (r *Reader) Close() error {
	//Recycle the buffer if it has been created
	if r.buf != nil {
		r.Session().BufioSource().RecycleReader(r.buf)
		r.buf = nil
	}
	var err error
	if r.pipedBody != nil {
		err = r.pipedBody.Close()
		r.pipedBody = nil
		r.rawBody = nil
	} else if r.rawBody != nil {
		err = r.rawBody.Close()
		r.rawBody = nil
	}
	return err
}
