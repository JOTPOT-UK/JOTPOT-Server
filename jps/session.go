package jps

import (
	"errors"

	"github.com/JOTPOT-UK/JOTPOT-Server/jpvariable"
)

//EOS should be returned from Frame() when the session has ended.
var EOS = errors.New("end of session")

type Session interface {
	ConnectionDetails
	//End ends the session; it does not close existing frames or streams.
	End() error

	Variables() jpvariable.Variables
	BufioSource() BufioSource
}

type StreamSession interface {
	Session
	//Close ends the session and closes all streams.
	Close() error
	AcceptStream() (Stream, error)
	RequestStream() (Stream, error)
}

/*type P2PSession interface {
	Session
	ReadFrame() (*Reader, error)
	WriteFrame() (*Writer, error)
}*/

type Hyjacker interface {
	Hyjack() (Stream, error)
}

type ServerSession interface {
	Session
	Hyjacker
	Server() *Server
	Frame() (*ServerFrame, error)
}

type ClientSession interface {
	Session
	Hyjacker
	Frame() (*ClientFrame, error)
}
