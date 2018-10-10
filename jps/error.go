package jps

var ErrInternalServerError = &HTTPError{500, "Internal server error"}
var ErrUnsupportedTransferEncoding = &HTTPError{501, "Unsupported Transfer-Encoding"}
var ErrUnsupportedContentEncoding = &HTTPError{415, "Unsupported Content-Encoding"}
var ErrContentLengthAndTransferEncoding = &HTTPError{400, "It is forbidden to have both a Content-Length and a Transfer-Encoding header"}
var ErrMalformedContentLength = &HTTPError{400, "Malformed Content-Length"}
var ErrTooManyContentLength = &HTTPError{400, "Too many Content-Length values"}
var ErrMustOnlyChunkOnce = &HTTPError{400, "Chunked encoding must only be applied once"}
var ErrMustChunkLast = &HTTPError{400, "The final Transfer-Encoding must be chunked"}

type HTTPError struct {
	Code    int
	Message string
}
