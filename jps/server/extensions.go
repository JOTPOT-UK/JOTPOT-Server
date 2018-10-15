package server

//HandlerCallback is the function signatre of a hander. It should take the IncomingRequest, and return false if the Server should continue, or true if the request has been handled.
type HandlerCallback func(*IncomingRequest) (bool, error)

//Handler contains a name for the handler, and its callback.
type Handler struct {
	Name string
	Cb   HandlerCallback
}

//HandlerStage is the stage in which the handler should be called
type HandlerStage uint8

//Order policy of handler stages is as follows:
// Each handler stage should have a disused stage before and after it.
// And, when I'm initially designing this, I'm leaving gaps for ones I may forget in between stages.
// So, for the first stage, it will be stage 1, to have 0 before it, and 2 after it.
//  The next stage would be stage 4 (3 before, 5 after), however I will save this, so it will be 7 instead.

const (
	//StageInit is the first stage of a request.
	// Modifications to the request should not be made at this stage,
	// therefore making it suitable for inspection of an original request.
	StageInit = HandlerStage(1)
	//StageModify should be used to modify aspects of the request, except for the URL.
	// For example, if a header is to imply other header fields, or the data to be piped, then the handler to implement this should be at this stage.
	StageModify = HandlerStage(7)
	//StageAlias should be used to apply URL aliases.
	StageAlias = HandlerStage(13)
	//StageRedirect should be used to redirect any requests to different a URL.
	StageRedirect = HandlerStage(19)
	//StageHandler should be used to implement handlers to create a response from the request!
	StageHandler = HandlerStage(25)
)

//HandlerListItem is an item within a HandleList
type HandlerListItem struct {
	handler  *Handler
	next     *HandlerListItem
	stage    HandlerStage
	priority uint16
}

//HandlerList is a priority linked list of Handlers.
type HandlerList struct {
	start *HandlerListItem
}

//insert inserts j into the HandlerList
func (l *HandlerList) insert(j *HandlerListItem) {
	if l.start == nil {
		//If there are no items in the list, then this is the only item
		l.start = j
	} else if (j.stage == l.start.stage && j.priority > l.start.priority) || j.stage < l.start.stage {
		//If we are at a lower stage or at a higher priority and the same stage as the first item,
		// then this is the new first item!
		j.next = l.start
		l.start = j
	} else {
		//We must now iterate through the list and find its place...
		//I originally did this recursively, but not having to bother with call overhead makes this a fast approach.
		//For every item, see if we go after it...
		i := l.start
		for ; i.next != nil; i = i.next {
			//Check if we go after (same check as above)
			if (j.stage == i.next.stage && j.priority > i.next.priority) || j.stage < i.next.stage {
				//Place j after i
				j.next = i.next
				i.next = j
				return
			}
		}
		//If the loop doesn't return, then we are at the end of the list.
		i.next = j
		j.next = nil
	}
}

//CallFrom calls all of the Handlers from c. It returns if one of the handlers returns either true, or a non-nil error.
func (l *HandlerList) CallFrom(req *IncomingRequest, c *HandlerListItem) (handled bool, err error) {
	for {
		handled, err = c.handler.Cb(req)
		c = c.next
		if handled || err != nil || c == nil {
			return
		}
	}
}

//CallFromTo calls all the Handlers from c, and stops if one returns either true or a non-nil error, or, the Handler stage is greater than to.
func (l *HandlerList) CallFromTo(req *IncomingRequest, c *HandlerListItem, to HandlerStage) (*HandlerListItem, bool, error) {
	var handled bool
	var err error
	for {
		handled, err = c.handler.Cb(req)
		if handled || err != nil {
			return c, handled, err
		}
		c = c.next
		if c == nil || c.stage > to {
			return c, handled, err
		}
	}
}

//Call calls all the Handlers in the HandlerList. If a Handler returns true or a non-nil error, then no more are called, and the returns parameters are returned.
func (l *HandlerList) Call(req *IncomingRequest) (bool, error) {
	if l.start == nil {
		return false, nil
	}
	return l.CallFrom(req, l.start)
}

//CallTo ... you get the idea
func (l *HandlerList) CallTo(req *IncomingRequest, to HandlerStage) (*HandlerListItem, bool, error) {
	if l.start == nil {
		return nil, false, nil
	}
	return l.CallFromTo(req, l.start, to)
}
