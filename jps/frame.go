package jps

type ServerFrame struct {
	Request  IncomingRequest
	Response OutgoingResponse
	Session  ServerSession
	ConnectionDetails
}

type ClientFrame struct {
	Request  OutgoingRequest
	Response IncomingResponse
	Session  ClientSession
	ConnectionDetails
}
