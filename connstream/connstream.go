package connstream

import (
	"net"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
	"github.com/JOTPOT-UK/JOTPOT-Server/tcp"
	"github.com/JOTPOT-UK/JOTPOT-Server/udp"
)

type ConnStream struct{ net.Conn }

func (_ *ConnStream) Via() ([]jps.Hop, bool) {
	return nil, false
}

func (_ *ConnStream) Session() jps.StreamSession {
	return nil
}

func (cs *ConnStream) Protocol() jps.Protocol {
	switch cs.Conn.(type) {
	case *net.TCPConn:
		return tcp.TCPProtocol{}
	case *net.UDPConn:
		return udp.UDPProtocol{}
	case *net.UnixConn:
		return UnixProtocol{}
	}
	return nil
}
