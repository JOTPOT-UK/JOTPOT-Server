package jps

//Response represents a generic response.
type Response interface {
	//Status is the ResponseStatus of the response.
	//The bool value will be true if the ResponseStatus accuratly represents the status of the response. This could be false if a specific enough ResponseStatus is not available. If you're implementing a protocol, we encourage you to create an issue wuth the ResponseStatus you are requesting, as we want this API to be as universal as possible.
	Status() (ResponseStatus, bool)
	//SetStatus sets the status of the response.
	//Note that some protcols may not have a response type as specific as the one supplied. In which case, a sensible reponse status will be used instead. Which statuses are used in these cases is up to the implementer.
	//The return value is the ResponseStatus that was used (either the same as the one passed, or a more generic ones, see the last sentance).
	SetStatus(ResponseStatus) error

	CacheSettings() (ResourceCacheSettings, error)
}

//IncomingResponse has a Response interface, along with an IncomingBody to read the response body from and a ClientSession.
type IncomingResponse struct {
	Response
	ConnectionDetails
	Body IncomingBody
}

//OutgoingResponse has a Response interface, aling with an OutgoingBody to write the response body to and a ServerSession.
type OutgoingResponse struct {
	Response
	ConnectionDetails
	Body OutgoingBody
}
