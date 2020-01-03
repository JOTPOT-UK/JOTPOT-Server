package jps

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/jpvariable"
)

//State is used to identify the state of the server.
type State byte

const (
	//StateClosed is the state of the server when it is closed - ie it will not accept incoming requests, and there are no open requests.
	StateClosed State = iota
	//StateClosing is the state of the server while it is closing. In this state, the server will not accept any requests, however will finish handling already open requests.
	// When all open requests are served, the server will automatically change to StateCllosed.
	StateClosing
	//StateOpen is the state of the server when it is open; when in this state, the server will accept connections and handle them.
	StateOpen
)

type ClosedHandler func(net.Conn)
type StreamHandler func(*Server, Stream)

//ErrServerClosed is returned when the operation cannot be completed because the server is closed
var ErrServerClosed = errors.New("server closed")

//ErrServerClosing is returned when the operation cannot be completed because the server is closing
var ErrServerClosing = errors.New("server closing")

//ErrServerOpen is returned when the operation cannot be completed because the server is open
var ErrServerOpen = errors.New("server open")

//Server is an instance of an indipendently configured server.
type Server struct {
	Config jpvariable.Variables

	BufioSource BufioSource

	StreamSessionHandlers StreamSessionHandlerList
	SessionHandlers       ServerSessionHandlerList
	FrameHandlers         ServerFrameHandlerList
	DefaultStreamHandler  StreamHandler
	//ClosedHandler will be called on a connection that is accepted while the Server is closed.
	// It is expected to close the connection before it returns.
	//If it is nil, the connection is just closed.
	ClosedHandler ClosedHandler

	state       State
	stateLock   sync.RWMutex
	stateChange chan State

	connections     uint64
	connectionsLock sync.Mutex

	ErrorLog io.Writer
}

//registerConnection increments s.connections iff the s.state==StateOpen
//it returns true if the server was open or false if not.
func (s *Server) registerConnection() bool {
	s.stateLock.RLock()
	if s.state != StateOpen {
		s.stateLock.RUnlock()
		return false
	}
	//We have another connection!
	s.connectionsLock.Lock()
	s.connections++
	s.connectionsLock.Unlock()
	//Unlock the state afterwards, to make sure that it doesn't change to closed just before we inc connections
	s.stateLock.RUnlock()
	return true
}

//deregisterConnection decrements s.connections.
//If s.connections is then 0, it calls s.changeStateIf(StateClosing, StateClosed) to close the server if it is closing.
func (s *Server) deregisterConnection() {
	//Connection is handled, so dec connections, and close if yeah...
	s.connectionsLock.Lock()
	s.connections--
	s.connectionsLock.Unlock()
	if s.connections == 0 {
		s.changeStateIf(StateClosing, StateClosed)
	}
}

func (s *Server) ListenOn(l StreamSessionListener) error {
	for {
		sess, err := l.AcceptSession(s.Config, s.BufioSource)
		if err != nil {
			return err
		}
		err = s.HandleStreamSession(sess)
		if err != nil {
			return err
		}
	}
}

func (s *Server) ListenOnWithHandler(l StreamSessionListener, handler StreamSessionHandler) error {
	for {
		sess, err := l.AcceptSession(s.Config, s.BufioSource)
		if err != nil {
			return err
		}
		err = s.HandleStreamSessionWithHandler(sess, handler)
		if err != nil {
			return err
		}
	}
}

func (s *Server) handleStreamSessionFallback(sess StreamSession) error {
	for {
		stream, err := sess.AcceptStream()
		if err != nil {
			return err
		}
		/*stream = &WrappedStream{
			Reader: util.PrintyReader{
				Reader: stream,
				CC:     os.Stdout,
			},
			Writer: stream,
			Closer: stream,
			Stream: stream,
		}
		stream = &WrappedStream{
			Writer: util.WriterSpliter{
				stream, os.Stdout,
			},
			Reader: stream,
			Closer: stream,
			Stream: stream,
		}*/
		go s.HandleStream(stream, s.DefaultStreamHandler)
	}
}

func (s *Server) HandleStreamSession(sess StreamSession) error {
	if s.registerConnection() {
		handled, err := s.StreamSessionHandlers.Handle(sess)
		if !handled {
			s.handleStreamSessionFallback(sess)
		}
		s.deregisterConnection()
		return err
	}
	//TODO: Closed handler
	return nil
}

func (s *Server) HandleStreamSessionWithHandler(sess StreamSession, handler StreamSessionHandler) error {
	if s.registerConnection() {
		handled, err := s.StreamSessionHandlers.HandleWithHandler(sess, handler)
		if !handled {
			s.handleStreamSessionFallback(sess)
		}
		s.deregisterConnection()
		return err
	} else {
		//TODO: Closed handler
		return nil
	}
}

//HandleStream ... Basically, give the stream to this for the server to handle it!
func (s *Server) HandleStream(con Stream, handler StreamHandler) {
	//TODO: Deadlines

	s.stateLock.RLock()
	if s.state != StateOpen {
		s.stateLock.RUnlock()
		//If the server isn't open, either the ClosedHandler should be called, or the connection should be closed.
		if s.ClosedHandler == nil {
			con.Close()
		} else {
			s.ClosedHandler(nil) //TODO
		}
	} else {
		//We have another connection!
		s.connectionsLock.Lock()
		s.connections++
		s.connectionsLock.Unlock()
		//Unlock the state afterwards, to make sure that it doesn't change to closed just before we inc connections
		s.stateLock.RUnlock()

		handler(s, con)

		//Connection is handled, so dec connections, and close if yeah...
		s.connectionsLock.Lock()
		s.connections--
		s.connectionsLock.Unlock()
		if s.connections == 0 {
			s.changeStateIf(StateClosing, StateClosed)
		}
	}
}

//State returns the current State of the Server.
func (s *Server) State() State {
	return s.state
}

/*
//LockState locks the state of the server; it cannot be changed until UnlockState is called.
// Note that this also prevents the state from changing to StateClosed when all connections are closed. (It will chage when the state is unlocked)
func (s *Server) LockState() {
	s.stateLock.RLock()
}

//UnlockState unlocks the State of the server, allowing it to change.
func (s *Server) UnlockState() {
	s.stateLock.RUnlock()
}
*/

//WaitForState returns when the Server is in the given state
func (s *Server) WaitForState(state State) {
	s.stateLock.RLock()
	if s.state != state {
		s.stateLock.RUnlock()
		for {
			if state == <-s.stateChange {
				return
			}
		}
	}
	s.stateLock.RUnlock()
}

func (s *Server) changeState(state State) {
	s.stateLock.Lock()
	s.state = state
	s.stateLock.Unlock()
	select {
	case s.stateChange <- state:
	default:
	}
}

func (s *Server) changeStateIf(cond, to State) {
	//Check the condition as a reader (so we don't have to wait for too long first)
	s.stateLock.RLock()
	if s.state == cond {
		s.stateLock.RUnlock()
		//Now lock as a writer
		s.stateLock.Lock()
		//Recheck the condition in case it's changed between unlocking and locking again
		if s.state == cond {
			s.state = to
			select {
			case s.stateChange <- to:
			default:
			}
		}
		s.stateLock.Unlock()
	} else {
		s.stateLock.RUnlock()
	}
}

//Open opens the server
func (s *Server) Open() {
	s.changeState(StateOpen)
}

//Close closes the server; it won't accept anymore incoming requests.
// The server will only be closed when the there are no open requests.
// See WaitForState or CloseAndWait methods if you want to wait for the server to be closed.
func (s *Server) Close() {
	s.changeState(StateClosing)
}

//CloseAndWait closes the server, and then only returns once it is fully closed.
func (s *Server) CloseAndWait() {
	s.Close()
	s.WaitForState(StateClosed)
}

func (s *Server) AddFrameHandler(stage ServerFrameHandlerStage, priority uint16, handler *ServerFrameHandler) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()
	if s.state == StateClosed {
		s.FrameHandlers.insert(&ServerFrameHandlerListItem{
			handler:  handler,
			stage:    stage,
			priority: priority,
		})
		return nil
	} else if s.state == StateOpen {
		return ErrServerOpen
	} else {
		return ErrServerClosing
	}
}

func (s *Server) AddSessionHandler(stage ServerSessionHandlerStage, priority uint16, handler *ServerSessionHandler) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()
	if s.state == StateClosed {
		s.SessionHandlers.insert(&ServerSessionHandlerListItem{
			handler:  handler,
			stage:    stage,
			priority: priority,
		})
		return nil
	} else if s.state == StateOpen {
		return ErrServerOpen
	} else {
		return ErrServerClosing
	}
}

func (s *Server) AddStreamSessionHandler(stage StreamSessionHandlerStage, priority uint16, handler *StreamSessionHandler) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()
	if s.state == StateClosed {
		s.StreamSessionHandlers.insert(&StreamSessionHandlerListItem{
			handler:  handler,
			stage:    stage,
			priority: priority,
		})
		return nil
	} else if s.state == StateOpen {
		return ErrServerOpen
	} else {
		return ErrServerClosing
	}
}

var ParseTimes []time.Duration
var SetupTimes []time.Duration
var HandleTimes []time.Duration
var CheckTimes []time.Duration
var MimeTimes []time.Duration
var BodyTimes []time.Duration
var BodySetupTimes []time.Duration
var BodyWriteTimes []time.Duration
var BodyWriteSetupTimes []time.Duration
var BodyWriteSetupRawBodyTimes []time.Duration
var HBLengthTimes []time.Duration
var HBHeadWriteTimes []time.Duration
var BodyWriteSetupPipesTimes []time.Duration
var BodyWriteWriteTimes []time.Duration

func (s *Server) HandleSession(sess ServerSession) error {
	handled, err := s.SessionHandlers.Handle(sess)
	if handled || err != nil {
		return err
	}
	var frame *ServerFrame
	for {
		frame, err = sess.Frame()
		if err != nil {
			if err == EOS {
				return nil
			}
			return err
		}
		//TIMER:start := time.Now()
		err = s.HandleFrame(frame)
		//TIMER:HandleTimes = append(HandleTimes, time.Since(start))
		if err != nil {
			return err
		}
	}
}

func (s *Server) HandleFrame(frame *ServerFrame) (err error) {
	_, err = s.FrameHandlers.Handle(frame)
	return
}

func (s *Server) Install(ext Extension) error {
	return ext.ServerExtensionInit(s)
}
