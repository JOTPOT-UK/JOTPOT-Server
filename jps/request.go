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
	//Conditions returns a list of conditions for the request.
	//See the documentation on Condition for how they should be handled.
	//It is sometimes okay to ignore conditions and behave as if there were none, however this will lead to unesisary data transmition and, in quite a few cases, the conditions MUST be followed.
	Conditions() ([]Condition, error)
	CacheSettings() (RequestCacheSettings, error)
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
