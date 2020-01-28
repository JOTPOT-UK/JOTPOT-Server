package http1

import (
	"io"
	"strconv"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"

	"github.com/JOTPOT-UK/JOTPOT-Server/util"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe"
)

type BodyReader struct {
	ses       jps.Session
	config    *http.Config
	header    *header.Header
	req       *http.Request
	hasBody   func() bool
	rawReader io.ReadCloser
	reader    io.ReadCloser
	length    int64
}

func NewBodyReader(
	ses jps.Session,
	config *http.Config,
	header *header.Header,
	req *http.Request,
	hasBody func() bool,
	rawReader io.ReadCloser,
) BodyReader {
	return BodyReader{
		ses:       ses,
		config:    config,
		header:    header,
		req:       req,
		hasBody:   hasBody,
		rawReader: rawReader,
		length:    -2,
	}
}

func (r *BodyReader) Session() jps.Session {
	return r.ses
}

func (r *BodyReader) getBody() error {
	if !r.hasBody() || (r.req != nil && r.req.MethodStr == "HEAD") { //TODO: Only use hasBody()
		r.length = 0
		r.reader = util.EOFReadCloser{}
		return nil
	}
	codes := r.header.GetValues("Transfer-Encoding")
	lengths := r.header.GetValues("Content-Length")
	if len(codes) >= 1 {
		if len(lengths) != 0 {
			return http.ErrContentLengthAndTransferEncoding
		}
		//Check that none of the codes before the last are "chunked"
		for i := 0; i < len(codes)-1; i++ {
			if codes[i] == "chunked" { //TODO: Case sensitive? Does the header do this for us?
				return http.ErrMustChunkLast
			}
		}
		pipes, ok := r.config.TransferEncodings.GetReaderPipeGenerators(codes)
		if !ok {
			return http.ErrUnsupportedTransferEncoding
		}
		var err error
		r.reader, err = pipe.From(r.rawReader, pipes)
		if err != nil {
			return err
		}
		r.length = -1
	} else if len(lengths) == 1 {
		length, err := strconv.ParseUint(lengths[0], 10, 63)
		if err != nil {
			//TODO: Negative lengths?????
			return http.MakeErrMalformedContentLength(err)
		}
		r.length = int64(length)
		r.reader = util.ReaderClosePassOn{io.LimitReader(r.rawReader, int64(r.length)), r.rawReader}
	} else if len(lengths) > 1 {
		return http.ErrTooManyContentLength
	} else {
		//TODO: Should we expose this?
		r.reader = r.rawReader
		r.length = -1
	}
	return nil
}

func (r *BodyReader) BodyLength() (int64, error) {
	var err error
	if r.reader == nil {
		err = r.getBody()
	}
	return r.length, err
}

func (r *BodyReader) Body() (io.ReadCloser, error) {
	var err error
	if r.reader == nil {
		err = r.getBody()
	}
	return r.reader, err
}
