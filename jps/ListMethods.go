package jps

//insert inserts j into the HandlerList
func (l *ServerSessionHandlerList) insert(j *ServerSessionHandlerListItem) {
	l.lock.Lock()
	defer l.lock.Unlock()
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
		i := l.start
		for ; i.next != nil; i = i.next {
			//Insert j after i if j goes before the item after i
			if (j.stage == i.next.stage && j.priority > i.next.priority) || j.stage < i.next.stage {
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

//insert inserts j into the HandlerList
func (l *StreamSessionHandlerList) insert(j *StreamSessionHandlerListItem) {
	l.lock.Lock()
	defer l.lock.Unlock()
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
		i := l.start
		for ; i.next != nil; i = i.next {
			//Insert j after i if j goes before the item after i
			if (j.stage == i.next.stage && j.priority > i.next.priority) || j.stage < i.next.stage {
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

//insert inserts j into the HandlerList
func (l *ServerFrameHandlerList) insert(j *ServerFrameHandlerListItem) {
	l.lock.Lock()
	defer l.lock.Unlock()
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
		i := l.start
		for ; i.next != nil; i = i.next {
			//Insert j after i if j goes before the item after i
			if (j.stage == i.next.stage && j.priority > i.next.priority) || j.stage < i.next.stage {
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

//callFrom calls all of the Handlers from c onwards.
//It returns if one of the handlers returns either true, or a non-nil error.
//The return value is the return value of the last handler called.
func (l *ServerSessionHandlerList) callFrom(c *ServerSessionHandlerListItem, ses ServerSession) (handled bool, err error) {
	proto := ses.Protocol().Protocol()
	for {
		if proto == c.handler.Protocol || c.handler.Protocol == "" {
			handled, err = c.handler.F(ses)
			if handled || err != nil {
				return
			}
		}
		c = c.next
		if c == nil {
			return
		}
	}
}

//callFrom calls all of the Handlers from c onwards.
//It returns if one of the handlers returns either true, or a non-nil error.
//The return value is the return value of the last handler called.
func (l *StreamSessionHandlerList) callFrom(c *StreamSessionHandlerListItem, ses StreamSession) (handled bool, err error) {
	proto := ses.Protocol().Protocol()
	for {
		if proto == c.handler.Protocol || c.handler.Protocol == "" {
			handled, err = c.handler.F(ses)
			if handled || err != nil {
				return
			}
		}
		c = c.next
		if c == nil {
			return
		}
	}
}

//callFrom calls all of the Handlers from c onwards.
//It returns if one of the handlers returns either true, or a non-nil error.
//The return value is the return value of the last handler called.
func (l *StreamSessionHandlerList) callFromWithHandler(c *StreamSessionHandlerListItem, h StreamSessionHandler, ses StreamSession) (handled bool, err error) {
	proto := ses.Protocol().Protocol()
	for {
		if c.stage == StreamSessionStageHandle {
			if proto == h.Protocol || h.Protocol == "" {
				handled, err = h.F(ses)
				if handled || err != nil {
					return
				}
			}
			return l.callFrom(c, ses)
		}
		if proto == c.handler.Protocol || c.handler.Protocol == "" {
			handled, err = c.handler.F(ses)
			if handled || err != nil {
				return
			}
		}
		c = c.next
		if c == nil {
			return
		}
	}
}

//callFrom calls all of the Handlers from c onwards.
//It returns if one of the handlers returns either true, or a non-nil error.
//The return value is the return value of the last handler called.
func (l *ServerFrameHandlerList) callFrom(c *ServerFrameHandlerListItem, frame *ServerFrame) (handled bool, err error) {
	proto := frame.Protocol().Protocol()
	for {
		if c.handler.Protocol == "" || proto == c.handler.Protocol {
			handled, err = c.handler.F(frame)
			if handled || err != nil {
				return
			}
		}
		c = c.next
		if c == nil {
			return
		}
	}
}

//Handle calls all the Handlers in the HandlerList in priorty and stage order. If a Handler returns true or a non-nil error, then no more are called, and the returned parameters are returned by Handle.
func (l *ServerSessionHandlerList) Handle(ses ServerSession) (bool, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.start == nil {
		return false, nil
	}
	return l.callFrom(l.start, ses)
}

func (l *StreamSessionHandlerList) Handle(ses StreamSession) (bool, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.start == nil {
		return false, nil
	}
	return l.callFrom(l.start, ses)
}

func (l *StreamSessionHandlerList) HandleWithHandler(ses StreamSession, h StreamSessionHandler) (bool, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.start == nil {
		return false, nil
	}
	return l.callFromWithHandler(l.start, h, ses)
}

//Handle calls all the Handlers in the HandlerList in priorty and stage order. If a Handler returns true or a non-nil error, then no more are called, and the returned parameters are returned by Handle.
func (l *ServerFrameHandlerList) Handle(frame *ServerFrame) (bool, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.start == nil {
		return false, nil
	}
	return l.callFrom(l.start, frame)
}
