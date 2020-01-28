package http1

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe"
	"github.com/JOTPOT-UK/JOTPOT-Server/util"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
)

var ErrBodyLengthExceded = errors.New("body length exceded")
var ErrNoBody = errors.New("no body")

type writerWrapper struct {
	WriteFlusher util.WriteFlusher
	Closer       func() error
	HeaderCB     func() []byte
	HeadWritten  bool
}

func (ww *writerWrapper) writeHeader() error {
	head := ww.HeaderCB()
	n, err := ww.WriteFlusher.Write(head)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if n != len(head) {
		return fmt.Errorf("failed to write header: %w", io.ErrShortWrite)
	}
	ww.HeadWritten = true
	return nil
}

func (ww *writerWrapper) Write(src []byte) (int, error) {
	if !ww.HeadWritten {
		head := ww.HeaderCB()
		n, err := ww.WriteFlusher.Write(append(head, src...))
		n -= len(head)
		if n < 0 {
			if err == nil {
				err = fmt.Errorf("failed to write header: %w", io.ErrShortWrite)
			}
		} else {
			ww.HeadWritten = true
		}
		return n, err
	}
	return ww.WriteFlusher.Write(src)
}

func (ww *writerWrapper) Flush() error {
	if !ww.HeadWritten {
		err := ww.writeHeader()
		if err != nil {
			ww.WriteFlusher.Flush()
			return err
		}
	}
	return ww.WriteFlusher.Flush()
}

func (ww *writerWrapper) Close() error {
	if !ww.HeadWritten {
		err := ww.writeHeader()
		if err != nil {
			ww.Closer()
			return err
		}
	}
	return ww.Closer()
}

type nilBody struct {
	util.CloseFlusher
}

func (_ nilBody) Write(_ []byte) (int, error) {
	return 0, ErrNoBody
}

type LimitWriteFlushCloser struct {
	i, l int64
	w    jps.WriteFlushCloser
}

func (w LimitWriteFlushCloser) Write(buf []byte) (int, error) {
	w.i += int64(len(buf))
	if w.i > w.l {
		overflow := len(buf) - int(w.i-w.l)
		w.i = w.l
		toWrite := len(buf) - overflow
		if toWrite == 0 {
			return 0, ErrBodyLengthExceded
		}
		n, err := w.w.Write(buf[:toWrite])
		if n == toWrite || err == nil {
			err = ErrBodyLengthExceded
		}
		return n, err
	}
	return w.w.Write(buf)
}

func (w LimitWriteFlushCloser) Close() error {
	return w.w.Close()
}

func (w LimitWriteFlushCloser) Flush() error {
	return w.w.Flush()
}

type CloseCBWriteFlushCloser struct {
	closed bool
	close  func() (error, bool)
	passon jps.WriteFlushCloser
}

func (w *CloseCBWriteFlushCloser) Write(src []byte) (int, error) {
	if w.closed {
		return 0, io.ErrClosedPipe
	}
	return w.passon.Write(src)
}

func (w *CloseCBWriteFlushCloser) Flush() error {
	if w.closed {
		return io.ErrClosedPipe
	}
	return w.passon.Flush()
}

func (w *CloseCBWriteFlushCloser) Close() (err error) {
	if w.closed {
		return io.ErrClosedPipe
	}
	//TODO: Fix properly!!!
	/*if err = w.passon.Close(); err != nil {
		return
	}*/
	err, w.closed = w.close()
	return
}

type BodyWriter struct {
	ses         jps.Session
	config      *http.Config
	header      *header.Header
	req         *http.Request
	hasBody     func() bool
	finalWriter writerWrapper
	writer      jps.WriteFlushCloser
}

func NewBodyWriter(
	ses jps.Session,
	config *http.Config,
	header *header.Header,
	req *http.Request,
	hasBody func() bool,
	rawWriter jps.WriteFlushCloser,
	headerGenerator func() []byte,
	close func() error,
) BodyWriter {
	return BodyWriter{
		ses:     ses,
		config:  config,
		header:  header,
		req:     req,
		hasBody: hasBody,
		finalWriter: writerWrapper{
			WriteFlusher: rawWriter,
			Closer:       close,
			HeaderCB:     headerGenerator,
		},
	}
}

func (w *BodyWriter) Session() jps.Session {
	return w.ses
}

func (w *BodyWriter) BodyLength() (int64, error) {
	if w.req != nil && w.req.MethodStr == "HEAD" {
		return 0, nil
	}
	lens := w.header.GetValues("Content-Length")
	if len(lens) == 0 {
		return -1, nil
	} else if len(lens) == 1 {
		l, err := strconv.ParseUint(lens[0], 10, 63)
		if err != nil {
			return -2, http.MakeErrMalformedContentLength(err)
		}
		return int64(l), nil
	} else {
		return -2, http.ErrTooManyContentLength
	}
}

func (w *BodyWriter) SetBodyLength(length int64) error {
	if length >= 0 {
		w.header.Set("Content-Length", strconv.FormatUint(uint64(length), 10))
		codes := w.header.GetValues("Transfer-Encoding")
		last := len(codes) - 1
		if last > -1 && codes[last] == "chunked" {
			w.header.SetValues("Transfer-Encoding", codes[:last])
		}
	} else if length == -1 {
		w.header.Del("Content-Length")
		codes := w.header.GetValues("Transfer-Encoding")
		last := len(codes) - 1
		if last < 0 || codes[last] != "chunked" {
			w.header.Add("Transfer-Encodings", "chunked")
		}
	} else {
		panic("body length cannot be less than -1")
	}
	return nil
}

func (w *BodyWriter) Body() (jps.WriteFlushCloser, error) {
	if w.writer == nil {
		if w.hasBody() {
			//TIMER:start := time.Now()
			l, err := w.BodyLength()
			if err != nil {
				return nil, err
			}
			//TIMER:jps.HBLengthTimes = append(jps.HBLengthTimes, time.Since(start))
			//TIMER:start = time.Now()
			if l == -1 {
				codes := w.header.GetValuesRawKey("Transfer-Encoding")
				lm1 := len(codes) - 1
				for i := 0; i < lm1; i++ {
					if codes[i] == "chunked" {
						return nil, http.ErrMustChunkLast
					}
				}
				//TODO: Close connection if chunked is not last
				pipes, ok := w.config.TransferEncodings.GetWriterPipeGenerators(codes)
				if !ok {
					return nil, http.ErrUnsupportedTransferEncoding
				}
				if w.writer, err = pipe.To(&w.finalWriter, pipes); err != nil {
					return w.writer, err
				}
			} else {
				//TIMER:jps.HBHeadWriteTimes = append(jps.HBHeadWriteTimes, time.Since(start))
				w.writer = LimitWriteFlushCloser{
					l: l,
					w: &w.finalWriter,
				}
			}
		} else {
			w.writer = nilBody{&w.finalWriter}
		}
	}
	return w.writer, nil
}
