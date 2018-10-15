package server

import (
	"bufio"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/http1/encoding"
)

//State is used to identify the state of the server.
type State byte

type ClosedHandler func(net.Conn)
type ProtocolHandler func(*Server, net.Conn)
type ProtocolNegotiator func(net.Conn, string) (string, error)

const (
	//StateClosed is the state of the server when it is closed - ie it will not accept incoming requests, and there are no open requests.
	StateClosed State = iota
	//StateClosing is the state of the server while it is closing. In this state, the server will not accept any requests, however will finish handling already open requests.
	// When all open requests are served, the server will automatically change to StateCllosed.
	StateClosing
	//StateOpen is the state of the server when it is open; when in this state, the server will accept connections and handle them.
	StateOpen
)

//ErrServerClosed is returned when the operation cannot be completed because the server is closed
var ErrServerClosed = errors.New("server closed")

//ErrServerClosing is returned when the operation cannot be completed because the server is closing
var ErrServerClosing = errors.New("server closing")

//ErrServerOpen is returned when the operation cannot be completed because the server is open
var ErrServerOpen = errors.New("server open")

//DefaultReaderBufSize will be used as Server.ReaderBufSize if it is 0
var DefaultReaderBufSize = 1024

//Server is an instance of an indipendently configured server.
type Server struct {
	state       State
	stateLock   sync.RWMutex
	stateChange chan State

	connections     uint64
	connectionsLock sync.Mutex

	ProtocolNegotiators []ProtocolNegotiator
	DefaultProtocol     string
	Protocols           map[string]ProtocolHandler

	Handlers HandlerList
	//ClosedHandler will be called on a connection that is accepted while the Server is closed.
	// It is expected to close the connection before it returns.
	//If it is nil, the connection is just closed.
	ClosedHandler ClosedHandler

	TransferEncodings encoding.List
	ContentEncodings  encoding.List
	ReaderBufSize     int
}

//NewBufioReader returns a new *bufio.Reader reading from the source r, using the ReaderBufSize property, or, if that is 0, DefaultReaderBufSize variable.
func (s *Server) NewBufioReader(r io.Reader) *bufio.Reader {
	size := s.ReaderBufSize
	if size == 0 {
		size = DefaultReaderBufSize
	}
	return bufio.NewReaderSize(r, size)
}

//HandleConnection ... Basically, give the connection to this for the server to handle it!
func (s *Server) HandleConnection(con net.Conn) {
	//TODO: Deadlines

	s.stateLock.RLock()
	if s.state != StateOpen {
		s.stateLock.RUnlock()
		//If the server isn't open, either the ClosedHandler should be called, or the connection should be closed.
		if s.ClosedHandler == nil {
			con.Close()
		} else {
			s.ClosedHandler(con)
		}
	} else {
		//We have another connection!
		s.connectionsLock.Lock()
		s.connections++
		s.connectionsLock.Unlock()
		//Unlock the state afterwards, to make sure that it doesn't change to closed just before we inc connections
		s.stateLock.RUnlock()

		//Determine the negotiated protocol
		var proto string
		var err error
		for i := range s.ProtocolNegotiators {
			proto, err = s.ProtocolNegotiators[i](con, proto)
			if err != nil {
				con.Close()
				goto decConnections
			}
		}
		if proto == "" {
			proto = s.DefaultProtocol
		}
		//And call the handler
		s.Protocols[proto](s, con)

	decConnections:
		//Connection is handled, so dec connections, and close if yeah...
		s.connectionsLock.Lock()
		s.connections--
		s.connectionsLock.Unlock()
		if s.connections == 0 {
			s.changeStateIf(StateClosing, StateClosed)
		}
	}
}

//ListenOn listens on l, with server s, and for every connection, Handler is called in a new goroutine.
func (s *Server) ListenOn(l net.Listener) error {
	var con net.Conn
	var err error
	var ne net.Error
	var ok, delayed bool
	//Start with a 2ms timeout if there is a temporary error
	retryDelay := 2 * time.Millisecond
	for {
		con, err = l.Accept()
		if err != nil {
			//If there is a temporary error, exponensially back off retrying...
			if ne, ok = err.(net.Error); ok && ne.Temporary() {
				delayed = true
				time.Sleep(retryDelay)
				if retryDelay < 128 {
					retryDelay *= 2
				}
				continue
			}
			return err
		} else if delayed {
			delayed = false
			retryDelay = 2 * time.Millisecond
		}
		go s.HandleConnection(con)
	}
}

//State returns the current State of the Server.
func (s *Server) State() State {
	return s.state
}

//LockState locks the state of the server; it cannot be changed until UnlockState is called.
// Note that this also prevents the state from changing to StateClosed when all connections are closed. (It will chage when the state is unlocked)
func (s *Server) LockState() {
	s.stateLock.RLock()
}

//UnlockState unlocks the State of the server, allowing it to change.
func (s *Server) UnlockState() {
	s.stateLock.RUnlock()
}

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
	s.stateChange <- state
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
			s.stateChange <- to
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

//Handle registers the given handler on the Server. The server must be closed for this to occur.
func (s *Server) Handle(stage HandlerStage, priority uint16, handler *Handler) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()
	if s.state == StateClosed {
		s.Handlers.insert(&HandlerListItem{
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
