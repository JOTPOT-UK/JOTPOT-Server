package server

type HandlerCallback func(*IncomingRequest) bool

type Handler struct {
	Name string
	Cb   HandlerCallback
}

type HandlerStage uint8

const (
	StageM0 HandlerStage = iota
	StageInit
	StageM1
	StageModify
	StageM2
	StageAlias
	StageM3
	StageRedirect
	StageM4
	StageHandler
	StageM5
	StageLog
	StageM6
	StageM7
	StageM8
	StageM9
)

type HandlerListItem struct {
	Handler  *Handler
	Next     *HandlerListItem
	Stage    HandlerStage
	Priority uint16
}

type HandlerList struct {
	Start *HandlerListItem
}

func (i *HandlerListItem) InsertAfter(j *HandlerListItem) {
	if (j.Stage == i.Next.Stage && j.Priority < i.Next.Priority) || j.Stage < i.Next.Stage {
		j.Next = i.Next
		i.Next = j
	} else {
		i.Next.InsertAfter(j)
	}
}
func (l *HandlerList) Insert(j *HandlerListItem) {
	if l.Start == nil {
		l.Start = j
	} else if (j.Stage == l.Start.Stage && j.Priority < l.Start.Priority) || j.Stage < l.Start.Stage {
		j.Next = l.Start
		l.Start = j
	} else {
		l.Start.InsertAfter(j)
	}
}

func (l *HandlerList) CallFrom(req *IncomingRequest, c *HandlerListItem) bool {
	for {
		if c.Handler.Cb(req) {
			return true
		}
		c = c.Next
		if c == nil {
			return false
		}
	}
}
func (l *HandlerList) CallFromTo(req *IncomingRequest, c *HandlerListItem, to HandlerStage) (bool, *HandlerListItem) {
	for {
		if c.Handler.Cb(req) {
			return true, c
		}
		c = c.Next
		if c == nil || c.Stage > to {
			return false, c
		}
	}
}
func (l *HandlerList) Call(req *IncomingRequest) bool {
	if l.Start == nil {
		return false
	}
	return l.CallFrom(req, l.Start)
}
func (l *HandlerList) CallTo(req *IncomingRequest, to HandlerStage) (bool, *HandlerListItem) {
	if l.Start == nil {
		return false, nil
	}
	return l.CallFromTo(req, l.Start, to)
}
