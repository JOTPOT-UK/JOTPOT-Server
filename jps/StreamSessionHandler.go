package jps

import "sync"

type StreamSessionHandler struct {
	Protocol string
	F        func(StreamSession) (bool, error)
}

//StreamSessionHandlerStage is the stage in which the handler should be called
type StreamSessionHandlerStage uint8

const (
	StreamSessionStageInit       = StreamSessionHandlerStage(1)
	StreamSessionStageDeadlines  = StreamSessionHandlerStage(49)
	StreamSessionStageHandle     = StreamSessionHandlerStage(97)
	StreamSessionStageNotHandled = StreamSessionHandlerStage(145)
)

//StreamSessionHandlerListItem represents an item in a StreamSessionHandlerList
type StreamSessionHandlerListItem struct {
	handler  *StreamSessionHandler
	next     *StreamSessionHandlerListItem
	stage    StreamSessionHandlerStage
	priority uint16
}

//StreamSessionHandlerList represents a list of server frame handlers
type StreamSessionHandlerList struct {
	start *StreamSessionHandlerListItem
	lock  sync.RWMutex
}
