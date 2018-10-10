package server

import (
	"bufio"
	"io"
	"jotpot/net/jps"
	"jotpot/net/jps/pipe"
	"jotpot/net/jps/pipe/limiter"
	"jotpot/net/util"
	"net"
	"strconv"
	"strings"
)

type IncomingRequest struct {
	Server  *Server
	Request *jps.Request
	Pipes   []*pipe.ReaderPipeGenerator

	con       net.Conn
	rawReader *bufio.Reader
	body      io.ReadCloser
}

func NewIncomingRequest(s *Server, con net.Conn, reader *bufio.Reader) *IncomingRequest {
	return &IncomingRequest{
		Server:    s,
		con:       con,
		rawReader: reader,
	}
}

type incomingRequestRawReadCloser struct {
	r *IncomingRequest
}

func (r incomingRequestRawReadCloser) Read(buf []byte) (int, error) {
	return r.r.rawReader.Read(buf)
}
func (r incomingRequestRawReadCloser) Close() error {
	return r.r.con.Close()
}
func (r *IncomingRequest) RawReadCloser() io.ReadCloser {
	if r.rawReader == nil {
		r.rawReader = bufio.NewReaderSize(r.con, r.Server.ReaderBufSize)
	}
	return incomingRequestRawReadCloser{r}
}

func (r *IncomingRequest) GotRawReader() bool {
	return r.rawReader != nil
}
func (r *IncomingRequest) GetRawReader() (*bufio.Reader, bool) {
	if r.rawReader == nil {
		r.rawReader = bufio.NewReaderSize(r.con, r.Server.ReaderBufSize)
		return r.rawReader, true
	}
	return r.rawReader, false
}
func (r *IncomingRequest) RawReader() *bufio.Reader {
	if r.rawReader == nil {
		r.rawReader = bufio.NewReaderSize(r.con, r.Server.ReaderBufSize)
	}
	return r.rawReader
}

func (r *IncomingRequest) HasBody() bool {
	return len(r.Request.Header.GetValuesRawKey("Transfer-Encoding")) != 0 || len(r.Request.Header.GetValuesRawKey("Content-Length")) != 0
}
func (r *IncomingRequest) GotBody() bool {
	return r.body != nil
}
func (r *IncomingRequest) GetBody() (io.ReadCloser, bool, *jps.HTTPError, error) {
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
				return nil, false, jps.ErrContentLengthAndTransferEncoding, nil
			}
			//Parse the length
			var CL uint64
			CL, err = strconv.ParseUint(CLH[0], 10, 63)
			if err != nil {
				return nil, false, jps.ErrMalformedContentLength, nil
			}
			//The decoder is just a limiter
			limiter := limiter.GenerateLimiterGenerator(int64(CL))
			transferPipe = []*pipe.ReaderPipeGenerator{&limiter}
		} else if len(TEH) != 0 {
			//If we have a transfer-encoding, we can't also have a content-length
			if len(CLH) != 0 {
				return nil, false, jps.ErrContentLengthAndTransferEncoding, nil
			}
			//Checl that we are chunking last and only last
			if strings.ToLower(TEH[len(TEH)-1]) != "chunked" {
				//TEH = append(TEH, "chunked")
				return nil, false, jps.ErrMustChunkLast, nil
			}
			for i := 0; i < len(TEH)-1; i++ {
				if TEH[i] == "chunked" {
					return nil, false, jps.ErrMustOnlyChunkOnce, nil
				}
			}
			//Get the transfer pipes!
			transferPipe, ok = r.Server.TransferEncodings.GetReaderPipeGenerators(TEH)
			if !ok {
				return nil, false, jps.ErrUnsupportedTransferEncoding, nil
			}
		} else if len(CLH) == 0 {
			//If we have no content length and no tranfer encoding, then there is no body
			r.body = util.EOFReadCloser{}
			return r.body, true, nil, nil
		} else {
			//At this point, we must have more than 1 content-length header
			return nil, false, jps.ErrTooManyContentLength, nil
		}
		//Get the pipes for the Content-Encoding
		CEcodes, ok := r.Server.ContentEncodings.GetReaderPipeGenerators(r.Request.Header.GetValuesRawKey("Content-Encoding"))
		if !ok {
			return nil, false, jps.ErrUnsupportedContentEncoding, nil
		}
		//Create the body reader!
		r.body, err = pipe.From(r.RawReadCloser(), transferPipe, CEcodes, r.Pipes)
		return r.body, true, nil, err
	}
	return r.body, false, nil, nil
}
