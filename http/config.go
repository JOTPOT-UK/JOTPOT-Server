package http

import (
	"io"
	gohttp "net/http"
	"strconv"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/http1/encoding"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

//SimpleResponseWriter writes a HTTP response with the given status code, connection: close, content-type: text/plain; charser=utf-8, and with the given body to w.
func SimpleResponseWriter(w io.Writer, code int, body string) {
	w.Write([]byte(
		"HTTP/1.1 " + strconv.FormatUint(uint64(code), 10) + " " + gohttp.StatusText(code) +
			"\r\nConnection: close" +
			"\r\nContent-Length: " + strconv.FormatUint(uint64(len(body)), 10) +
			"\r\nContent-Type: text/plain; charset=utf-8" +
			"\r\n\r\n" + body,
	))
}

type ServerErrorHandler func(jps.ServerFrame, io.Writer, error)

func DefaultServerErrorHandler(frame jps.ServerFrame, w io.Writer, err error) {
	var hErr *Error
	switch e := err.(type) {
	case Error:
		hErr = &e
	case *Error:
		hErr = e
	default:
		hErr = &Error{
			Code:    500,
			Message: "Internal Server Error",
		}
	}
	SimpleResponseWriter(w, int(hErr.Code), hErr.Message)
}

type Config struct {
	TransferEncodings encoding.List
	ContentEncodings  encoding.List
	HeaderProcessor   *header.Processor
	UseJPHeaders      bool

	ServerErrorHandler ServerErrorHandler
}

var DefaultConfig = Config{
	UseJPHeaders: false,

	ServerErrorHandler: DefaultServerErrorHandler,
}
