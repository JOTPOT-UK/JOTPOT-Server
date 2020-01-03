package http

import (
	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

type HTTPHeader interface {
	HTTPHeader() *header.Header
}

type HTTPMethod interface {
	HTTPMethod() string
	SetHTTPMethod(method string)
}

type HTTPStatus interface {
	HTTPStatus() (uint16, string)
	SetHTTPStatus(code uint16, text string)
	HTTPStatusCode() uint16
	SetHTTPStatusCode(code uint16)
	HTTPStatusText() string
	SetHTTPStatusText(text string)
}

type RequestInterface interface {
	jps.Request
	HTTPMethod
	HTTPHeader
}

type ResponseInterface interface {
	jps.Response
	HTTPStatus
	HTTPHeader
}
