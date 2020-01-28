package http

import (
	"errors"
	"strconv"
)

func MakeErrInternalServerError(err error) error {
	return Error{Code: 500, Message: "Internal server error", Err: err}
}

var ErrInternalServerError = MakeErrInternalServerError(nil)
var ErrUnsupportedTransferEncoding = &Error{Code: 501, Message: "Unsupported Transfer-Encoding"}
var ErrUnsupportedContentEncoding = &Error{Code: 415, Message: "Unsupported Content-Encoding"}
var ErrContentLengthAndTransferEncoding = &Error{Code: 400, Message: "It is forbidden to have both a Content-Length and a Transfer-Encoding header"}

func MakeErrMalformedContentLength(err error) error {
	return Error{Code: 400, Message: "Malformed Content-Length", Err: err}
}

var ErrTooManyContentLength = &Error{Code: 400, Message: "Too many Content-Length values"}
var ErrMustOnlyChunkOnce = &Error{Code: 400, Message: "Chunked encoding must only be applied once"}
var ErrMustChunkLast = &Error{Code: 400, Message: "The final Transfer-Encoding must be chunked"}

func MakeErrMalformedHTTPRequest(err error) error {
	return Error{Code: 400, Message: "Malformed HTTP request", Err: err}
}

var ErrExpectingCarraigeReturnBeforeNewline = errors.New("expecting carraige return before newline")
var ErrExpecting2SpacesOnFirstRequestLine = errors.New("expecting 2 spaces on first request line")
var ErrMalformedHTTPRequest = MakeErrMalformedHTTPRequest(nil)
var ErrMalformedHTTPRequestExpectingCarraigeReturnBeforeNewline = MakeErrMalformedHTTPRequest(ErrExpectingCarraigeReturnBeforeNewline)
var ErrMalformedHTTPRequestExpecting2SpacesOnFirstRequestLine = MakeErrMalformedHTTPRequest(ErrExpecting2SpacesOnFirstRequestLine)

func MakeErrMalformedHTTPVersion(err error) error {
	return Error{Code: 400, Message: "Malformed HTTP version", Err: err}
}

var ErrHTTPVersionDoesNotStartWithHTTP = errors.New("HTTP version does not start with \"HTTP/\"")
var ErrHTTPVersionNoVersion = errors.New("HTTP version does have anything \"HTTP/\"")
var ErrHTTPVersionNoDot = errors.New("HTTP version does contain a '.'")
var ErrMalformedHTTPVersion = MakeErrMalformedHTTPVersion(nil)
var ErrMalformedHTTPVersionDoesNotStartWithHTTP = MakeErrMalformedHTTPVersion(ErrHTTPVersionDoesNotStartWithHTTP)
var ErrMalformedHTTPVersionNoVersion = MakeErrMalformedHTTPVersion(ErrHTTPVersionNoVersion)
var ErrMalformedHTTPVersionNoDot = MakeErrMalformedHTTPVersion(ErrHTTPVersionNoDot)

func MakeErrMalformedURI(err error) error {
	return Error{Code: 400, Message: "Malformed URI", Err: err}
}

func MakeErrMalformedHeaders(err error) error {
	return Error{Code: 400, Message: "Malformed HTTP headers", Err: err}
}

var ErrMustBe1HostHeader = &Error{Code: 400, Message: "There must be 1 Host header", Err: errors.New("the HTTP version requires 1 Host to be given")}
var ErrCannotChunkAnAlreadyChunkedBody = &Error{Code: 400, Message: "Cannot chunk an already chunked body"}

var ErrBytesRangeSpecMustContainDash = errors.New("a byte-range-spec must contain a \"-\"")

func MakeErrMalformedRangeHeader(err error) error {
	return Error{Code: 400, Message: "Malformed Range header", Err: err}
}
func MakeErrRangeNotSatisfiable(err error) error {
	return Error{Code: 416, Message: "Range Not Satisfiable", Err: err}
}

func MakeErrMalformedCacheControlHeader(err error) error {
	return Error{Code: 400, Message: "Malformed Cache-Control header", Err: err}
}
var ErrExpectingArgumentAfterEquals = errors.New("expecting argument after equals")
var ErrMalformedCacheControlHeaderExpectingArgumentAfterEquals = MakeErrMalformedCacheControlHeader(ErrExpectingArgumentAfterEquals)

type Error struct {
	Code    uint16
	Message string
	Err     error
}

func (err Error) Error() string {
	return "(HTTP " + strconv.FormatInt(int64(err.Code), 10) + ") " + err.Message + ": " + err.Err.Error()
}

func (err Error) UserSafeStr() string {
	return err.Message
}

func (err Error) Unwrap() error {
	return err.Err
}
