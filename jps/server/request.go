package server

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps/jpserror"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe/limiter"
	"github.com/JOTPOT-UK/JOTPOT-Server/util"
)

//IncomingRequest represents an incoming request to a server
type IncomingRequest struct {
	//Server is the Server which recieved the IncomingRequest
	Server *Server
	//Request is the parsed Request
	Request *jps.Request
	//Pipes is a slice of pointers to pipe generators, which, when the body is read, the data will be piped through after all the Transfer-Encoding and Content-Encoding pipes.
	// Any changes made to this slice after GetBody has been called do not have an effect.
	Pipes []*pipe.ReaderPipeGenerator

	con       net.Conn
	rawReader *bufio.Reader
	body      io.ReadCloser
}

//NewIncomingRequest creates a new IncomingRequest struct, with the given server, connection, and reader
func NewIncomingRequest(s *Server, con net.Conn, reader *bufio.Reader) *IncomingRequest {
	return &IncomingRequest{
		Server:    s,
		con:       con,
		rawReader: reader,
	}
}

//incomingRequestRawReadCloser is to provide an interface from which the raw data can be read from the request
type incomingRequestRawReadCloser struct {
	r *IncomingRequest
}

//Read calls Read on the underlying bufio.Reader of the request
func (r incomingRequestRawReadCloser) Read(buf []byte) (int, error) {
	return r.r.rawReader.Read(buf)
}

//Read calls Close on the underlying connection of the request
func (r incomingRequestRawReadCloser) Close() error {
	return r.r.con.Close()
}

//RawReadCloser returns a ReadCloser which reads from the raw data of the incoming request
// (implemented by the incomingRawReadCloser type)
func (r *IncomingRequest) RawReadCloser() io.ReadCloser {
	if r.rawReader == nil {
		r.rawReader = bufio.NewReaderSize(r.con, r.Server.ReaderBufSize)
	}
	return incomingRequestRawReadCloser{r}
}

//GotRawReader returns true if the rawReader of the request is
func (r *IncomingRequest) GotRawReader() bool {
	return r.rawReader != nil
}

//GetRawReader returns a bufio.Reader that reads from the connection - it only creates one once, and the boolean result is false if it has used a prexisting one.
func (r *IncomingRequest) GetRawReader() (*bufio.Reader, bool) {
	if r.rawReader == nil {
		r.rawReader = bufio.NewReaderSize(r.con, r.Server.ReaderBufSize)
		return r.rawReader, true
	}
	return r.rawReader, false
}

//RawReader is equivilent to GetRawReader, except that the second return parameter is not given.
func (r *IncomingRequest) RawReader() *bufio.Reader {
	if r.rawReader == nil {
		r.rawReader = bufio.NewReaderSize(r.con, r.Server.ReaderBufSize)
	}
	return r.rawReader
}

//HasBody returns true if the request has a body
func (r *IncomingRequest) HasBody() bool {
	return len(r.Request.Header.GetValuesRawKey("Transfer-Encoding")) != 0 || len(r.Request.Header.GetValuesRawKey("Content-Length")) != 0
}

//GotBody returns true if a body reader has already been created
func (r *IncomingRequest) GotBody() bool {
	return r.body != nil
}

//GetBody the reader which reads the deocoded body. If a new reader is created (because it hasn't been created before), true is returned as the boolean result.
func (r *IncomingRequest) GetBody() (io.ReadCloser, bool, error) {
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
