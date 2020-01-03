package jps

import "sync"

type ServerSessionHandler struct {
	Protocol string
	F        func(ServerSession) (bool, error)
}

//ServerSessionHandlerStage is the stage in which the handler should be called
type ServerSessionHandlerStage uint8

const (
	ServerSessionStageInit = ServerSessionHandlerStage(1)
)

//ServerSessionHandlerListItem represents an item in a ServerSessionHandlerList
type ServerSessionHandlerListItem struct {
	handler  *ServerSessionHandler
	next     *ServerSessionHandlerListItem
	stage    ServerSessionHandlerStage
	priority uint16
}

//ServerSessionHandlerList represents a list of server frame handlers
type ServerSessionHandlerList struct {
	start *ServerSessionHandlerListItem
	lock  sync.RWMutex
}
