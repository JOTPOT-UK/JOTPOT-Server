package server

import (
	"jotpot/net/http/http1/encoding"
)

type Server struct {
	Handlers          HandlerList
	TransferEncodings encoding.List
	ContentEncodings  encoding.List
	ReaderBufSize     int
}

func (s *Server) Handle(stage HandlerStage, priority uint16, handler *Handler) {
	s.Handlers.Insert(&HandlerListItem{
		Handler:  handler,
		Stage:    stage,
		Priority: priority,
	})
}
