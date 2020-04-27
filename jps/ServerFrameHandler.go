package jps

import "sync"

//ServerFrameHandler is a handler for a server frame.
type ServerFrameHandler struct {
	//Protocol specifies which protocol this handler should be called for.
	//If Protocol is an empty string, the handler will be called for all protocols.
	Protocol string
	//If is the function that will be called to handle the server frame.
	//If the boolean result is true, or the error result is not nil, no further handlers will be called.
	F func(*ServerFrame) (bool, error)
}

//ServerFrameHandlerStage represents a stage in the process of handling a server frame.
type ServerFrameHandlerStage uint8

//Order policy of handler stages is as follows:
// Each handler stage should have a disused stage before and after it.
// When I'm initially designing this, I'm leaving gaps for ones I may wan to further add between these stages.
//So, for the first stage, it will be stage 1, to have 0 before it, and 2 after it.
//  The next stage would be stage 4 (3 before, 5 after), however I will leave some as gaps.

const (
	//StageInit is the first stage of a request.
	// Modifications to the request should not be made at this stage,
	// therefore making it suitable for inspection of an original request.
	StageInit = ServerFrameHandlerStage(1)
	//StageModify should be used to modify aspects of the request, except for the URL.
	// For example, if a header is to imply other header fields, or the data to be piped, then the handler to implement this should be at this stage.
	StageModify = ServerFrameHandlerStage(49)
	//StageAlias should be used to apply URL aliases.
	StageAlias = ServerFrameHandlerStage(97)
	//StageRedirect should be used to redirect any requests to different a URL.
	StageRedirect = ServerFrameHandlerStage(145)
	StageSettings = ServerFrameHandlerStage(190)
	//StageHandler should be used to implement handlers to create a response from the request!
	StageHandler = ServerFrameHandlerStage(193)
	//StageNotHandled should be used to catch requests that have not been handled.
	StageNotHandled = ServerFrameHandlerStage(241)
)

//ServerFrameHandlerListItem represents an item in a ServerFrameHandlerList
type ServerFrameHandlerListItem struct {
	handler  *ServerFrameHandler
	next     *ServerFrameHandlerListItem
	stage    ServerFrameHandlerStage
	priority uint16
}

//ServerFrameHandlerList represents a list of server frame handlers
type ServerFrameHandlerList struct {
	start *ServerFrameHandlerListItem
	lock  sync.RWMutex
}
