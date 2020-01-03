package jps

import "net/url"

//Request represents a generic request.
type Request interface {
	URL() *url.URL
	Method() Method
	SetMethod(Method) error
	//Range is the requested range of the response.
	//The first
	//Eg if the request is for a file of length 500 and Request.Range()=(5, 100)
	Ranges() ([]Range, error)
	SetRanges([]Range) error
}

type IncomingRequest struct {
	Request
	ConnectionDetails
	Body IncomingBody
}

type OutgoingRequest struct {
	Request
	ConnectionDetails
	Body OutgoingBody
}
