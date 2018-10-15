package jpserror

import (
	"strconv"
)

var ErrInternalServerError = &HTTPError{Code: 500, Message: "Internal server error"}
var ErrUnsupportedTransferEncoding = &HTTPError{Code: 501, Message: "Unsupported Transfer-Encoding"}
var ErrUnsupportedContentEncoding = &HTTPError{Code: 415, Message: "Unsupported Content-Encoding"}
var ErrContentLengthAndTransferEncoding = &HTTPError{Code: 400, Message: "It is forbidden to have both a Content-Length and a Transfer-Encoding header"}
var ErrMalformedContentLength = &HTTPError{Code: 400, Message: "Malformed Content-Length"}
var ErrTooManyContentLength = &HTTPError{Code: 400, Message: "Too many Content-Length values"}
var ErrMustOnlyChunkOnce = &HTTPError{Code: 400, Message: "Chunked encoding must only be applied once"}
var ErrMustChunkLast = &HTTPError{Code: 400, Message: "The final Transfer-Encoding must be chunked"}
var ErrMalformedHTTPRequest = &HTTPError{Code: 400, Message: "Malformed HTTP request"}
var ErrMalformedHTTPVersion = &HTTPError{Code: 400, Message: "Malformed HTTP version"}
var ErrMalformedURI = &HTTPError{Code: 400, Message: "Malformed URI"}
var ErrMalformedHeaders = &HTTPError{Code: 400, Message: "Malformed headers"}
var ErrMustBe1HostHeader = &HTTPError{Code: 400, Message: "There must be 1 Host header"}
var ErrCannotChunkAnAlreadyChunkedBody = &HTTPError{Code: 400, Message: "Cannot chunk an already chunked body"}

//HTTPError represents an error, but with a custom code and message.
type HTTPError struct {
	//Err is the error which caused this error
	Err error
	//Code is the code which the server should respond with
	Code int
	//Message is the message which an error page should display
	Message string
}

func (err *HTTPError) Error() string {
	if err.Err == nil {
		return strconv.FormatInt(int64(err.Code), 10) + ": " + err.Message
	}
	return err.Err.Error()
}
