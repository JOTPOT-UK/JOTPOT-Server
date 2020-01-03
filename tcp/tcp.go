package tcp

import (
	"net"
	"sync"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps"

	"github.com/JOTPOT-UK/JOTPOT-Server/jpvariable"
)

type TCPListener struct {
	listener *net.TCPListener
}

func ListenOn(addr *net.TCPAddr) (TCPListener, error) {
	l, err := net.ListenTCP("tcp", addr)
	return TCPListener{l}, err
}

func (l *TCPListener) AcceptSession(vars jpvariable.Variables, bufiosource jps.BufioSource) (jps.StreamSession, error) {
	conn, err := l.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}
	sess := NewTCPSession(conn, vars, bufiosource)
	return &sess, nil
}

type TCPSession struct {
	stream      TCPStream
	accepted    bool
	lock        sync.Mutex
	vars        jpvariable.Variables
	bufiosource jps.BufioSource
}

func NewTCPSession(conn *net.TCPConn, vars jpvariable.Variables, bufiosource jps.BufioSource) TCPSession {
	return TCPSession{
		stream: TCPStream{
			TCPConn: conn,
		},
		bufiosource: bufiosource,
		vars:        vars,
	}
}

func (sess *TCPSession) BufioSource() jps.BufioSource {
	return sess.bufiosource
}

func (sess *TCPSession) Variables() jpvariable.Variables {
	return sess.vars
}

func (sess *TCPSession) AcceptStream() (jps.Stream, error) {
	sess.lock.Lock()
	if sess.accepted {
		sess.lock.Unlock()
		return nil, jps.EOS
	}
	sess.accepted = true
	sess.lock.Unlock()
	sess.stream.session = sess
	return &sess.stream, nil
}

func (sess *TCPSession) End() error {
	sess.lock.Lock()
	if sess.accepted {
		sess.lock.Unlock()
		return nil
	}
	sess.accepted = true
	sess.lock.Unlock()
	return sess.stream.Close()
}

func (sess *TCPSession) Close() error {
	sess.accepted = true
	return sess.stream.Close()
}

func (_ *TCPSession) RequestStream() (jps.Stream, error) {
	return nil, jps.ErrNotSupported
}

func (sess *TCPSession) LocalAddr() net.Addr {
	return sess.stream.TCPConn.LocalAddr()
}

func (sess *TCPSession) RemoteAddr() net.Addr {
	return sess.stream.TCPConn.RemoteAddr()
}

func (_ *TCPSession) Via() ([]jps.Hop, bool) {
	return nil, false
}

func (_ *TCPSession) Protocol() jps.Protocol {
	return TCPProtocol{}
}

type TCPStream struct {
	*net.TCPConn
	session *TCPSession
}

func (stream *TCPStream) Session() jps.StreamSession {
	return stream.session
}

func (_ *TCPStream) Via() ([]jps.Hop, bool) {
	return nil, false
}

func (_ *TCPStream) Protocol() jps.Protocol {
	return TCPProtocol{}
}

type TCPProtocol struct{}

func (_ TCPProtocol) Protocol() string {
	return "tcp"
}

func (_ TCPProtocol) String() string {
	return "tcp"
}

func (_ TCPProtocol) Version() interface{} {
	return nil
}
