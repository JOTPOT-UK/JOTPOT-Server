package jps

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe"
)

type UnderlyingWriter interface {
	Session() Session
	BodyLength() (int64, error)
	SetBodyLength(int64) error
	Body() (WriteFlushCloser, error)
}

type Writer struct {
	//Pipes is a slice of pointers to pipe generators, which, when the body is read, the data will be piped through after all the Transfer-Encoding and Content-Encoding pipes.
	// Any changes made to this slice after GetBody has been called do not have an effect.
	Pipes []*pipe.WriterPipeGenerator

	response UnderlyingWriter

	rawBody    WriteFlushCloser
	pipedBody  WriteFlushCloser
	buf        *bufio.Writer
	bufferBody bool
	closed     bool
}

func NewWriter(underlyingWriter UnderlyingWriter) Writer {
	return Writer{
		response: underlyingWriter,
	}
}

var ErrBodyAlreadyStarted = errors.New("body already started")

func (w *Writer) StringBody(s string) error {
	if w.rawBody != nil {
		return ErrBodyAlreadyStarted
	}
	err := w.SetLength(int64(len(s)))
	if err != nil {
		return err
	}
	n, err := w.Write([]byte(s))
	if err != nil {
		return err
	}
	if n != len(s) {
		return io.ErrShortWrite
	}
	return w.Close()
}

func (w *Writer) BytesBody(s []byte) error {
	//TIMER:start := time.Now()
	if w.rawBody != nil {
		return ErrBodyAlreadyStarted
	}
	err := w.SetLength(int64(len(s)))
	if err != nil {
		return err
	}
	//TIMER:BodySetupTimes = append(BodySetupTimes, time.Since(start))
	//TIMER:start = time.Now()
	n, err := w.Write(s)
	if err != nil {
		return err
	}
	//TIMER:BodyWriteTimes = append(BodyWriteTimes, time.Since(start))
	if n != len(s) {
		return io.ErrShortWrite
	}
	return w.Close()
}

func (r *Writer) Session() Session {
	return r.response.Session()
}

//SetLength sets the length of the body. You shouldn't ususally use this as setting metadata properties such as size and range should set this.
func (r *Writer) SetLength(l int64) error {
	return r.response.SetBodyLength(l)
}

//Length returns the total length of the body.
func (r *Writer) Length() (int64, error) {
	return r.response.BodyLength()
}

func (r *Writer) getRawBody() (WriteFlushCloser, error) {
	var err error
	if r.rawBody == nil {
		//Incoming represents an incoming request to a server
		if r.closed {
			panic("Writer: body closed")
		}
		r.rawBody, err = r.response.Body()
	}
	return r.rawBody, err
}

func (r *Writer) getPipedBody() (WriteFlushCloser, error) {
	var err error
	if r.pipedBody == nil {
		//TIMER:start := time.Now()
		var raw WriteFlushCloser
		raw, err = r.getRawBody()
		//TIMER:BodyWriteSetupRawBodyTimes = append(BodyWriteSetupRawBodyTimes, time.Since(start))
		//TIMER:start = time.Now()
		if err == nil {
			r.pipedBody, err = pipe.To(raw, r.Pipes)
		}
		//TIMER:BodyWriteSetupPipesTimes = append(BodyWriteSetupPipesTimes, time.Since(start))
	}
	return r.pipedBody, err
}

func (r *Writer) getBufioWriter() (*bufio.Writer, error) {
	if r.buf != nil {
		piped, err := r.getPipedBody()
		if err != nil {
			return nil, err
		}
		r.buf = r.Session().BufioSource().NewWriter(piped)
	}
	return r.buf, nil
}

func (r *Writer) getWriter() (io.Writer, error) {
	if r.IsBuffered() {
		return r.getBufioWriter()
	}
	return r.getPipedBody()
}

func (r *Writer) IsBuffered() bool {
	return r.bufferBody
}

func (r *Writer) SetBuffered(buffer bool) bool {
	if r.buf != nil {
		return true
	}
	r.bufferBody = buffer
	return buffer
}

func (r *Writer) Write(src []byte) (int, error) {
	//TIMER:start := time.Now()
	writer, err := r.getWriter()
	if err != nil {
		return 0, err
	}
	//TIMER:BodyWriteSetupTimes = append(BodyWriteSetupTimes, time.Since(start))
	//TIMER:start = time.Now()
	//TIMER:defer func() {
	//TIMER:BodyWriteWriteTimes = append(BodyWriteWriteTimes, time.Since(start))
	//TIMER:}()
	return writer.Write(src)
}

func (r *Writer) GetBufioWriter() (*bufio.Writer, error) {
	if r.buf == nil {
		r.bufferBody = true
		return r.getBufioWriter()
	}
	return r.buf, nil
}

func (r *Writer) Close() error {
	//Recycle the buffer if it has been created
	var err error
	if r.buf != nil {
		err = r.buf.Flush()
		r.Session().BufioSource().RecycleWriter(r.buf)
		r.buf = nil
	}
	_, err2 := r.getWriter()
	var err3 error
	var err4 error
	if r.pipedBody != nil {
		err3 = r.pipedBody.Close()
		r.pipedBody = nil
		r.rawBody = nil
	} else if r.rawBody != nil {
		err4 = r.rawBody.Close()
		r.rawBody = nil
	}
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	fmt.Println("Closed!")
	return err4
}

func (r *Writer) Flush() error {
	var err error
	if r.buf != nil {
		err = r.buf.Flush()
		if err != nil {
			return err
		}
	}
	if r.pipedBody != nil {
		err = r.pipedBody.Flush()
	} else if r.rawBody != nil {
		err = r.rawBody.Flush()
	}
	return err
}

/*
//IncomingRawReadCloser is to provide an interface from which the raw data can be read from the request
type IncomingRawReadCloser struct {
	r *Incoming
}

//Read calls Read on the underlying bufio.Reader of the request
func (r IncomingRawReadCloser) Read(buf []byte) (int, error) {
	return r.r.rawReader.Read(buf)
}

//Close calls Close on the underlying connection of the request
func (r IncomingRawReadCloser) Close() error {
	return r.r.con.Close()
}

//RawReadCloser returns a ReadCloser which reads from the raw data of the incoming request
// (implemented by the incomingRawReadCloser type)
func (r *Incoming) RawReadCloser() io.ReadCloser {
	if r.rawReader == nil {
		r.rawReader = r.Server.NewBufioReader(r.con)
	}
	return IncomingRawReadCloser{r}
}

//GotRawReader returns true if the rawReader of the request is
func (r *Incoming) GotRawReader() bool {
	return r.rawReader != nil
}

//GetRawReader returns a bufio.Reader that reads from the connection - it only creates one once, and the boolean result is false if it has used a prexisting one.
func (r *Incoming) GetRawReader() (*bufio.Reader, bool) {
	if r.rawReader == nil {
		r.rawReader = r.Server.NewBufioReader(r.con)
		return r.rawReader, true
	}
	return r.rawReader, false
}

//RawReader is equivilent to GetRawReader, except that the second return parameter is not given.
func (r *Incoming) RawReader() *bufio.Reader {
	if r.rawReader == nil {
		r.rawReader = r.Server.NewBufioReader(r.con)
	}
	return r.rawReader
}

//HasBody returns true if the request has a body
func (r *Incoming) HasBody() bool {
	return len(r.Request.Header.GetValuesRawKey("Transfer-Encoding")) != 0 || len(r.Request.Header.GetValuesRawKey("Content-Length")) != 0
}

//GotBody returns true if a body reader has already been created
func (r *Incoming) GotBody() bool {
	return r.body != nil
}

//GetBody the reader which reads the deocoded body. If a new reader is created (because it hasn't been created before), true is returned as the boolean result.
func (r *Incoming) GetBody() (io.ReadCloser, bool, error) {
	if r.body == nil {
		var err error
		var ok bool
		//This list of pipes should be used to decode the body, and should return EOF at the end
		var transferPipe []*pipe.ReaderPipeGenerator
		TEH := r.Request.Header.GetValuesRawKey("Transfer-Encoding")
		CLH := r.Request.Header.GetValuesRawKey("Content-Length")
		if len(CLH) == 1 {
			//If there is only 1 Content-Length header, there must not be any transfer-encodings
			if len(TEH) != 0 {
				return nil, false, jpserror.ErrContentLengthAndTransferEncoding
			}
			//Parse the length
			var CL uint64
			CL, err = strconv.ParseUint(CLH[0], 10, 63)
			if err != nil {
				return nil, false, jpserror.ErrMalformedContentLength
			}
			//The decoder is just a limiter
			limiter := limiter.GenerateLimiterGenerator(int64(CL))
			transferPipe = []*pipe.ReaderPipeGenerator{&limiter}
		} else if len(TEH) != 0 {
			//If we have a transfer-encoding, we can't also have a content-length
			if len(CLH) != 0 {
				return nil, false, jpserror.ErrContentLengthAndTransferEncoding
			}
			//Check that we are chunking last and only last
			if strings.ToLower(TEH[len(TEH)-1]) != "chunked" {
				return nil, false, jpserror.ErrMustChunkLast
			}
			for i := 0; i < len(TEH)-1; i++ {
				if TEH[i] == "chunked" {
					return nil, false, jpserror.ErrMustOnlyChunkOnce
				}
			}
			//Get the transfer pipes!
			transferPipe, ok = r.Server.TransferEncodings.GetReaderPipeGenerators(TEH)
			if !ok {
				return nil, false, jpserror.ErrUnsupportedTransferEncoding
			}
		} else if len(CLH) == 0 {
			//If we have no content length and no tranfer encoding, then there is no body
			r.body = util.EOFReadCloser{}
			return r.body, true, nil
		} else {
			//At this point, we must have more than 1 content-length header
			return nil, false, jpserror.ErrTooManyContentLength
		}
		//Get the pipes for the Content-Encoding
		CEcodes, ok := r.Server.ContentEncodings.GetReaderPipeGenerators(r.Request.Header.GetValuesRawKey("Content-Encoding"))
		if !ok {
			return nil, false, jpserror.ErrUnsupportedContentEncoding
		}
		//Create the body reader!
		r.body, err = pipe.From(r.RawReadCloser(), transferPipe, CEcodes, r.Pipes)
		return r.body, true, err
	}
	return r.body, false, nil
}
*/
