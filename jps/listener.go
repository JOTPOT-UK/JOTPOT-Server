package jps

import (
	"github.com/JOTPOT-UK/JOTPOT-Server/jpvariable"
)

type StreamSessionListener interface {
	AcceptSession(jpvariable.Variables, BufioSource) (StreamSession, error)
}

type StreamListener interface {
	AcceptStream() (Stream, error)
}
