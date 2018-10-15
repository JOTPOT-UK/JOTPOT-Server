package handler

import (
	"crypto/tls"
	"net"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/http1"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps/server"
)

//Handle takes a connection.
// If it is a TCP connection, it performs the handshake.
// After that, it calls the correct handler (ie http1.Handle, http2.Handler etc...)
func Handle(s *server.Server, con net.Conn) {
	defer con.Close()
	var err error
	if tlsCon, ok := con.(*tls.Conn); ok {
		if err = tlsCon.Handshake(); err != nil {
			return
		}
		tlsState := tlsCon.ConnectionState()
		if tlsState.NegotiatedProtocol != "" && tlsState.NegotiatedProtocol != "http/1.0" && tlsState.NegotiatedProtocol != "http/1.1" {
			if tlsState.NegotiatedProtocol == "h2" {
				//HandleHTTP2(con)
				panic("No HTTP/2 support!")
			}
			return
		}
	}
	http1.Handle(s, con)
}
